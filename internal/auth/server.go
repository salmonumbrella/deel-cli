package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/secrets"
)

var validAccountName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// clientLimit tracks attempts for a specific client
type clientLimit struct {
	count   int
	resetAt time.Time
}

// rateLimiter tracks attempts per client IP and endpoint to prevent brute-force
type rateLimiter struct {
	mu          sync.Mutex
	attempts    map[string]*clientLimit // key: "clientIP:endpoint"
	maxAttempts int
	window      time.Duration
}

// newRateLimiter creates a rate limiter with the given max attempts and time window
func newRateLimiter(maxAttempts int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts:    make(map[string]*clientLimit),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

// check verifies if the client has exceeded the rate limit for this endpoint
func (rl *rateLimiter) check(clientIP, endpoint string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := clientIP + ":" + endpoint
	now := time.Now()

	// Reset if window expired
	if limit, exists := rl.attempts[key]; exists && now.After(limit.resetAt) {
		delete(rl.attempts, key)
	}

	// Initialize new limit if doesn't exist
	if rl.attempts[key] == nil {
		rl.attempts[key] = &clientLimit{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return nil
	}

	// Increment and check limit
	rl.attempts[key].count++
	if rl.attempts[key].count > rl.maxAttempts {
		return fmt.Errorf("too many attempts, please try again later")
	}
	return nil
}

// startCleanup starts a background goroutine that periodically removes expired entries
func (rl *rateLimiter) startCleanup(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rl.cleanup()
			case <-stop:
				return
			}
		}
	}()
}

// cleanup removes all expired entries from the rate limiter
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, limit := range rl.attempts {
		if now.After(limit.resetAt) {
			delete(rl.attempts, key)
		}
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If no port separator, assume the whole string is the IP
		return r.RemoteAddr
	}
	return host
}

// ValidateAccountName validates an account name
func ValidateAccountName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("account name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("account name too long (max 64 characters)")
	}
	if !validAccountName.MatchString(name) {
		return fmt.Errorf("account name contains invalid characters (use only letters, numbers, dash, underscore)")
	}
	return nil
}

// ValidateToken validates a Deel PAT token
func ValidateToken(token string) error {
	if len(token) == 0 {
		return fmt.Errorf("token cannot be empty")
	}
	if len(token) > 4096 {
		return fmt.Errorf("token too long (max 4096 characters)")
	}
	return nil
}

// SetupResult contains the result of a browser-based setup
type SetupResult struct {
	AccountName string
	Error       error
}

// SetupServer handles the browser-based authentication flow
type SetupServer struct {
	result        chan SetupResult
	shutdown      chan struct{}
	stopCleanup   chan struct{}
	pendingResult *SetupResult
	pendingMu     sync.Mutex
	csrfToken     string
	store         secrets.Store
	limiter       *rateLimiter
}

// NewSetupServer creates a new setup server
func NewSetupServer(store secrets.Store) (*SetupServer, error) {
	// Generate CSRF token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	stopCleanup := make(chan struct{})
	limiter := newRateLimiter(10, 15*time.Minute)
	limiter.startCleanup(5*time.Minute, stopCleanup)

	return &SetupServer{
		result:      make(chan SetupResult, 1),
		shutdown:    make(chan struct{}),
		stopCleanup: stopCleanup,
		csrfToken:   hex.EncodeToString(tokenBytes),
		store:       store,
		limiter:     limiter,
	}, nil
}

// Start starts the setup server and opens the browser
func (s *SetupServer) Start(ctx context.Context) (*SetupResult, error) {
	// Ensure cleanup goroutine is stopped when server exits
	defer close(s.stopCleanup)

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)
	mux.HandleFunc("/complete", s.handleComplete)
	mux.HandleFunc("/accounts", s.handleListAccounts)
	mux.HandleFunc("/remove-account", s.handleRemoveAccount)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Open browser
	go func() {
		if err := openBrowser(baseURL); err != nil {
			slog.Info("failed to open browser, user can navigate manually", "url", baseURL)
		}
	}()

	// Wait for result or context cancellation
	select {
	case result := <-s.result:
		_ = server.Shutdown(context.Background())
		return &result, nil
	case err := <-serverErr:
		return nil, fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return nil, ctx.Err()
	case <-s.shutdown:
		_ = server.Shutdown(context.Background())
		s.pendingMu.Lock()
		defer s.pendingMu.Unlock()
		if s.pendingResult != nil {
			return s.pendingResult, nil
		}
		return nil, fmt.Errorf("setup cancelled")
	}
}

// handleSetup serves the main setup page
func (s *SetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"CSRFToken": s.csrfToken,
	}

	// Set security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	if err := tmpl.Execute(w, data); err != nil {
		slog.Error("setup template execution failed", "error", err)
	}
}

