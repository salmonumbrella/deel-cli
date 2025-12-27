package cmd

import (
	"github.com/spf13/cobra"
)

var offboardingCmd = &cobra.Command{
	Use:   "offboarding",
	Short: "Manage offboarding and terminations",
	Long:  "View offboarding records and termination details.",
}

var offboardingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all offboarding records",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		records, err := client.ListOffboarding(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list offboarding: %v", err)
			return err
		}

		return f.Output(func() {
			if len(records) == 0 {
				f.PrintText("No offboarding records found.")
				return
			}
			table := f.NewTable("ID", "WORKER", "TYPE", "STATUS", "EFFECTIVE DATE", "CREATED")
			for _, r := range records {
				table.AddRow(r.ID, r.WorkerName, r.Type, r.Status, r.EffectiveDate, r.CreatedAt)
			}
			table.Render()
		}, records)
	},
}

var terminationsGetCmd = &cobra.Command{
	Use:   "termination <termination-id>",
	Short: "Get termination details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		termination, err := client.GetTerminationDetails(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get termination: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:             " + termination.ID)
			f.PrintText("Contract ID:    " + termination.ContractID)
			f.PrintText("Reason:         " + termination.Reason)
			f.PrintText("Status:         " + termination.Status)
			f.PrintText("Notice Date:    " + termination.NoticeDate)
			f.PrintText("Effective Date: " + termination.EffectiveDate)
			if termination.FinalPayDate != "" {
				f.PrintText("Final Pay Date: " + termination.FinalPayDate)
			}
		}, termination)
	},
}

func init() {
	// Add subcommands
	offboardingCmd.AddCommand(offboardingListCmd)
	offboardingCmd.AddCommand(terminationsGetCmd)
}
