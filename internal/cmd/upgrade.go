package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/outfmt"
	"github.com/salmonumbrella/deel-cli/internal/update"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade deel-cli to the latest version",
	Long: `Check for updates and upgrade deel-cli to the latest version.

On macOS with Homebrew, this will run 'brew upgrade deel-cli'.
On other platforms, it will show manual upgrade instructions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Check for updates
		result, err := update.CheckForUpdate(cmd.Context(), Version)
		if err != nil {
			return f.OutputFiltered(cmd.Context(), func() {
				f.PrintWarning("Could not check for updates: %v", err)
			}, map[string]any{
				"checked": false,
				"error":   err.Error(),
			})
		}

		payload := map[string]any{
			"checked":          true,
			"current_version":  result.CurrentVersion,
			"latest_version":   result.LatestVersion,
			"update_url":       result.UpdateURL,
			"update_available": result.UpdateAvailable,
		}

		if !result.UpdateAvailable {
			return f.OutputFiltered(cmd.Context(), func() {
				if result.LatestVersion != "" {
					f.PrintSuccess("deel-cli is up to date (version %s)", result.CurrentVersion)
				} else {
					f.PrintText("Running a development build, skipping update check")
				}
			}, payload)
		}

		// Try to upgrade based on platform
		if runtime.GOOS == "darwin" && isBrewInstalled() {
			payload["upgrade"] = map[string]any{
				"attempted": true,
				"method":    "homebrew",
				"success":   false,
			}
			if err := runBrewUpgrade(); err != nil {
				payload["upgrade"].(map[string]any)["error"] = err.Error()
				return f.OutputFiltered(cmd.Context(), func() {
					f.PrintText(fmt.Sprintf("Update available: %s -> %s", result.CurrentVersion, result.LatestVersion))
					f.PrintText("Attempting upgrade via Homebrew...")
					f.PrintWarning("Homebrew upgrade failed: %v", err)
					showManualInstructions(f, result)
				}, payload)
			}

			payload["upgrade"].(map[string]any)["success"] = true
			return f.OutputFiltered(cmd.Context(), func() {
				f.PrintText(fmt.Sprintf("Update available: %s -> %s", result.CurrentVersion, result.LatestVersion))
				f.PrintText("Attempting upgrade via Homebrew...")
				f.PrintSuccess("Successfully upgraded to %s", result.LatestVersion)
			}, payload)
		}

		payload["upgrade"] = map[string]any{
			"attempted": false,
			"method":    "manual",
		}
		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText(fmt.Sprintf("Update available: %s -> %s", result.CurrentVersion, result.LatestVersion))
			showManualInstructions(f, result)
		}, payload)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

// isBrewInstalled checks if Homebrew is available on the system.
func isBrewInstalled() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// runBrewUpgrade attempts to upgrade deel-cli via Homebrew.
func runBrewUpgrade() error {
	cmd := exec.Command("brew", "upgrade", "deel-cli")
	return cmd.Run()
}

// showManualInstructions displays manual upgrade instructions.
func showManualInstructions(f *outfmt.Formatter, result *update.CheckResult) {
	f.PrintText("\nTo upgrade manually:")
	f.PrintText(fmt.Sprintf("  1. Download the latest release from: %s", result.UpdateURL))
	f.PrintText("  2. Extract and replace your current binary")
	f.PrintText("")
	f.PrintText("Or install via go:")
	f.PrintText("  go install github.com/salmonumbrella/deel-cli@latest")
}
