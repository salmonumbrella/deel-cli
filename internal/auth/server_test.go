package auth

import (
	"testing"
	"time"
)

func TestValidateAccountName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "production", false},
		{"valid with numbers", "prod123", false},
		{"valid with dash", "my-account", false},
		{"valid with underscore", "my_account", false},
		{"valid mixed", "My-Account_123", false},
		{"empty", "", true},
		{"too long - 65 chars", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},
		{"max length - 64 chars", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
		{"invalid space", "my account", true},
		{"invalid special char", "account!", true},
		{"invalid dot", "my.account", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccountName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAccountName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid token", "abc123xyz", false},
		{"empty", "", true},
		{"too long - 4097 chars", string(make([]byte, 4097)), true},
		{"max length - 4096 chars", string(make([]byte, 4096)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			// For byte slice tests, fill with valid characters
			if tt.name == "too long - 4097 chars" {
				b := make([]byte, 4097)
				for i := range b {
					b[i] = 'a'
				}
				input = string(b)
			}
			if tt.name == "max length - 4096 chars" {
				b := make([]byte, 4096)
				for i := range b {
					b[i] = 'a'
				}
				input = string(b)
			}

			err := ValidateToken(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken(%q...) error = %v, wantErr %v", input[:min(10, len(input))], err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "abc123", "abc123"},
		{"leading spaces", "  token", "token"},
		{"trailing spaces", "token  ", "token"},
		{"both spaces", "  token  ", "token"},
		{"tabs", "\ttoken\t", "token"},
		{"newlines", "\ntoken\n", "token"},
		{"null byte", "tok\x00en", "token"},
		{"control chars", "tok\x1Fen", "token"},
		{"DEL char", "tok\x7Fen", "token"},
		{"mixed", "  \x00hello\x7F world\t\n", "hello world"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeToken(tt.input); got != tt.want {
				t.Errorf("SanitizeToken(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	t.Run("allows requests under limit", func(t *testing.T) {
		rl := newRateLimiter(3, time.Minute)
		for i := 0; i < 3; i++ {
			if err := rl.check("127.0.0.1", "/test"); err != nil {
				t.Errorf("request %d should be allowed: %v", i+1, err)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		rl := newRateLimiter(2, time.Minute)
		_ = rl.check("127.0.0.1", "/test")
		_ = rl.check("127.0.0.1", "/test")
		if err := rl.check("127.0.0.1", "/test"); err == nil {
			t.Error("third request should be blocked")
		}
	})

	t.Run("separate limits per endpoint", func(t *testing.T) {
		rl := newRateLimiter(1, time.Minute)
		if err := rl.check("127.0.0.1", "/endpoint1"); err != nil {
			t.Errorf("endpoint1 should be allowed: %v", err)
		}
		if err := rl.check("127.0.0.1", "/endpoint2"); err != nil {
			t.Errorf("endpoint2 should be allowed: %v", err)
		}
	})

	t.Run("separate limits per IP", func(t *testing.T) {
		rl := newRateLimiter(1, time.Minute)
		if err := rl.check("127.0.0.1", "/test"); err != nil {
			t.Errorf("IP1 should be allowed: %v", err)
		}
		if err := rl.check("127.0.0.2", "/test"); err != nil {
			t.Errorf("IP2 should be allowed: %v", err)
		}
	})
}
