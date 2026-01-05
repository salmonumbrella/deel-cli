package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/salmonumbrella/deel-cli/internal/config"
)

// escapePath escapes a path segment for safe use in URLs.
// This prevents path traversal attacks from malicious IDs.
func escapePath(segment string) string {
	return url.PathEscape(segment)
}

const (
	maxRetries           = 3
	baseBackoff          = 1 * time.Second
	maxBackoff           = 30 * time.Second
	circuitBreakerLimit  = 5
	circuitBreakerWindow = 30 * time.Second
)

// Client is the Deel API client
type Client struct {
	httpClient     *http.Client
	token          string
	baseURL        string
	debug          bool
	idempotencyKey string

	// Circuit breaker state
	mu               sync.Mutex
	consecutiveFails int
	circuitOpenedAt  time.Time
}

// NewClient creates a new Deel API client
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		token:   token,
		baseURL: config.BaseURL,
	}
}

// SetDebug enables or disables debug logging
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

// SetIdempotencyKey sets the idempotency key used for write requests.
func (c *Client) SetIdempotencyKey(key string) {
	c.idempotencyKey = key
}

// SetBaseURL sets the base URL for API requests.
// Note: For tests, prefer using testClient() from client_test.go
// which handles this automatically.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPatch, path, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	// Check circuit breaker
	if err := c.checkCircuitBreaker(); err != nil {
		return nil, err
	}

	url := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if attempt > 0 {
			backoff := c.calculateBackoff(attempt)
			if c.debug {
				slog.Info("retrying request", "attempt", attempt, "backoff", backoff)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := c.doRequest(ctx, method, url, body)
		if err != nil {
			lastErr = err
			continue
		}

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := c.parseRetryAfter(resp)
			if c.debug {
				slog.Info("rate limited", "retry_after", retryAfter)
			}
			if err := resp.Body.Close(); err != nil {
				slog.Debug("failed to close response body", "error", err)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryAfter):
			}
			lastErr = fmt.Errorf("rate limited")
			continue
		}

		// Handle server errors (5xx)
		if resp.StatusCode >= 500 {
			c.recordFailure()
			// Read response body to get error details
			errBody, _ := io.ReadAll(resp.Body)
			if err := resp.Body.Close(); err != nil {
				slog.Debug("failed to close response body", "error", err)
			}
			if c.debug && len(errBody) > 0 {
				slog.Info("server error response", "status", resp.StatusCode, "body", string(errBody))
			}
			lastErr = fmt.Errorf("server error: %d: %s", resp.StatusCode, string(errBody))
			continue
		}

		// Success or client error
		c.recordSuccess()

		respBody, err := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		if closeErr != nil {
			slog.Debug("failed to close response body", "error", closeErr)
		}

		if resp.StatusCode >= 400 {
			if c.debug {
				slog.Info("api error response", "status", resp.StatusCode, "body", string(respBody))
			}
			return nil, c.parseError(resp.StatusCode, respBody)
		}

		if c.debug {
			slog.Info("api response", "status", resp.StatusCode, "content_length", len(respBody))
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) doRequest(ctx context.Context, method, url string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.idempotencyKey != "" && method != http.MethodGet {
		req.Header.Set("Idempotency-Key", c.idempotencyKey)
	}

	if c.debug {
		slog.Info("api request", "method", method, "url", url)
		if body != nil {
			bodyBytes, _ := json.Marshal(body)
			slog.Info("request body", "body", string(bodyBytes))
		}
	}

	return c.httpClient.Do(req)
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := baseBackoff * time.Duration(math.Pow(2, float64(attempt-1)))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	// Add jitter (0-25% of backoff)
	jitter := time.Duration(rand.Float64() * 0.25 * float64(backoff))
	return backoff + jitter
}

func (c *Client) parseRetryAfter(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return baseBackoff
	}

	// Try parsing as seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date
	if t, err := http.ParseTime(retryAfter); err == nil {
		return time.Until(t)
	}

	return baseBackoff
}

