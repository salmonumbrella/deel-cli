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
			f.PrintWarning("Could not check for updates: %v", err)
			return nil
		}

		if !result.UpdateAvailable {
			if result.LatestVersion != "" {
				f.PrintSuccess("deel-cli is up to date (version %s)", result.CurrentVersion)
			} else {
				f.PrintText("Running a development build, skipping update check")
			}
			return nil
		}

		f.PrintText(fmt.Sprintf("Update available: %s -> %s", result.CurrentVersion, result.LatestVersion))

		// Try to upgrade based on platform
		if runtime.GOOS == "darwin" && isBrewInstalled() {
			f.PrintText("Attempting upgrade via Homebrew...")
			if err := runBrewUpgrade(); err != nil {
				f.PrintWarning("Homebrew upgrade failed: %v", err)
				showManualInstructions(f, result)
			} else {
				f.PrintSuccess("Successfully upgraded to %s", result.LatestVersion)
			}
			return nil
		}

		showManualInstructions(f, result)
		return nil
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