// validateCredentials tests credentials against the Deel API
func (s *SetupServer) validateCredentials(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("Token is required")
	}

	client := api.NewClient(token)
	// Use /rest/v2/contracts with limit=1 as a lightweight validation endpoint
	// The /rest/v2/profiles/me endpoint doesn't exist in Deel's API
	_, err := client.Get(ctx, "/rest/v2/contracts?limit=1")
	if err != nil {
		// Check if it's an API error and provide cleaner messages
		if apiErr, ok := err.(*api.APIError); ok {
			switch apiErr.StatusCode {
			case 401:
				return fmt.Errorf("Invalid or expired token. Please check your Personal Access Token")
			case 403:
				return fmt.Errorf("Access denied. Your token may not have the required permissions")
			case 429:
				return fmt.Errorf("Rate limited. Please wait a moment and try again")
			default:
				return fmt.Errorf("API error (%d): %s", apiErr.StatusCode, apiErr.Message)
			}
		}
		return fmt.Errorf("Connection failed: %v", err)
	}

	return nil
}

// handleValidate tests credentials without saving
func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token FIRST (before rate limiting)
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Check rate limit per client IP
	clientIP := getClientIP(r)
	if err := s.limiter.check(clientIP, "/validate"); err != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		AccountName string `json:"account_name"`
		Token       string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize inputs - strip all whitespace and control characters
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.Token = strings.TrimSpace(req.Token)
	// Remove any invisible characters that might have been pasted
	req.Token = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1 // Remove control characters
		}
		return r
	}, req.Token)

	// Validate input format
	if err := ValidateAccountName(req.AccountName); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if err := ValidateToken(req.Token); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate credentials
	if err := s.validateCredentials(r.Context(), req.Token); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Connection successful!",
	})
}

// handleSubmit saves credentials after validation
func (s *SetupServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token FIRST (before rate limiting)
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Check rate limit per client IP
	clientIP := getClientIP(r)
	if err := s.limiter.check(clientIP, "/submit"); err != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		AccountName string `json:"account_name"`
		Token       string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize inputs - strip all whitespace and control characters
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.Token = strings.TrimSpace(req.Token)
	// Remove any invisible characters that might have been pasted
	req.Token = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1 // Remove control characters
		}
		return r
	}, req.Token)

	// Validate input format
	if err := ValidateAccountName(req.AccountName); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if err := ValidateToken(req.Token); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate credentials
	if err := s.validateCredentials(r.Context(), req.Token); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Save to keychain
	err := s.store.Set(req.AccountName, secrets.Credentials{
		Token: req.Token,
	})
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   "Failed to save credentials to secure storage",
		})
		return
	}

	// Store pending result
	s.pendingMu.Lock()
	s.pendingResult = &SetupResult{
		AccountName: req.AccountName,
	}
	s.pendingMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success":      true,
		"account_name": req.AccountName,
	})
}

// handleSuccess serves the success page
func (s *SetupServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Use server state instead of URL parameter to prevent spoofing
	s.pendingMu.Lock()
	accountName := ""
	if s.pendingResult != nil {
		accountName = s.pendingResult.AccountName
	}
	s.pendingMu.Unlock()

	data := map[string]string{
		"AccountName": accountName,
		"CSRFToken":   s.csrfToken,
	}

	// Set security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	if err := tmpl.Execute(w, data); err != nil {
		slog.Error("success template execution failed", "error", err)
	}
}

// handleComplete signals that setup is done
func (s *SetupServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	s.pendingMu.Lock()
	if s.pendingResult != nil {
		s.result <- *s.pendingResult
	}
	s.pendingMu.Unlock()
	close(s.shutdown)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleListAccounts returns all stored accounts
func (s *SetupServer) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	creds, err := s.store.List()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"accounts": []any{},
		})
		return
	}

	accounts := make([]map[string]any, 0, len(creds))
	for _, c := range creds {
		accounts = append(accounts, map[string]any{
			"name":      c.Name,
			"createdAt": c.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accounts": accounts,
	})
}

// handleRemoveAccount removes an account from the store
func (s *SetupServer) handleRemoveAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Check rate limit per client IP
	clientIP := getClientIP(r)
	if err := s.limiter.check(clientIP, "/remove-account"); err != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if err := s.store.Delete(req.Name); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to remove account: %v", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("JSON encoding failed", "error", err)
	}
}

