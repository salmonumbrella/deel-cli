package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

func requireForce(cmd *cobra.Command, f *outfmt.Formatter, force bool, action, resource, id string, suggestedCommand string) (bool, error) { //nolint:unparam // action kept as param for future extensibility
	if force {
		return true, nil
	}

	if outfmt.IsAgent(cmd.Context()) && f.IsJSON() && !AgentErrorEmitted() {
		suggestions := []string{}
		if suggestedCommand != "" {
			suggestions = append(suggestions, suggestedCommand)
		}
		_ = f.PrintJSON(map[string]any{
			"ok": false,
			"error": map[string]any{
				"category": "confirmation_required",
				"message":  fmt.Sprintf("%s requires confirmation; rerun with --force", action),
				"details": map[string]any{
					"action":   action,
					"resource": resource,
					"id":       id,
					"flag":     "--force",
				},
				"suggestions": suggestions,
			},
		})
		markAgentErrorEmitted()
		return false, fmt.Errorf("confirmation required")
	}

	f.PrintText(fmt.Sprintf("Are you sure you want to %s %s %s?", action, resource, id))
	f.PrintText("Use --force to confirm.")
	return false, nil
}
