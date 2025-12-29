package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
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
	tasksCursorFlag      string
	tasksAllFlag         bool
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

		cursor := tasksCursorFlag
		var allTasks []api.Task
		var next string

		for {
			resp, err := client.ListTasks(cmd.Context(), api.TasksListParams{
				ContractID: tasksContractIDFlag,
				Status:     tasksStatusFlag,
				Limit:      tasksLimitFlag,
				Cursor:     cursor,
			})
			if err != nil {
				f.PrintError("Failed to list tasks: %v", err)
				return err
			}
			allTasks = append(allTasks, resp.Data...)
			next = resp.Page.Next
			if !tasksAllFlag || next == "" {
				if !tasksAllFlag {
					allTasks = resp.Data
				}
				break
			}
			cursor = next
		}

		response := api.TasksListResponse{
			Data: allTasks,
		}
		response.Page.Next = ""

		return f.Output(func() {
			if len(allTasks) == 0 {
				f.PrintText("No tasks found.")
				return
			}
			table := f.NewTable("ID", "TITLE", "AMOUNT", "STATUS")
			for _, t := range allTasks {
				table.AddRow(t.ID, t.Title, fmt.Sprintf("%.2f %s", t.Amount, t.Currency), t.Status)
			}
			table.Render()
			if !tasksAllFlag && next != "" {
				f.PrintText("")
				f.PrintText("More results available. Use --cursor to paginate or --all to fetch everything.")
			}
		}, response)
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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Task",
			Description: "Create task",
			Details: map[string]string{
				"ContractID":  tasksContractIDFlag,
				"Title":       tasksTitleFlag,
				"Description": tasksDescriptionFlag,
				"Amount":      fmt.Sprintf("%.2f", tasksAmountFlag),
			},
		}); ok {
			return err
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

// Flags for update command
var (
	tasksUpdateTitleFlag       string
	tasksUpdateDescriptionFlag string
	tasksUpdateAmountFlag      float64
)

var tasksUpdateCmd = &cobra.Command{
	Use:   "update <task-id>",
	Short: "Update a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		amountSet := cmd.Flags().Changed("amount")
		if tasksUpdateTitleFlag == "" && tasksUpdateDescriptionFlag == "" && !amountSet {
			f.PrintError("At least one of --title, --description, or --amount is required")
			return nil
		}
		if amountSet && tasksUpdateAmountFlag <= 0 {
			f.PrintError("--amount must be positive")
			return nil
		}

		details := map[string]string{
			"ID": args[0],
		}
		if tasksUpdateTitleFlag != "" {
			details["Title"] = tasksUpdateTitleFlag
		}
		if tasksUpdateDescriptionFlag != "" {
			details["Description"] = tasksUpdateDescriptionFlag
		}
		if amountSet {
			details["Amount"] = fmt.Sprintf("%.2f", tasksUpdateAmountFlag)
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Task",
			Description: "Update task",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.UpdateTaskParams{
			Title:       tasksUpdateTitleFlag,
			Description: tasksUpdateDescriptionFlag,
			Amount:      tasksUpdateAmountFlag,
		}

		task, err := client.UpdateTask(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to update task: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Task updated successfully")
			f.PrintText("ID:     " + task.ID)
			f.PrintText("Title:  " + task.Title)
			f.PrintText(fmt.Sprintf("Amount: %.2f %s", task.Amount, task.Currency))
			f.PrintText("Status: " + task.Status)
		}, task)
	},
}

// Flags for review-many command
var (
	tasksReviewManyStatusFlag string
	tasksReviewManyIDsFlag    []string
)

var tasksReviewManyCmd = &cobra.Command{
	Use:   "review-many",
	Short: "Approve or reject multiple tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		status := strings.ToLower(tasksReviewManyStatusFlag)
		switch status {
		case "approve", "approved":
			status = "approved"
		case "reject", "rejected":
			status = "rejected"
		default:
			f.PrintError("--status must be approve or reject")
			return nil
		}

		if len(tasksReviewManyIDsFlag) == 0 {
			f.PrintError("--ids is required")
			return nil
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REVIEW",
			Resource:    "Task",
			Description: "Review multiple tasks",
			Details: map[string]string{
				"Status": status,
				"IDs":    strings.Join(tasksReviewManyIDsFlag, ","),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		if err := client.ReviewMultipleTasks(cmd.Context(), tasksReviewManyIDsFlag, status); err != nil {
			f.PrintError("Failed to review tasks: %v", err)
			return err
		}

		f.PrintSuccess("Tasks %s successfully.", status)
		return nil
	},
}

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete <task-id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "Task",
			Description: "Delete task",
			Details: map[string]string{
				"ID": args[0],
			},
		}); ok {
			return err
		}

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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REVIEW",
			Resource:    "Task",
			Description: "Approve task",
			Details: map[string]string{
				"ID":     args[0],
				"Status": "approved",
			},
		}); ok {
			return err
		}

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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REVIEW",
			Resource:    "Task",
			Description: "Reject task",
			Details: map[string]string{
				"ID":     args[0],
				"Status": "rejected",
			},
		}); ok {
			return err
		}

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
	tasksListCmd.Flags().StringVar(&tasksCursorFlag, "cursor", "", "Pagination cursor")
	tasksListCmd.Flags().BoolVar(&tasksAllFlag, "all", false, "Fetch all pages")

	tasksCreateCmd.Flags().StringVar(&tasksContractIDFlag, "contract-id", "", "Contract ID (required)")
	tasksCreateCmd.Flags().StringVar(&tasksTitleFlag, "title", "", "Task title (required)")
	tasksCreateCmd.Flags().StringVar(&tasksDescriptionFlag, "description", "", "Task description")
	tasksCreateCmd.Flags().Float64Var(&tasksAmountFlag, "amount", 0, "Task amount (required)")

	tasksUpdateCmd.Flags().StringVar(&tasksUpdateTitleFlag, "title", "", "Task title")
	tasksUpdateCmd.Flags().StringVar(&tasksUpdateDescriptionFlag, "description", "", "Task description")
	tasksUpdateCmd.Flags().Float64Var(&tasksUpdateAmountFlag, "amount", 0, "Task amount")

	tasksReviewManyCmd.Flags().StringVar(&tasksReviewManyStatusFlag, "status", "", "approve or reject")
	tasksReviewManyCmd.Flags().StringSliceVar(&tasksReviewManyIDsFlag, "ids", nil, "Task IDs (comma-separated or repeat)")

	tasksDeleteCmd.Flags().BoolVar(&tasksForceFlag, "force", false, "Confirm deletion")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksCreateCmd)
	tasksCmd.AddCommand(tasksUpdateCmd)
	tasksCmd.AddCommand(tasksDeleteCmd)
	tasksCmd.AddCommand(tasksApproveCmd)
	tasksCmd.AddCommand(tasksRejectCmd)
	tasksCmd.AddCommand(tasksReviewManyCmd)
}
