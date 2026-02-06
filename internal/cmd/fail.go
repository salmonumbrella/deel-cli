package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

// fail emits a structured error in agent+JSON mode, and returns a Go error.
// For non-agent invocations, it prints a human-friendly error to stderr.
func fail(cmd *cobra.Command, f *outfmt.Formatter, operation, category, message string, suggestions ...string) error {
	if f == nil {
		return fmt.Errorf("%s", message)
	}

	// Human/debug output.
	f.PrintError("%s", message)
	for _, s := range suggestions {
		if s == "" {
			continue
		}
		f.PrintText("  -> " + s)
	}

	// Agent structured error on stdout.
	if cmd != nil && outfmt.IsAgent(cmd.Context()) && f.IsJSON() && !AgentErrorEmitted() {
		_ = f.PrintJSON(map[string]any{
			"ok": false,
			"error": map[string]any{
				"operation":   operation,
				"category":    category,
				"message":     message,
				"suggestions": suggestions,
			},
		})
		markAgentErrorEmitted()
	}

	return fmt.Errorf("%s", message)
}

func failValidation(cmd *cobra.Command, f *outfmt.Formatter, message string, suggestions ...string) error {
	return fail(cmd, f, "validating input", "validation", message, suggestions...)
}
