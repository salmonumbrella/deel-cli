package cmd

import (
	"github.com/spf13/cobra"
)

var offboardingCmd = &cobra.Command{
	Use:   "offboarding",
	Short: "Manage offboarding and terminations",
	Long:  "View offboarding records and termination details.",
}

var offboardingGetCmd = &cobra.Command{
	Use:   "get <tracker-id>",
	Short: "Get offboarding tracker details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		record, err := client.GetOffboardingTracker(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get offboarding: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:             " + record.ID)
			f.PrintText("Contract ID:    " + record.ContractID)
			f.PrintText("Worker:         " + record.WorkerName)
			f.PrintText("Type:           " + record.Type)
			f.PrintText("Status:         " + record.Status)
			f.PrintText("Effective Date: " + record.EffectiveDate)
			f.PrintText("Created:        " + record.CreatedAt)
		}, record)
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
	offboardingCmd.AddCommand(offboardingGetCmd)
	offboardingCmd.AddCommand(terminationsGetCmd)
}
