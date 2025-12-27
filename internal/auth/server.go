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
	if _, err := client.Get(ctx, "/rest/v2/profiles/me"); err != nil {
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

	// Normalize inputs
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.Token = strings.TrimSpace(req.Token)

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

	// Normalize inputs
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.Token = strings.TrimSpace(req.Token)

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
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: #fff;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 480px;
            width: 100%;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo h1 {
            font-size: 28px;
            color: #1a1a2e;
            font-weight: 700;
        }
        .logo p {
            color: #666;
            margin-top: 8px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            font-weight: 600;
            margin-bottom: 8px;
            color: #333;
        }
        input {
            width: 100%;
            padding: 14px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.2s, box-shadow 0.2s;
        }
        input:focus {
            outline: none;
            border-color: #4F46E5;
            box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
        }
        .hint {
            font-size: 13px;
            color: #666;
            margin-top: 6px;
        }
        .hint a {
            color: #4F46E5;
            text-decoration: none;
        }
        .hint a:hover {
            text-decoration: underline;
        }
        .buttons {
            display: flex;
            gap: 12px;
            margin-top: 30px;
        }
        button {
            flex: 1;
            padding: 14px 24px;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s;
        }
        .btn-secondary {
            background: #f0f0f0;
            color: #333;
        }
        .btn-secondary:hover {
            background: #e0e0e0;
        }
        .btn-primary {
            background: #4F46E5;
            color: #fff;
        }
        .btn-primary:hover {
            background: #4338CA;
        }
        .btn-primary:disabled {
            background: #a5a5a5;
            cursor: not-allowed;
        }
        .status {
            margin-top: 20px;
            padding: 12px 16px;
            border-radius: 8px;
            font-size: 14px;
            display: none;
        }
        .status.error {
            background: #FEE2E2;
            color: #DC2626;
            display: block;
        }
        .status.success {
            background: #D1FAE5;
            color: #059669;
            display: block;
        }
        .status.loading {
            background: #E0E7FF;
            color: #4F46E5;
            display: block;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <h1>Deel CLI</h1>
            <p>Connect your Deel account</p>
        </div>
        <form id="setupForm">
            <input type="hidden" id="csrfToken" value="{{.CSRFToken}}">
            <div class="form-group">
                <label for="accountName">Account Name</label>
                <input type="text" id="accountName" placeholder="e.g., production, sandbox" required>
                <p class="hint">A friendly name to identify this account</p>
            </div>
            <div class="form-group">
                <label for="token">Personal Access Token</label>
                <input type="password" id="token" placeholder="Paste your PAT here" required>
                <p class="hint">Get your token from <a href="https://app.deel.com/settings/api" target="_blank">Deel Settings → API</a></p>
            </div>
            <div class="buttons">
                <button type="button" class="btn-secondary" onclick="testConnection()">Test</button>
                <button type="submit" class="btn-primary" id="saveBtn">Save & Connect</button>
            </div>
            <div id="status" class="status"></div>
        </form>
    </div>
    <script>
        const csrfToken = document.getElementById('csrfToken').value;
        const statusEl = document.getElementById('status');
        const saveBtn = document.getElementById('saveBtn');

        function showStatus(type, message) {
            statusEl.className = 'status ' + type;
            statusEl.textContent = message;
        }

        async function testConnection() {
            const accountName = document.getElementById('accountName').value.trim();
            const token = document.getElementById('token').value.trim();

            if (!accountName || !token) {
                showStatus('error', 'Please fill in all fields');
                return;
            }

            showStatus('loading', 'Testing connection...');

            try {
                const resp = await fetch('/validate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify({ account_name: accountName, token: token })
                });
                const data = await resp.json();
                if (data.success) {
                    showStatus('success', data.message);
                } else {
                    showStatus('error', data.error);
                }
            } catch (e) {
                showStatus('error', 'Connection failed: ' + e.message);
            }
        }

        document.getElementById('setupForm').addEventListener('submit', async (e) => {
            e.preventDefault();

            const accountName = document.getElementById('accountName').value.trim();
            const token = document.getElementById('token').value.trim();

            if (!accountName || !token) {
                showStatus('error', 'Please fill in all fields');
                return;
            }

            saveBtn.disabled = true;
            showStatus('loading', 'Saving credentials...');

            try {
                const resp = await fetch('/submit', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify({ account_name: accountName, token: token })
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
    </script>
</body>
</html>`

var successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Deel CLI - Success</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: #fff;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 480px;
            width: 100%;
            text-align: center;
        }
        .checkmark {
            width: 80px;
            height: 80px;
            background: #D1FAE5;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 24px;
        }
        .checkmark svg {
            width: 40px;
            height: 40px;
            stroke: #059669;
        }
        h1 {
            font-size: 24px;
            color: #1a1a2e;
            margin-bottom: 12px;
        }
        p {
            color: #666;
            margin-bottom: 24px;
        }
        .account-name {
            background: #f5f5f5;
            padding: 12px 20px;
            border-radius: 8px;
            font-family: monospace;
            font-size: 16px;
            color: #333;
            margin-bottom: 24px;
        }
        .hint {
            font-size: 14px;
            color: #888;
        }
        code {
            background: #f0f0f0;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 13px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
            </svg>
        </div>
        <h1>Successfully Connected!</h1>
        <p>Your Deel account has been configured.</p>
        <div class="account-name">{{.AccountName}}</div>
        <p class="hint">You can close this window and return to your terminal.<br>Try <code>deel auth test</code> to verify.</p>
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
