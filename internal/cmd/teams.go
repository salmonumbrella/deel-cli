package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage teams",
	Long:  "List and view team information.",
}

var (
	teamsLimitFlag  int
	teamsCursorFlag string
)

var teamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListTeams(cmd.Context(), teamsLimitFlag, teamsCursorFlag)
		if err != nil {
			f.PrintError("Failed to list teams: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No teams found.")
				return
			}
			table := f.NewTable("ID", "NAME", "MANAGER", "MEMBERS")
			for _, t := range resp.Data {
				table.AddRow(t.ID, t.Name, t.ManagerName, fmt.Sprintf("%d", t.MemberCount))
			}
			table.Render()
		}, resp)
	},
}

var teamsGetCmd = &cobra.Command{
	Use:   "get <team-id>",
	Short: "Get team details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		team, err := client.GetTeam(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get team: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:          " + team.ID)
			f.PrintText("Name:        " + team.Name)
			f.PrintText("Manager:     " + team.ManagerName)
			f.PrintText(fmt.Sprintf("Members:     %d", team.MemberCount))
			if team.Description != "" {
				f.PrintText("Description: " + team.Description)
			}
		}, team)
	},
}

func init() {
	teamsListCmd.Flags().IntVar(&teamsLimitFlag, "limit", 50, "Maximum results")
	teamsListCmd.Flags().StringVar(&teamsCursorFlag, "cursor", "", "Pagination cursor")

	teamsCmd.AddCommand(teamsListCmd)
	teamsCmd.AddCommand(teamsGetCmd)
}
