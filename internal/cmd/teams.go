package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage teams",
	Long:  "List and view team information.",
}

var (
	teamsLimitFlag  int
	teamsCursorFlag string
	teamsAllFlag    bool
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

		cursor := teamsCursorFlag
		var allTeams []api.Team
		var next string

		for {
			resp, err := client.ListTeams(cmd.Context(), teamsLimitFlag, cursor)
			if err != nil {
				f.PrintError("Failed to list teams: %v", err)
				return err
			}
			allTeams = append(allTeams, resp.Data...)
			next = resp.Page.Next
			if !teamsAllFlag || next == "" {
				if !teamsAllFlag {
					allTeams = resp.Data
				}
				break
			}
			cursor = next
		}

		response := api.TeamsListResponse{
			Data: allTeams,
		}
		response.Page.Next = ""

		return f.Output(func() {
			if len(allTeams) == 0 {
				f.PrintText("No teams found.")
				return
			}
			table := f.NewTable("ID", "NAME", "MANAGER", "MEMBERS")
			for _, t := range allTeams {
				table.AddRow(t.ID, t.Name, t.ManagerName, fmt.Sprintf("%d", t.MemberCount))
			}
			table.Render()
			if !teamsAllFlag && next != "" {
				f.PrintText("")
				f.PrintText("More results available. Use --cursor to paginate or --all to fetch everything.")
			}
		}, response)
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
	teamsListCmd.Flags().BoolVar(&teamsAllFlag, "all", false, "Fetch all pages")

	teamsCmd.AddCommand(teamsListCmd)
	teamsCmd.AddCommand(teamsGetCmd)
}
