package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var milestonesCmd = &cobra.Command{
	Use:     "milestones",
	Aliases: []string{"milestone", "ms"},
	Short:   "Manage contract milestones",
	Long:    "Create, list, and delete milestones for milestone-based contracts.",
}

var (
	milestonesContractIDFlag string
	milestonesLimitFlag      int
)

var milestonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List milestones for a contract",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if milestonesContractIDFlag == "" {
			return failValidation(cmd, f, "--contract-id is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		milestones, err := client.ListMilestones(cmd.Context(), milestonesContractIDFlag)
		if err != nil {
			return HandleError(f, err, "list milestones")
		}

		// Apply client-side limit
		if milestonesLimitFlag > 0 && len(milestones) > milestonesLimitFlag {
			milestones = milestones[:milestonesLimitFlag]
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(milestones) == 0 {
				f.PrintText("No milestones found.")
				return
			}
			table := f.NewTable("ID", "TITLE", "AMOUNT", "STATUS", "DUE DATE")
			for _, m := range milestones {
				table.AddRow(m.ID, m.Title, fmt.Sprintf("%.2f %s", m.Amount, m.Currency), m.Status, m.DueDate)
			}
			table.Render()
		}, milestones)
	},
}

var (
	milestonesTitleFlag       string
	milestonesDescriptionFlag string
	milestonesAmountFlag      float64
	milestonesDueDateFlag     string
	milestonesForceFlag       bool
)

var milestonesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new milestone",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if milestonesContractIDFlag == "" {
			return failValidation(cmd, f, "--contract-id is required")
		}
		if milestonesTitleFlag == "" {
			return failValidation(cmd, f, "--title is required")
		}
		if milestonesAmountFlag <= 0 {
			return failValidation(cmd, f, "--amount is required and must be positive")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Milestone",
			Description: "Create milestone",
			Details: map[string]string{
				"ContractID":  milestonesContractIDFlag,
				"Title":       milestonesTitleFlag,
				"Description": milestonesDescriptionFlag,
				"Amount":      fmt.Sprintf("%.2f", milestonesAmountFlag),
				"DueDate":     milestonesDueDateFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		milestone, err := client.CreateMilestone(cmd.Context(), api.CreateMilestoneParams{
			ContractID:  milestonesContractIDFlag,
			Title:       milestonesTitleFlag,
			Description: milestonesDescriptionFlag,
			Amount:      milestonesAmountFlag,
			DueDate:     milestonesDueDateFlag,
		})
		if err != nil {
			return HandleError(f, err, "create milestone")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Milestone created successfully!")
			f.PrintText("ID:     " + milestone.ID)
			f.PrintText("Title:  " + milestone.Title)
		}, milestone)
	},
}

var milestonesDeleteCmd = &cobra.Command{
	Use:   "delete <milestone-id>",
	Short: "Delete a milestone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "Milestone",
			Description: "Delete milestone",
			Details: map[string]string{
				"ID": args[0],
			},
		}); ok {
			return err
		}

		if ok, err := requireForce(cmd, f, milestonesForceFlag, "delete", "milestone", args[0], "deel milestones delete "+args[0]+" --force"); !ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		if err := client.DeleteMilestone(cmd.Context(), args[0]); err != nil {
			return HandleError(f, err, "delete milestone")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Milestone deleted successfully.")
		}, map[string]any{
			"deleted":      true,
			"milestone_id": args[0],
		})
	},
}

func init() {
	milestonesListCmd.Flags().StringVar(&milestonesContractIDFlag, "contract-id", "", "Contract ID (required)")
	milestonesListCmd.Flags().IntVar(&milestonesLimitFlag, "limit", 100, "Maximum results")

	milestonesCreateCmd.Flags().StringVar(&milestonesContractIDFlag, "contract-id", "", "Contract ID (required)")
	milestonesCreateCmd.Flags().StringVar(&milestonesTitleFlag, "title", "", "Milestone title (required)")
	milestonesCreateCmd.Flags().StringVar(&milestonesDescriptionFlag, "description", "", "Milestone description")
	milestonesCreateCmd.Flags().Float64Var(&milestonesAmountFlag, "amount", 0, "Milestone amount (required)")
	milestonesCreateCmd.Flags().StringVar(&milestonesDueDateFlag, "due-date", "", "Due date (YYYY-MM-DD)")

	milestonesDeleteCmd.Flags().BoolVar(&milestonesForceFlag, "force", false, "Confirm deletion")

	milestonesCmd.AddCommand(milestonesListCmd)
	milestonesCmd.AddCommand(milestonesCreateCmd)
	milestonesCmd.AddCommand(milestonesDeleteCmd)
}
