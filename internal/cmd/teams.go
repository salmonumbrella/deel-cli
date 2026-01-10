package cmd

import (
	"context"
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
		client, f, err := initClient("listing teams")
		if err != nil {
			return err
		}

		teams, _, hasMore, err := collectCursorItems(cmd.Context(), teamsAllFlag, teamsCursorFlag, teamsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.Team], error) {
			resp, err := client.ListTeams(ctx, limit, cursor)
			if err != nil {
				return CursorListResult[api.Team]{}, err
			}
			return CursorListResult[api.Team]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing teams")
		}

		response := api.TeamsListResponse{
			Data: teams,
		}
		response.Page.Next = ""

		return outputList(cmd, f, teams, hasMore, "No teams found.", []string{"ID", "NAME", "MANAGER", "MEMBERS"}, func(t api.Team) []string {
			return []string{t.ID, t.Name, t.ManagerName, fmt.Sprintf("%d", t.MemberCount)}
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
			return HandleError(f, err, "initializing client")
		}

		team, err := client.GetTeam(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "get team")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
	teamsListCmd.Flags().IntVar(&teamsLimitFlag, "limit", 100, "Maximum results")
	teamsListCmd.Flags().StringVar(&teamsCursorFlag, "cursor", "", "Pagination cursor")
	teamsListCmd.Flags().BoolVar(&teamsAllFlag, "all", false, "Fetch all pages")

	teamsCmd.AddCommand(teamsListCmd)
	teamsCmd.AddCommand(teamsGetCmd)
}
