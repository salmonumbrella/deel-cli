package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var onboardingCmd = &cobra.Command{
	Use:   "onboarding",
	Short: "Manage employee onboarding",
	Long:  "View onboarding status and details for employees.",
}

var (
	onboardingStatusFlag string
	onboardingLimitFlag  int
	onboardingCursorFlag string
	onboardingAllFlag    bool
)

var onboardingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List employees in onboarding",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, f, err := initClient("listing onboarding")
		if err != nil {
			return err
		}

		employees, page, hasMore, err := collectCursorItems(cmd.Context(), onboardingAllFlag, onboardingCursorFlag, onboardingLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.OnboardingEmployee], error) {
			resp, err := client.ListOnboardingEmployees(ctx, api.OnboardingListParams{
				Status: onboardingStatusFlag,
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.OnboardingEmployee]{}, err
			}
			return CursorListResult[api.OnboardingEmployee]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing onboarding")
		}

		response := makeListResponse(employees, page)

		return outputList(cmd, f, employees, hasMore, "No employees in onboarding.", []string{"ID", "NAME", "COUNTRY", "STATUS", "STAGE", "PROGRESS"}, func(e api.OnboardingEmployee) []string {
			return []string{e.ID, e.Name, e.Country, e.Status, e.Stage, fmt.Sprintf("%d%%", e.Progress)}
		}, response)
	},
}

var onboardingGetCmd = &cobra.Command{
	Use:   "get <employee-id>",
	Short: "Get onboarding details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "getting onboarding details")
		}

		details, err := client.GetOnboardingDetails(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting onboarding details")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("Employee:     " + details.EmployeeName)
			f.PrintText("Status:       " + details.Status)
			f.PrintText("Stage:        " + details.Stage)
			f.PrintText(fmt.Sprintf("Progress:     %d%%", details.Progress))
			f.PrintText("Start Date:   " + details.StartDate)
			f.PrintText("Est. End:     " + details.EstimatedEnd)
			f.PrintText(fmt.Sprintf("Pending:      %d tasks", len(details.PendingTasks)))
			f.PrintText(fmt.Sprintf("Completed:    %d tasks", len(details.CompletedTasks)))
		}, details)
	},
}

func init() {
	onboardingListCmd.Flags().StringVar(&onboardingStatusFlag, "status", "", "Filter by status")
	onboardingListCmd.Flags().IntVar(&onboardingLimitFlag, "limit", 100, "Maximum results")
	onboardingListCmd.Flags().StringVar(&onboardingCursorFlag, "cursor", "", "Pagination cursor")
	onboardingListCmd.Flags().BoolVar(&onboardingAllFlag, "all", false, "Fetch all pages")

	onboardingCmd.AddCommand(onboardingListCmd)
	onboardingCmd.AddCommand(onboardingGetCmd)
}