func (c *Client) parseError(statusCode int, body []byte) error {
	// Try simple error format first: {"error": "..."} or {"message": "..."}
	var simpleErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &simpleErr); err == nil {
		if simpleErr.Error != "" {
			return &APIError{StatusCode: statusCode, Message: simpleErr.Error}
		}
		if simpleErr.Message != "" {
			return &APIError{StatusCode: statusCode, Message: simpleErr.Message}
		}
	}

	// Try nested Deel error format: {"errors": {"errors": [{"message": "..."}]}}
	var nestedErr struct {
		Errors struct {
			Errors []struct {
				Key     string `json:"key"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &nestedErr); err == nil && len(nestedErr.Errors.Errors) > 0 {
		// Collect all error messages
		var messages []string
		for _, e := range nestedErr.Errors.Errors {
			if e.Message != "" {
				messages = append(messages, e.Message)
			}
		}
		if len(messages) == 1 {
			return &APIError{StatusCode: statusCode, Message: messages[0]}
		}
		if len(messages) > 1 {
			return &APIError{StatusCode: statusCode, Message: fmt.Sprintf("%d errors: %v", len(messages), messages)}
		}
	}

	return &APIError{StatusCode: statusCode, Message: string(body)}
}

func (c *Client) checkCircuitBreaker() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.consecutiveFails >= circuitBreakerLimit {
		if time.Since(c.circuitOpenedAt) < circuitBreakerWindow {
			return fmt.Errorf("circuit breaker open: too many consecutive failures")
		}
		// Reset circuit breaker
		c.consecutiveFails = 0
	}
	return nil
}

func (c *Client) recordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails++
	if c.consecutiveFails >= circuitBreakerLimit {
		c.circuitOpenedAt = time.Now()
	}
}

func (c *Client) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.consecutiveFails = 0
}

// doMultipart performs an HTTP request with multipart/form-data body,
// using the same retry logic, circuit breaker, and error handling as do().
func (c *Client) doMultipart(ctx context.Context, method, path string, body io.Reader, contentType string) (json.RawMessage, error) {
	// Check circuit breaker
	if err := c.checkCircuitBreaker(); err != nil {
		return nil, err
	}

	url := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if attempt > 0 {
			backoff := c.calculateBackoff(attempt)
			if c.debug {
				slog.Info("retrying request", "attempt", attempt, "backoff", backoff)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		// For retries, we need to be able to re-read the body.
		// The caller should pass a *bytes.Reader or similar seekable reader.
		if seeker, ok := body.(io.Seeker); ok && attempt > 0 {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return nil, fmt.Errorf("failed to reset body for retry: %w", err)
			}
		}

		resp, err := c.doMultipartRequest(ctx, method, url, body, contentType)
		if err != nil {
			lastErr = err
			continue
		}

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := c.parseRetryAfter(resp)
			if c.debug {
				slog.Info("rate limited", "retry_after", retryAfter)
			}
			if err := resp.Body.Close(); err != nil {
				slog.Debug("failed to close response body", "error", err)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryAfter):
			}
			lastErr = fmt.Errorf("rate limited")
			continue
		}

		// Handle server errors (5xx)
		if resp.StatusCode >= 500 {
			c.recordFailure()
			// Read response body to get error details
			errBody, _ := io.ReadAll(resp.Body)
			if err := resp.Body.Close(); err != nil {
				slog.Debug("failed to close response body", "error", err)
			}
			if c.debug && len(errBody) > 0 {
				slog.Info("server error response", "status", resp.StatusCode, "body", string(errBody))
			}
			lastErr = fmt.Errorf("server error: %d: %s", resp.StatusCode, string(errBody))
			continue
		}

		// Success or client error
		c.recordSuccess()

		respBody, err := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		if closeErr != nil {
			slog.Debug("failed to close response body", "error", closeErr)
		}

		if resp.StatusCode >= 400 {
			if c.debug {
				slog.Info("api error response", "status", resp.StatusCode, "body", string(respBody))
			}
			return nil, c.parseError(resp.StatusCode, respBody)
		}

		if c.debug {
			slog.Info("api response", "status", resp.StatusCode, "content_length", len(respBody))
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doMultipartRequest creates and executes a single multipart HTTP request.
func (c *Client) doMultipartRequest(ctx context.Context, method, url string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	if c.idempotencyKey != "" && method != http.MethodGet {
		req.Header.Set("Idempotency-Key", c.idempotencyKey)
	}

	if c.debug {
		slog.Info("api request", "method", method, "url", url, "content_type", contentType)
	}

	return c.httpClient.Do(req)
}

// APIError represents an API error response.
//
//revive:disable-next-line:exported
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// APIStatusCode returns the HTTP status code (implements climerrors.StatusCoder)
func (e *APIError) APIStatusCode() int {
	return e.StatusCode
}

// APIMessage returns the raw error message (implements climerrors.Messager)
func (e *APIError) APIMessage() string {
	return e.Message
}

// FlexFloat64 handles JSON number fields that may be strings or numbers
type FlexFloat64 float64

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexFloat64) UnmarshalJSON(data []byte) error {
	// Try as number first
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexFloat64(num)
		return nil
	}

	// Try as string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			*f = 0
			return nil
		}
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as float64: %w", str, err)
		}
		*f = FlexFloat64(parsed)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexFloat64", string(data))
}
