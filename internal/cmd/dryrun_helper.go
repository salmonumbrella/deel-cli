package cmd

import (
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/dryrun"
	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

func handleDryRun(cmd *cobra.Command, f *outfmt.Formatter, preview *dryrun.Preview) (bool, error) {
	if !dryrun.IsEnabled(cmd.Context()) {
		return false, nil
	}
	if preview == nil {
		preview = &dryrun.Preview{
			Operation:   "DRY_RUN",
			Resource:    "request",
			Description: "Dry-run enabled; no changes were made.",
		}
	}
	return true, f.PrintDryRun(preview)
}