// openBrowser opens the URL in the default browser.
// For security, only localhost URLs are allowed.
func openBrowser(url string) error {
	// Validate URL is localhost only to prevent command injection
	if !strings.HasPrefix(url, "http://127.0.0.1:") &&
		!strings.HasPrefix(url, "http://localhost:") {
		return fmt.Errorf("invalid URL for browser: must be localhost")
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// HTML Templates

var setupTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Deel CLI Setup</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700&display=swap" rel="stylesheet">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }

        :root {
            --deel-black: #1B1B1B;
            --deel-purple: #B794F6;
            --deel-purple-light: #D4B9FC;
            --deel-purple-dark: #9F7AEA;
            --deel-blue: #0EA5E9;
            --deel-gray: #6B7280;
            --deel-light-gray: #E5E7EB;
            --deel-white: #FFFFFF;
            --deel-success: #22C55E;
            --deel-error: #EF4444;
        }

        body {
            font-family: 'DM Sans', -apple-system, BlinkMacSystemFont, sans-serif;
            background: linear-gradient(135deg, #C4B5FD 0%, #A78BFA 50%, #8B5CF6 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 24px;
            position: relative;
            overflow: hidden;
        }

        body::before {
            content: '';
            position: absolute;
            top: -50%;
            right: -20%;
            width: 80%;
            height: 150%;
            background: linear-gradient(135deg, rgba(167, 139, 250, 0.6) 0%, rgba(139, 92, 246, 0.4) 100%);
            border-radius: 50%;
            z-index: 0;
        }

        body::after {
            content: '';
            position: absolute;
            bottom: -30%;
            left: -10%;
            width: 50%;
            height: 80%;
            background: linear-gradient(45deg, rgba(196, 181, 253, 0.5) 0%, rgba(167, 139, 250, 0.3) 100%);
            border-radius: 50%;
            z-index: 0;
        }

        .hidden { display: none !important; }

        .container {
            background: var(--deel-white);
            border-radius: 20px;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.15);
            padding: 48px 40px 40px;
            max-width: 420px;
            width: 100%;
            position: relative;
            z-index: 1;
        }

        .logo {
            text-align: center;
            margin-bottom: 16px;
        }

        .logo svg {
            height: 100px;
            width: auto;
        }

        .subtitle {
            text-align: center;
            color: var(--deel-gray);
            font-size: 15px;
            margin-bottom: 28px;
        }

        /* Accounts section */
        .accounts-section {
            margin-bottom: 20px;
        }

        .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 12px;
        }

        .section-title {
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--deel-gray);
        }

        .account-count {
            font-size: 11px;
            color: var(--deel-gray);
            background: #F3F4F6;
            padding: 3px 10px;
            border-radius: 999px;
        }

        .accounts-list {
            display: flex;
            flex-direction: column;
            gap: 8px;
            margin-bottom: 12px;
        }

        .account-card {
            background: #F9FAFB;
            border: 1px solid var(--deel-light-gray);
            border-radius: 10px;
            padding: 12px 14px;
            display: flex;
            align-items: center;
            gap: 12px;
            transition: all 0.2s ease;
        }

        .account-card:hover {
            border-color: #D1D5DB;
            background: #F3F4F6;
        }

        .account-avatar {
            width: 36px;
            height: 36px;
            background: linear-gradient(135deg, #1c4396, #2c71f0);
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: 700;
            font-size: 14px;
            color: white;
            flex-shrink: 0;
        }

        .account-info {
            flex: 1;
            min-width: 0;
        }

        .account-name {
            font-size: 14px;
            font-weight: 600;
            color: var(--deel-black);
        }

        .account-date {
            font-size: 11px;
            color: var(--deel-gray);
        }

        .remove-btn {
            width: 28px;
            height: 28px;
            background: transparent;
            border: none;
            border-radius: 6px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            color: var(--deel-gray);
            transition: all 0.2s ease;
            opacity: 0;
            flex-shrink: 0;
        }

        .account-card:hover .remove-btn {
            opacity: 1;
        }

        .remove-btn:hover {
            background: #FEF2F2;
            color: var(--deel-error);
        }

        .remove-btn svg {
            width: 14px;
            height: 14px;
        }

        .add-account-btn {
            width: 100%;
            background: transparent;
            border: 1px dashed #D1D5DB;
            border-radius: 10px;
            padding: 14px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
            color: var(--deel-gray);
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .add-account-btn:hover {
            border-color: var(--deel-black);
            color: var(--deel-black);
            background: #F9FAFB;
        }

        .add-account-btn svg {
            width: 16px;
            height: 16px;
        }

        /* Empty state */
        .empty-state {
            text-align: center;
            padding: 24px 16px;
            background: #F9FAFB;
            border: 1px solid var(--deel-light-gray);
            border-radius: 10px;
            margin-bottom: 20px;
        }

        .empty-state h3 {
            font-size: 15px;
            font-weight: 600;
            margin-bottom: 4px;
            color: var(--deel-black);
        }

        .empty-state p {
            font-size: 13px;
            color: var(--deel-gray);
        }

        /* Setup card */
        .setup-card {
            background: #F9FAFB;
            border: 1px solid var(--deel-light-gray);
            border-radius: 12px;
            margin-bottom: 20px;
            overflow: hidden;
        }

        .setup-header {
            padding: 14px 16px;
            border-bottom: 1px solid var(--deel-light-gray);
            display: flex;
            align-items: center;
            justify-content: space-between;
            background: white;
        }

        .setup-header h2 {
            font-size: 14px;
            font-weight: 600;
        }

        .close-btn {
            width: 24px;
            height: 24px;
            background: #F3F4F6;
            border: none;
            border-radius: 6px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            color: var(--deel-gray);
            transition: all 0.2s ease;
        }

        .close-btn:hover {
            background: var(--deel-light-gray);
            color: var(--deel-black);
        }

        .close-btn svg {
            width: 12px;
            height: 12px;
        }

        .setup-body {
            padding: 16px;
        }

        .form-group {
            margin-bottom: 14px;
        }

        input {
            width: 100%;
            padding: 12px 14px;
            border: 1px solid var(--deel-light-gray);
            border-radius: 8px;
            font-size: 14px;
            font-family: inherit;
            color: var(--deel-black);
            background: var(--deel-white);
            transition: all 0.2s ease;
        }

        input::placeholder {
            color: #9CA3AF;
        }

        input:hover {
            border-color: #D1D5DB;
        }

        input:focus {
            outline: none;
            border-color: var(--deel-black);
        }

        input.error {
            border-color: var(--deel-error);
        }

        .hint {
            font-size: 11px;
            color: var(--deel-gray);
            margin-top: 6px;
        }

        .hint a {
            color: var(--deel-blue);
            text-decoration: none;
            font-weight: 500;
        }

        .hint a:hover {
            text-decoration: underline;
        }

        .buttons {
            display: flex;
            gap: 8px;
            margin-top: 16px;
        }

        button.btn {
            flex: 1;
            padding: 12px 16px;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 600;
            font-family: inherit;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .btn-primary {
            background: var(--deel-black);
            color: var(--deel-white);
            border: none;
        }

        .btn-primary:hover {
            background: #333;
        }

        .btn-primary:disabled {
            background: #D1D5DB;
            color: #9CA3AF;
            cursor: not-allowed;
        }

        .btn-secondary {
            background: var(--deel-white);
            color: var(--deel-black);
            border: 1.5px solid var(--deel-black);
        }

        .btn-secondary:hover {
            background: var(--deel-black);
            color: var(--deel-white);
        }

        .status {
            margin-top: 12px;
            padding: 10px 12px;
            border-radius: 8px;
            font-size: 12px;
            display: none;
            font-weight: 500;
        }

        .status.error {
            background: #FEF2F2;
            color: var(--deel-error);
            display: block;
        }

        .status.success {
            background: #F0FDF4;
            color: var(--deel-success);
            display: block;
        }

        .status.loading {
            background: #F0F9FF;
            color: var(--deel-blue);
            display: block;
        }

        .footer {
            text-align: center;
            margin-top: 24px;
            padding-top: 20px;
            border-top: 1px solid var(--deel-light-gray);
        }

        .footer a {
            color: var(--deel-gray);
            text-decoration: none;
            font-size: 13px;
            font-weight: 500;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            transition: color 0.2s ease;
        }

        .footer a:hover {
            color: var(--deel-black);
        }

        .footer svg {
            opacity: 0.6;
        }

        .footer a:hover svg {
            opacity: 1;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <svg viewBox="0 0 560 400" xmlns="http://www.w3.org/2000/svg">
                <g transform="matrix(5.25771 0 0 5.25771 100 136.799)">
                    <g fill="#1c4396" fill-rule="nonzero">
                        <path d="m56.133 23.533c-.101 0-.182-.081-.182-.182v-21.941c0-.086.06-.16.145-.178l3.832-.792c.113-.024.219.063.219.178v22.733c0 .101-.082.182-.182.182z"/>
                        <path d="m7.999 23.997c-1.53 0-2.896-.371-4.098-1.114s-2.153-1.76-2.852-3.049c-.699-1.29-1.049-2.754-1.049-4.393s.35-3.093 1.049-4.36c.699-1.29 1.65-2.295 2.852-3.016 1.202-.743 2.568-1.115 4.098-1.115 1.224 0 2.295.229 3.213.688.737.369 1.362.858 1.873 1.466.116.137.356.059.356-.12v-7.968c0-.086.061-.161.145-.178l3.832-.793c.113-.023.219.063.219.179v23.198c0 .1-.081.182-.182.182h-3.405c-.087 0-.162-.062-.179-.147l-.349-1.772c-.031-.156-.235-.202-.337-.08-.489.587-1.103 1.111-1.842 1.573-.852.546-1.967.819-3.344.819zm.885-3.671c1.355 0 2.459-.448 3.311-1.345.875-.917 1.312-2.087 1.312-3.507 0-1.421-.437-2.579-1.312-3.475-.852-.918-1.956-1.377-3.311-1.377-1.333 0-2.437.448-3.311 1.344s-1.311 2.054-1.311 3.475c0 1.42.437 2.59 1.311 3.508.874.917 1.978 1.377 3.311 1.377z"/>
                        <path d="m27.869 24c-1.639 0-3.092-.35-4.36-1.049-1.267-.7-2.262-1.683-2.983-2.951-.721-1.267-1.082-2.732-1.082-4.393 0-1.682.35-3.18 1.049-4.491.721-1.311 1.705-2.328 2.951-3.049 1.267-.743 2.753-1.114 4.458-1.114 1.596 0 3.005.349 4.229 1.049 1.224.699 2.175 1.661 2.852 2.885.7 1.202 1.049 2.546 1.049 4.032 0 .24-.011.492-.032.754 0 .21-.007.427-.022.651-.005.095-.085.169-.18.169h-11.995c-.106 0-.19.089-.179.194.117 1.175.549 2.105 1.295 2.789.809.721 1.781 1.082 2.918 1.082.852 0 1.562-.186 2.131-.558.556-.37.976-.838 1.26-1.403.032-.063.096-.105.167-.105h3.901c.122 0 .21.118.171.233-.312.937-.801 1.799-1.467 2.587-.7.83-1.574 1.486-2.623 1.967-1.027.48-2.196.721-3.508.721zm.033-13.638c-1.027 0-1.934.295-2.721.885-.738.534-1.226 1.336-1.464 2.408-.025.112.061.215.175.215h7.685c.105 0 .189-.089.178-.193-.099-.981-.487-1.769-1.165-2.364-.721-.634-1.617-.951-2.688-.951z"/>
                        <path d="m46.124 24c-1.639 0-3.093-.35-4.36-1.049-1.268-.7-2.262-1.683-2.983-2.951-.722-1.267-1.082-2.732-1.082-4.393 0-1.682.349-3.18 1.049-4.491.721-1.311 1.705-2.328 2.95-3.049 1.268-.743 2.754-1.114 4.459-1.114 1.595 0 3.005.349 4.229 1.049 1.224.699 2.174 1.661 2.852 2.885.699 1.202 1.049 2.546 1.049 4.032 0 .24-.011.492-.033.754 0 .21-.007.427-.021.651-.006.095-.085.169-.18.169h-11.996c-.105 0-.189.089-.178.194.117 1.175.548 2.105 1.294 2.789.809.721 1.782 1.082 2.918 1.082.853 0 1.563-.186 2.131-.558.556-.37.976-.838 1.261-1.403.032-.063.096-.105.167-.105h3.901c.122 0 .209.118.171.233-.312.937-.801 1.799-1.468 2.587-.699.83-1.573 1.486-2.622 1.967-1.027.48-2.197.721-3.508.721zm.033-13.638c-1.027 0-1.934.295-2.721.885-.739.534-1.227 1.336-1.465 2.408-.024.112.062.215.176.215h7.685c.105 0 .188-.089.178-.193-.099-.981-.488-1.769-1.165-2.364-.721-.634-1.617-.951-2.688-.951z"/>
                    </g>
                    <circle cx="65.646" cy="20.713" fill="#2c71f0" r="2.825"/>
                </g>
            </svg>
        </div>
        <p class="subtitle">Manage your accounts</p>

        <input type="hidden" id="csrfToken" value="{{.CSRFToken}}">

        <!-- Accounts Section -->
        <div id="accountsSection" class="accounts-section hidden">
            <div class="section-header">
                <span class="section-title">Connected Accounts</span>
                <span id="accountCount" class="account-count">0 accounts</span>
            </div>
            <div id="accountsList" class="accounts-list"></div>
            <button id="addAccountBtn" class="add-account-btn">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                Add Account
            </button>
        </div>

        <!-- Empty State - visible by default -->
        <div id="emptyState" class="empty-state">
            <h3>No accounts connected</h3>
            <p>Add your first Deel account to get started</p>
        </div>

        <!-- Setup Card - visible by default -->
        <div id="setupCard" class="setup-card">
            <div class="setup-header">
                <h2>Add Deel Account</h2>
                <button id="closeSetupBtn" class="close-btn">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
            </div>
            <div class="setup-body">
                <form id="setupForm">
                    <div class="form-group">
                        <input type="text" id="accountName" placeholder="Account name (e.g., production)" required>
                        <p class="hint">Local label only - pick any name to identify this account</p>
                    </div>
                    <div class="form-group">
                        <input type="password" id="token" placeholder="Deel API Token" required>
                        <p class="hint">Create a token in the <a href="https://app.deel.com/developer-center" target="_blank">Deel Developer Center</a></p>
                    </div>
                    <div id="status" class="status"></div>
                    <div class="buttons">
                        <button type="button" class="btn btn-secondary" id="testBtn">Test</button>
                        <button type="submit" class="btn btn-primary" id="saveBtn">Save Account</button>
                    </div>
                </form>
            </div>
        </div>

        <div class="footer">
            <a href="https://github.com/salmonumbrella/deel-cli" target="_blank">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                View on GitHub
            </a>
        </div>
    </div>
    <script>
        const csrfToken = document.getElementById('csrfToken').value;
        const accountsSection = document.getElementById('accountsSection');
        const accountsList = document.getElementById('accountsList');
        const accountCount = document.getElementById('accountCount');
        const emptyState = document.getElementById('emptyState');
        const setupCard = document.getElementById('setupCard');
        const addAccountBtn = document.getElementById('addAccountBtn');
        const closeSetupBtn = document.getElementById('closeSetupBtn');
        const form = document.getElementById('setupForm');
        const accountNameInput = document.getElementById('accountName');
        const tokenInput = document.getElementById('token');
        const testBtn = document.getElementById('testBtn');
        const saveBtn = document.getElementById('saveBtn');
        const statusEl = document.getElementById('status');

        let accounts = [];

        // Escape HTML special characters to prevent XSS
        function escapeHtml(str) {
            const div = document.createElement('div');
            div.textContent = str;
            return div.innerHTML;
        }

        // Load accounts on page load
        async function loadAccounts() {
            try {
                const response = await fetch('/accounts');
                const data = await response.json();
                accounts = data.accounts || [];
                renderAccounts();
            } catch (err) {
                accounts = [];
                renderAccounts();
            }
        }

        function renderAccounts() {
            accountCount.textContent = accounts.length + ' account' + (accounts.length !== 1 ? 's' : '');

            if (accounts.length === 0) {
                accountsSection.classList.add('hidden');
                emptyState.classList.remove('hidden');
                setupCard.classList.remove('hidden');
                closeSetupBtn.classList.add('hidden');
            } else {
                accountsSection.classList.remove('hidden');
                emptyState.classList.add('hidden');
                setupCard.classList.add('hidden');
                closeSetupBtn.classList.remove('hidden');

                accountsList.innerHTML = accounts.map(acc => {
                    const safeName = escapeHtml(acc.name);
                    const initial = escapeHtml(acc.name.charAt(0).toUpperCase());
                    const date = new Date(acc.createdAt);
                    const dateStr = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
                    return '<div class="account-card" data-name="' + safeName + '">' +
                        '<div class="account-avatar">' + initial + '</div>' +
                        '<div class="account-info">' +
                        '<div class="account-name">' + safeName + '</div>' +
                        '<div class="account-date">Added ' + dateStr + '</div>' +
                        '</div>' +
                        '<button class="remove-btn" onclick="removeAccount(' + JSON.stringify(acc.name) + ')" title="Remove account">' +
                        '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>' +
                        '</button>' +
                        '</div>';
                }).join('');
            }
        }

        async function removeAccount(name) {
            if (!confirm('Remove account "' + name + '" from Deel CLI?')) return;
            try {
                const response = await fetch('/remove-account', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify({ name })
                });
                const data = await response.json();
                if (data.success) await loadAccounts();
            } catch (err) {
                console.error('Failed to remove account:', err);
            }
        }

        // Show/hide setup form
        addAccountBtn.addEventListener('click', () => {
            setupCard.classList.remove('hidden');
            accountNameInput.focus();
        });

        closeSetupBtn.addEventListener('click', () => {
            if (accounts.length > 0) {
                setupCard.classList.add('hidden');
                form.reset();
                hideStatus();
            }
        });

        // Clear error state on input
        [accountNameInput, tokenInput].forEach(input => {
            input.addEventListener('input', () => input.classList.remove('error'));
        });

        function showStatus(type, message) {
            statusEl.className = 'status ' + type;
            statusEl.textContent = message;
        }

        function hideStatus() {
            statusEl.className = 'status';
        }

        function validateFields() {
            let valid = true;
            if (!accountNameInput.value.trim()) {
                accountNameInput.classList.add('error');
                valid = false;
            }
            if (!tokenInput.value.trim()) {
                tokenInput.classList.add('error');
                valid = false;
            }
            return valid;
        }

        // Test connection
        testBtn.addEventListener('click', async () => {
            if (!validateFields()) return;

            showStatus('loading', 'Testing connection...');
            testBtn.disabled = true;

            try {
                const resp = await fetch('/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify({ account_name: accountNameInput.value.trim(), token: tokenInput.value.trim() })
                });
                const data = await resp.json();
                if (data.success) {
                    showStatus('success', data.message);
                } else {
                    showStatus('error', data.error);
                }
            } catch (e) {
                showStatus('error', 'Connection failed: ' + e.message);
            } finally {
                testBtn.disabled = false;
            }
        });

        // Submit form
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (!validateFields()) return;

            saveBtn.disabled = true;
            showStatus('loading', 'Saving account...');

            try {
                const resp = await fetch('/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify({ account_name: accountNameInput.value.trim(), token: tokenInput.value.trim() })
                });
                const data = await resp.json();
                if (data.success) {
                    window.location.href = '/success';
                } else {
                    showStatus('error', data.error);
                    saveBtn.disabled = false;
                }
            } catch (e) {
                showStatus('error', 'Save failed: ' + e.message);
                saveBtn.disabled = false;
            }
        });

        // Initialize
        loadAccounts();
    </script>
