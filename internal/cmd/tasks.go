package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"task"},
	Short:   "Manage contractor tasks",
	Long:    "Create, list, review, and delete tasks for task-based contracts.",
}

var (
	tasksContractIDFlag  string
	tasksStatusFlag      string
	tasksLimitFlag       int
	tasksTitleFlag       string
	tasksDescriptionFlag string
	tasksAmountFlag      float64
	tasksForceFlag       bool
)

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListTasks(cmd.Context(), api.TasksListParams{
			ContractID: tasksContractIDFlag,
			Status:     tasksStatusFlag,
			Limit:      tasksLimitFlag,
		})
		if err != nil {
			f.PrintError("Failed to list tasks: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No tasks found.")
				return
			}
			table := f.NewTable("ID", "TITLE", "AMOUNT", "STATUS")
			for _, t := range resp.Data {
				table.AddRow(t.ID, t.Title, fmt.Sprintf("%.2f %s", t.Amount, t.Currency), t.Status)
			}
			table.Render()
		}, resp)
	},
}

var tasksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if tasksContractIDFlag == "" {
			f.PrintError("--contract-id is required")
			return nil
		}
		if tasksTitleFlag == "" {
			f.PrintError("--title is required")
			return nil
		}
		if tasksAmountFlag <= 0 {
			f.PrintError("--amount is required and must be positive")
			return nil
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		task, err := client.CreateTask(cmd.Context(), api.CreateTaskParams{
			ContractID:  tasksContractIDFlag,
			Title:       tasksTitleFlag,
			Description: tasksDescriptionFlag,
			Amount:      tasksAmountFlag,
		})
		if err != nil {
			f.PrintError("Failed to create task: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Task created successfully!")
			f.PrintText("ID:     " + task.ID)
			f.PrintText("Title:  " + task.Title)
		}, task)
	},
}

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete <task-id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if !tasksForceFlag {
			f.PrintText(fmt.Sprintf("Are you sure you want to delete task %s?", args[0]))
			f.PrintText("Use --force to confirm.")
			return nil
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		if err := client.DeleteTask(cmd.Context(), args[0]); err != nil {
			f.PrintError("Failed to delete task: %v", err)
			return err
		}

		f.PrintSuccess("Task deleted successfully.")
		return nil
	},
}

var tasksApproveCmd = &cobra.Command{
	Use:   "approve <task-id>",
	Short: "Approve a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		if err := client.ReviewTask(cmd.Context(), args[0], "approved"); err != nil {
			f.PrintError("Failed to approve task: %v", err)
			return err
		}

		f.PrintSuccess("Task approved successfully.")
		return nil
	},
}

var tasksRejectCmd = &cobra.Command{
	Use:   "reject <task-id>",
	Short: "Reject a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		if err := client.ReviewTask(cmd.Context(), args[0], "rejected"); err != nil {
			f.PrintError("Failed to reject task: %v", err)
			return err
		}

		f.PrintSuccess("Task rejected successfully.")
		return nil
	},
}

func init() {
	tasksListCmd.Flags().StringVar(&tasksContractIDFlag, "contract-id", "", "Filter by contract ID")
	tasksListCmd.Flags().StringVar(&tasksStatusFlag, "status", "", "Filter by status")
	tasksListCmd.Flags().IntVar(&tasksLimitFlag, "limit", 50, "Maximum results")

	tasksCreateCmd.Flags().StringVar(&tasksContractIDFlag, "contract-id", "", "Contract ID (required)")
	tasksCreateCmd.Flags().StringVar(&tasksTitleFlag, "title", "", "Task title (required)")
	tasksCreateCmd.Flags().StringVar(&tasksDescriptionFlag, "description", "", "Task description")
	tasksCreateCmd.Flags().Float64Var(&tasksAmountFlag, "amount", 0, "Task amount (required)")

	tasksDeleteCmd.Flags().BoolVar(&tasksForceFlag, "force", false, "Confirm deletion")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksCreateCmd)
	tasksCmd.AddCommand(tasksDeleteCmd)
	tasksCmd.AddCommand(tasksApproveCmd)
	tasksCmd.AddCommand(tasksRejectCmd)
}
