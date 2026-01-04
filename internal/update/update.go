// Package update provides version checking against GitHub releases.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// GitHubReleasesURL is the endpoint for checking latest releases.
// Exported as a var to allow testing.
var GitHubReleasesURL = "https://api.github.com/repos/salmonumbrella/deel-cli/releases/latest"

// CheckTimeout is the timeout for update checks.
const CheckTimeout = 5 * time.Second

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckResult contains the result of an update check.
type CheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateURL       string
	UpdateAvailable bool
}

// CheckForUpdate checks if a newer version is available on GitHub.
func CheckForUpdate(ctx context.Context, currentVersion string) (*CheckResult, error) {
	result := &CheckResult{
		CurrentVersion: currentVersion,
	}

	// Skip for dev builds or empty versions
	if currentVersion == "" || currentVersion == "dev" {
		return result, nil
	}

	// Create a timeout context if not already done
	checkCtx, cancel := context.WithTimeout(ctx, CheckTimeout)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, GitHubReleasesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "deel-cli")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse response
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	result.LatestVersion = release.TagName
	result.UpdateURL = release.HTMLURL

	// Compare versions
	current := normalizeVersion(currentVersion)
	latest := normalizeVersion(release.TagName)

	if semver.IsValid(current) && semver.IsValid(latest) {
		if semver.Compare(current, latest) < 0 {
			result.UpdateAvailable = true
		}
	}

	return result, nil
}

// normalizeVersion ensures a version string has a "v" prefix for semver comparison.
func normalizeVersion(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}
