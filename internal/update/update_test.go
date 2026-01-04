package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already has v prefix",
			input:    "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "no v prefix",
			input:    "1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "v",
		},
		{
			name:     "prerelease",
			input:    "1.0.0-beta.1",
			expected: "v1.0.0-beta.1",
		},
		{
			name:     "with build metadata",
			input:    "v1.0.0+build.123",
			expected: "v1.0.0+build.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckForUpdate_DevBuild(t *testing.T) {
	ctx := context.Background()
	result, err := CheckForUpdate(ctx, "dev")

	require.NoError(t, err)
	assert.False(t, result.UpdateAvailable)
	assert.Equal(t, "dev", result.CurrentVersion)
	assert.Empty(t, result.LatestVersion)
}

func TestCheckForUpdate_EmptyVersion(t *testing.T) {
	ctx := context.Background()
	result, err := CheckForUpdate(ctx, "")

	require.NoError(t, err)
	assert.False(t, result.UpdateAvailable)
	assert.Empty(t, result.CurrentVersion)
	assert.Empty(t, result.LatestVersion)
}

func TestCheckForUpdate_UpdateAvailable(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v2.0.0",
			HTMLURL: "https://github.com/salmonumbrella/deel-cli/releases/tag/v2.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Override the URL for testing
	originalURL := GitHubReleasesURL
	GitHubReleasesURL = server.URL
	defer func() { GitHubReleasesURL = originalURL }()

	ctx := context.Background()
	result, err := CheckForUpdate(ctx, "v1.0.0")

	require.NoError(t, err)
	assert.True(t, result.UpdateAvailable)
	assert.Equal(t, "v1.0.0", result.CurrentVersion)
	assert.Equal(t, "v2.0.0", result.LatestVersion)
	assert.Contains(t, result.UpdateURL, "github.com")
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v1.0.0",
			HTMLURL: "https://github.com/salmonumbrella/deel-cli/releases/tag/v1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Override the URL for testing
	originalURL := GitHubReleasesURL
	GitHubReleasesURL = server.URL
	defer func() { GitHubReleasesURL = originalURL }()

	ctx := context.Background()
	result, err := CheckForUpdate(ctx, "v1.0.0")

	require.NoError(t, err)
	assert.False(t, result.UpdateAvailable)
	assert.Equal(t, "v1.0.0", result.CurrentVersion)
	assert.Equal(t, "v1.0.0", result.LatestVersion)
}

func TestCheckForUpdate_Timeout(t *testing.T) {
	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		release := Release{
			TagName: "v2.0.0",
			HTMLURL: "https://github.com/salmonumbrella/deel-cli/releases/tag/v2.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Override the URL for testing
	originalURL := GitHubReleasesURL
	GitHubReleasesURL = server.URL
	defer func() { GitHubReleasesURL = originalURL }()

	// Use a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := CheckForUpdate(ctx, "v1.0.0")
	assert.Error(t, err)
}

func TestCheckForUpdate_InvalidJSON(t *testing.T) {
	// Create mock server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("not json")); err != nil {
			return
		}
	}))
	defer server.Close()

	// Override the URL for testing
	originalURL := GitHubReleasesURL
	GitHubReleasesURL = server.URL
	defer func() { GitHubReleasesURL = originalURL }()

	ctx := context.Background()
	_, err := CheckForUpdate(ctx, "v1.0.0")
	assert.Error(t, err)
}

func TestCheckForUpdate_ServerError(t *testing.T) {
	// Create mock server returning error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Override the URL for testing
	originalURL := GitHubReleasesURL
	GitHubReleasesURL = server.URL
	defer func() { GitHubReleasesURL = originalURL }()

	ctx := context.Background()
	_, err := CheckForUpdate(ctx, "v1.0.0")
	assert.Error(t, err)
}
