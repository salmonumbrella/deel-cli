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
)

var onboardingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List employees in onboarding",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListOnboardingEmployees(cmd.Context(), api.OnboardingListParams{
			Status: onboardingStatusFlag,
			Limit:  onboardingLimitFlag,
			Cursor: onboardingCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list onboarding: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No employees in onboarding.")
				return
			}
			table := f.NewTable("ID", "NAME", "COUNTRY", "STATUS", "STAGE", "PROGRESS")
			for _, e := range resp.Data {
				table.AddRow(e.ID, e.Name, e.Country, e.Status, e.Stage, fmt.Sprintf("%d%%", e.Progress))
			}
			table.Render()
		}, resp)
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
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		details, err := client.GetOnboardingDetails(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get details: %v", err)
			return err
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
	onboardingListCmd.Flags().IntVar(&onboardingLimitFlag, "limit", 50, "Maximum results")
	onboardingListCmd.Flags().StringVar(&onboardingCursorFlag, "cursor", "", "Pagination cursor")

	onboardingCmd.AddCommand(onboardingListCmd)
	onboardingCmd.AddCommand(onboardingGetCmd)
}