</body>
</html>`

var successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Deel CLI - Success</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }

        :root {
            --deel-navy: #1c4396;
            --deel-blue: #2c71f0;
            --deel-black: #1B1B1B;
            --deel-gray: #6B7280;
            --deel-light-gray: #E5E7EB;
            --deel-white: #FFFFFF;
        }

        body {
            font-family: 'DM Sans', -apple-system, BlinkMacSystemFont, sans-serif;
            background: linear-gradient(145deg, #1c4396 0%, #2c71f0 50%, #4a90f7 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 24px;
            position: relative;
            overflow: hidden;
        }

        body::before {
            content: '';
            position: absolute;
            top: -30%;
            right: -15%;
            width: 60%;
            height: 120%;
            background: radial-gradient(ellipse, rgba(44, 113, 240, 0.4) 0%, transparent 70%);
            z-index: 0;
        }

        body::after {
            content: '';
            position: absolute;
            bottom: -20%;
            left: -10%;
            width: 50%;
            height: 70%;
            background: radial-gradient(ellipse, rgba(28, 67, 150, 0.5) 0%, transparent 70%);
            z-index: 0;
        }

        .container {
            background: var(--deel-white);
            border-radius: 20px;
            box-shadow: 0 30px 60px -15px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255,255,255,0.1);
            padding: 44px 40px 36px;
            max-width: 440px;
            width: 100%;
            text-align: center;
            position: relative;
            z-index: 1;
        }

        .success-icon {
            width: 72px;
            height: 72px;
            background: linear-gradient(135deg, var(--deel-navy) 0%, var(--deel-blue) 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 28px;
            animation: scale-in 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
            box-shadow: 0 8px 24px -4px rgba(44, 113, 240, 0.4);
        }

        @keyframes scale-in {
            0% { transform: scale(0) rotate(-45deg); opacity: 0; }
            100% { transform: scale(1) rotate(0deg); opacity: 1; }
        }

        .success-icon svg {
            width: 32px;
            height: 32px;
            stroke: var(--deel-white);
            stroke-width: 3;
            animation: check-draw 0.6s ease-out 0.3s both;
            filter: drop-shadow(0 2px 4px rgba(0,0,0,0.1));
        }

        @keyframes check-draw {
            0% { stroke-dashoffset: 30; opacity: 0; }
            100% { stroke-dashoffset: 0; opacity: 1; }
        }

        .success-icon svg path {
            stroke-dasharray: 30;
        }

        h1 {
            font-size: 24px;
            font-weight: 700;
            color: var(--deel-black);
            margin-bottom: 6px;
            letter-spacing: -0.03em;
        }

        .subtitle {
            color: var(--deel-gray);
            margin-bottom: 28px;
            font-size: 14px;
        }

        /* Terminal Window */
        .terminal {
            background: #0f172a;
            border-radius: 12px;
            overflow: hidden;
            margin-bottom: 24px;
            box-shadow: 0 4px 20px -4px rgba(15, 23, 42, 0.3), inset 0 1px 0 rgba(255,255,255,0.05);
            animation: terminal-appear 0.6s ease-out 0.4s both;
        }

        @keyframes terminal-appear {
            0% { opacity: 0; transform: translateY(10px); }
            100% { opacity: 1; transform: translateY(0); }
        }

        .terminal-header {
            background: linear-gradient(180deg, #1e293b 0%, #172033 100%);
            padding: 10px 14px;
            display: flex;
            align-items: center;
            gap: 8px;
            border-bottom: 1px solid rgba(255,255,255,0.05);
        }

        .terminal-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
        }

        .terminal-dot.red { background: #ef4444; }
        .terminal-dot.yellow { background: #eab308; }
        .terminal-dot.green { background: #22c55e; }

        .terminal-title {
            flex: 1;
            text-align: center;
            font-family: 'JetBrains Mono', monospace;
            font-size: 11px;
            color: #64748b;
            letter-spacing: 0.02em;
        }

        .terminal-body {
            padding: 16px 18px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 13px;
            line-height: 1.7;
            text-align: left;
        }

        .terminal-line {
            display: flex;
            align-items: center;
            animation: line-appear 0.4s ease-out both;
        }

        .terminal-line:nth-child(1) { animation-delay: 0.6s; }
        .terminal-line:nth-child(2) { animation-delay: 0.9s; }
        .terminal-line:nth-child(3) { animation-delay: 1.2s; }

        @keyframes line-appear {
            0% { opacity: 0; transform: translateX(-8px); }
            100% { opacity: 1; transform: translateX(0); }
        }

        .terminal-prompt {
            color: var(--deel-blue);
            margin-right: 8px;
            font-weight: 600;
        }

        .terminal-text {
            color: #e2e8f0;
        }

        .terminal-success {
            color: #4ade80;
        }

        .terminal-account {
            color: #38bdf8;
        }

        .terminal-cursor-line {
            display: flex;
            align-items: center;
            margin-top: 4px;
            animation: line-appear 0.4s ease-out 1.5s both;
        }

        .cursor {
            display: inline-block;
            width: 9px;
            height: 18px;
            background: var(--deel-blue);
            margin-left: 2px;
            animation: blink 1.2s step-end infinite;
            border-radius: 1px;
        }

        @keyframes blink {
            0%, 50% { opacity: 1; }
            51%, 100% { opacity: 0; }
        }

        .hint {
            font-size: 13px;
            color: var(--deel-gray);
            line-height: 1.5;
            margin-bottom: 20px;
        }

        /* Quick Start */
        .quick-start {
            padding: 16px;
            background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
            border: 1px solid var(--deel-light-gray);
            border-radius: 12px;
            text-align: left;
        }

        .quick-start h3 {
            font-size: 11px;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 0.08em;
            color: var(--deel-navy);
            margin-bottom: 12px;
        }

        .command-row {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 8px;
        }

        .command-row:last-of-type {
            margin-bottom: 0;
        }

        .command-label {
            font-size: 11px;
            font-weight: 600;
            color: var(--deel-gray);
            min-width: 70px;
            text-transform: uppercase;
            letter-spacing: 0.03em;
        }

        .command-code {
            flex: 1;
            background: var(--deel-white);
            border: 1px solid var(--deel-light-gray);
            padding: 8px 12px;
            border-radius: 6px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 11px;
            color: var(--deel-black);
            box-shadow: 0 1px 2px rgba(0,0,0,0.04);
        }

        .tip-box {
            margin-top: 14px;
            padding: 12px 14px;
            background: linear-gradient(135deg, rgba(28, 67, 150, 0.08) 0%, rgba(44, 113, 240, 0.08) 100%);
            border: 1px solid rgba(44, 113, 240, 0.15);
            border-radius: 8px;
            font-size: 12px;
            color: var(--deel-navy);
            line-height: 1.5;
        }

        .tip-box strong {
            font-weight: 700;
            color: var(--deel-blue);
        }

        .tip-box code {
            background: rgba(44, 113, 240, 0.12);
            padding: 2px 6px;
            border-radius: 4px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 10px;
            font-weight: 500;
        }

        .footer {
            text-align: center;
            margin-top: 24px;
            padding-top: 18px;
            border-top: 1px solid var(--deel-light-gray);
        }

        .footer a {
            color: var(--deel-gray);
            text-decoration: none;
            font-size: 12px;
            font-weight: 600;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            transition: all 0.2s ease;
            padding: 6px 12px;
            border-radius: 6px;
        }

        .footer a:hover {
            color: var(--deel-navy);
            background: rgba(28, 67, 150, 0.06);
        }

        .footer svg {
            opacity: 0.7;
        }

        .footer a:hover svg {
            opacity: 1;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"></path>
            </svg>
        </div>
        <h1>Successfully Connected</h1>
        <p class="subtitle">Your Deel account is ready to use.</p>

        <div class="terminal">
            <div class="terminal-header">
                <div class="terminal-dot red"></div>
                <div class="terminal-dot yellow"></div>
                <div class="terminal-dot green"></div>
                <span class="terminal-title">Terminal</span>
            </div>
            <div class="terminal-body">
                <div class="terminal-line">
                    <span class="terminal-prompt">~</span>
                    <span class="terminal-text">deel auth test --account&nbsp;</span><span class="terminal-account">{{.AccountName}}</span>
                </div>
                <div class="terminal-line">
                    <span class="terminal-success">✓ Connection successful!</span>
                </div>
                <div class="terminal-cursor-line">
                    <span class="terminal-prompt">~</span>
                    <span class="cursor"></span>
                </div>
            </div>
        </div>

        <p class="hint">You can close this window and return to your terminal.</p>

        <div class="quick-start">
            <h3>Get Started</h3>
            <div class="command-row">
                <span class="command-label">Test:</span>
                <code class="command-code">deel auth test --account {{.AccountName}}</code>
            </div>
            <div class="command-row">
                <span class="command-label">Contracts:</span>
                <code class="command-code">deel contracts list --account {{.AccountName}}</code>
            </div>
            <div class="tip-box">
                <strong>Pro tip:</strong> Set <code>export DEEL_ACCOUNT={{.AccountName}}</code> to skip the <code>--account</code> flag.
            </div>
        </div>

        <div class="footer">
            <a href="https://github.com/salmonumbrella/deel-cli" target="_blank">
                <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                View on GitHub
            </a>
        </div>
    </div>
    <script>
        // Signal completion to the CLI
        fetch('/complete', {
            method: 'POST',
            headers: { 'X-CSRF-Token': '{{.CSRFToken}}' }
        });
    </script>
</body>
</html>`
