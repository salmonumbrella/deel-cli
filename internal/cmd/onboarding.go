package cmd

import (
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
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "listing onboarding")
		}

		cursor := onboardingCursorFlag
		var allEmployees []api.OnboardingEmployee
		var next string

		for {
			resp, err := client.ListOnboardingEmployees(cmd.Context(), api.OnboardingListParams{
				Status: onboardingStatusFlag,
				Limit:  onboardingLimitFlag,
				Cursor: cursor,
			})
			if err != nil {
				return HandleError(f, err, "listing onboarding")
			}
			allEmployees = append(allEmployees, resp.Data...)
			next = resp.Page.Next
			if !onboardingAllFlag || next == "" {
				if !onboardingAllFlag {
					allEmployees = resp.Data
				}
				break
			}
			cursor = next
		}

		response := api.OnboardingListResponse{
			Data: allEmployees,
		}
		response.Page.Next = ""

		return f.Output(func() {
			if len(allEmployees) == 0 {
				f.PrintText("No employees in onboarding.")
				return
			}
			table := f.NewTable("ID", "NAME", "COUNTRY", "STATUS", "STAGE", "PROGRESS")
			for _, e := range allEmployees {
				table.AddRow(e.ID, e.Name, e.Country, e.Status, e.Stage, fmt.Sprintf("%d%%", e.Progress))
			}
			table.Render()
			if !onboardingAllFlag && next != "" {
				f.PrintText("")
				f.PrintText("More results available. Use --cursor to paginate or --all to fetch everything.")
			}
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

		return f.Output(func() {
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
