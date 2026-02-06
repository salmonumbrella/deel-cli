package cmd

import (
	"context"
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
	tasksContractIDFlag    string
	tasksStatusFlag        string
	tasksLimitFlag         int
	tasksCursorFlag        string
	tasksAllFlag           bool
	tasksTitleFlag         string
	tasksDescriptionFlag   string
	tasksAmountFlag        float64
	tasksForceFlag         bool
	tasksDateSubmittedFlag string
)

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if tasksContractIDFlag == "" {
			return failValidation(cmd, f, "--contract-id is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "listing tasks")
		}

		tasks, page, hasMore, err := collectCursorItems(cmd.Context(), tasksAllFlag, tasksCursorFlag, tasksLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.Task], error) {
			resp, err := client.ListTasks(ctx, api.TasksListParams{
				ContractID: tasksContractIDFlag,
				Status:     tasksStatusFlag,
				Limit:      limit,
				Cursor:     cursor,
			})
			if err != nil {
				return CursorListResult[api.Task]{}, err
			}
			return CursorListResult[api.Task]{
				Items: resp.Data,
				Page: CursorPage{
					Next: resp.Page.Next,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing tasks")
		}

		response := makeListResponse(tasks, page)

		return outputList(cmd, f, tasks, hasMore, "No tasks found.", []string{"ID", "TITLE", "AMOUNT", "STATUS"}, func(t api.Task) []string {
			return []string{t.ID, t.Title, fmt.Sprintf("%.2f %s", t.Amount, t.Currency), t.Status}
		}, response)
	},
}

var tasksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if tasksContractIDFlag == "" {
			return failValidation(cmd, f, "--contract-id is required")
		}
		if tasksTitleFlag == "" {
			return failValidation(cmd, f, "--title is required")
		}
		if tasksAmountFlag <= 0 {
			return failValidation(cmd, f, "--amount is required and must be positive")
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
			return HandleError(f, err, "initializing client")
		}

		task, err := client.CreateTask(cmd.Context(), api.CreateTaskParams{
			ContractID:    tasksContractIDFlag,
			Title:         tasksTitleFlag,
			Description:   tasksDescriptionFlag,
			Amount:        tasksAmountFlag,
			DateSubmitted: tasksDateSubmittedFlag,
		})
		if err != nil {
			return HandleError(f, err, "create task")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Task created successfully!")
			f.PrintText("ID:     " + task.ID)
			f.PrintText("Title:  " + task.Title)
		}, task)
	},
}

// Flags for update command
var (
	tasksUpdateContractIDFlag  string
	tasksUpdateTitleFlag       string
	tasksUpdateDescriptionFlag string
	tasksUpdateAmountFlag      float64
)

var tasksUpdateCmd = &cobra.Command{
	Use:   "update <task-id>",
	Short: "Update a task",
	Long: `Update a task's title, description, or amount.

If --contract-id is not provided, the CLI will search across all active
task-based contracts to find the task (this may take a few seconds).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		taskID := args[0]

		amountSet := cmd.Flags().Changed("amount")
		if tasksUpdateTitleFlag == "" && tasksUpdateDescriptionFlag == "" && !amountSet {
			return failValidation(cmd, f, "At least one of --title, --description, or --amount is required")
		}
		if amountSet && tasksUpdateAmountFlag <= 0 {
			return failValidation(cmd, f, "--amount must be positive")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		contractID := tasksUpdateContractIDFlag
		if contractID == "" {
			f.PrintText(fmt.Sprintf("Looking up contract for task %s...", taskID))
			foundContractID, _, err := client.FindTaskContract(cmd.Context(), taskID)
			if err != nil {
				f.PrintText("\nTo find the contract ID:")
				f.PrintText("  deel contracts list --json --items | jq '.[] | select(.type | contains(\"payg\")) | {id, title, worker_name}'")
				return HandleError(f, err, "finding task contract")
			}
			contractID = foundContractID
			f.PrintText(fmt.Sprintf("Found task in contract: %s", contractID))
		}

		details := map[string]string{
			"ContractID": contractID,
			"ID":         taskID,
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

		params := api.UpdateTaskParams{
			Title:       tasksUpdateTitleFlag,
			Description: tasksUpdateDescriptionFlag,
			Amount:      tasksUpdateAmountFlag,
		}

		task, err := client.UpdateTask(cmd.Context(), contractID, taskID, params)
		if err != nil {
			return HandleError(f, err, "updating task")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Task updated successfully")
			f.PrintText("ID:     " + task.ID)
			f.PrintText("Title:  " + task.Title)
			f.PrintText(fmt.Sprintf("Amount: %.2f %s", task.Amount, task.Currency))
			f.PrintText("Status: " + task.Status)
		}, task)
	},
}

// Flags for get command
var tasksGetContractIDFlag string

var tasksGetCmd = &cobra.Command{
	Use:   "get <task-id>",
	Short: "Get a task by ID",
	Long: `Get a task by ID.

If --contract-id is not provided, the CLI will search across all active
task-based contracts to find the task (this may take a few seconds).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		taskID := args[0]

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		contractID := tasksGetContractIDFlag
		if contractID == "" {
			f.PrintText(fmt.Sprintf("Looking up contract for task %s...", taskID))
			foundContractID, _, err := client.FindTaskContract(cmd.Context(), taskID)
			if err != nil {
				return HandleError(f, err, "finding task contract")
			}
			contractID = foundContractID
			f.PrintText(fmt.Sprintf("Found task in contract: %s", contractID))
		}

		task, err := client.GetTask(cmd.Context(), contractID, taskID)
		if err != nil {
			return HandleError(f, err, "getting task")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("ID:       " + task.ID)
			f.PrintText("Contract: " + contractID)
			f.PrintText("Title:    " + task.Title)
			if task.Description != "" {
				f.PrintText("Desc:     " + task.Description)
			}
			f.PrintText(fmt.Sprintf("Amount:   %.2f %s", task.Amount, task.Currency))
			f.PrintText("Status:   " + task.Status)
		}, task)
	},
}

// Flags for review-many command
var (
	tasksReviewManyContractIDFlag string
	tasksReviewManyStatusFlag     string
	tasksReviewManyIDsFlag        []string
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
			return failValidation(cmd, f, "--status must be approve or reject")
		}

		if len(tasksReviewManyIDsFlag) == 0 {
			return failValidation(cmd, f, "--ids is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REVIEW",
			Resource:    "Task",
			Description: "Review multiple tasks",
			Details: map[string]string{
				"ContractID": tasksReviewManyContractIDFlag,
				"Status":     status,
				"IDs":        strings.Join(tasksReviewManyIDsFlag, ","),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		// If contract ID isn't provided, resolve it per task and group by contract.
		contractToTasks := map[string][]string{}
		if tasksReviewManyContractIDFlag != "" {
			contractToTasks[tasksReviewManyContractIDFlag] = append(contractToTasks[tasksReviewManyContractIDFlag], tasksReviewManyIDsFlag...)
		} else {
			f.PrintText("Looking up contracts for tasks...")
			contractByTask, _, err := client.FindTasksContracts(cmd.Context(), tasksReviewManyIDsFlag)
			if err != nil {
				f.PrintText("Hint: Provide --contract-id to avoid this lookup and reduce API calls.")
				return HandleError(f, err, "finding task contracts")
			}
			for _, taskID := range tasksReviewManyIDsFlag {
				contractID := contractByTask[taskID]
				contractToTasks[contractID] = append(contractToTasks[contractID], taskID)
			}
		}

		for contractID, taskIDs := range contractToTasks {
			if err := client.ReviewMultipleTasks(cmd.Context(), contractID, taskIDs, status); err != nil {
				return HandleError(f, err, "review tasks")
			}
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Tasks %s successfully.", status)
		}, map[string]any{
			"operation": "REVIEW",
			"resource":  "Task",
			"status":    status,
			"tasks":     contractToTasks,
		})
	},
}

// Flags for delete command
var tasksDeleteContractIDFlag string

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete <task-id>",
	Short: "Delete a task",
	Long: `Delete a task.

If --contract-id is not provided, the CLI will search across all active
task-based contracts to find the task (this may take a few seconds).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		taskID := args[0]

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		contractID := tasksDeleteContractIDFlag
		if contractID == "" {
			f.PrintText(fmt.Sprintf("Looking up contract for task %s...", taskID))
			foundContractID, _, err := client.FindTaskContract(cmd.Context(), taskID)
			if err != nil {
				f.PrintText("\nTo find the contract ID:")
				f.PrintText("  deel contracts list --json --items | jq '.[] | select(.type | contains(\"payg\")) | {id, title, worker_name}'")
				return HandleError(f, err, "finding task contract")
			}
			contractID = foundContractID
			f.PrintText(fmt.Sprintf("Found task in contract: %s", contractID))
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "Task",
			Description: "Delete task",
			Details: map[string]string{
				"ContractID": contractID,
				"ID":         taskID,
			},
		}); ok {
			return err
		}

		suggested := "deel tasks delete " + taskID + " --force"
		if contractID != "" {
			suggested = "deel tasks delete " + taskID + " --contract-id " + contractID + " --force"
		}
		if ok, err := requireForce(cmd, f, tasksForceFlag, "delete", "task", taskID, suggested); !ok {
			return err
		}

		if err := client.DeleteTask(cmd.Context(), contractID, taskID); err != nil {
			return HandleError(f, err, "deleting task")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Task deleted successfully.")
		}, map[string]any{
			"deleted":     true,
			"task_id":     taskID,
			"contract_id": contractID,
		})
	},
}

// Flags for approve/reject commands
var (
	tasksApproveContractIDFlag string
	tasksRejectContractIDFlag  string
)

var tasksApproveCmd = &cobra.Command{
	Use:   "approve <task-id>",
	Short: "Approve a task",
	Long: `Approve a task for payment.

If --contract-id is not provided, the CLI will search across all active
task-based contracts to find the task (this may take a few seconds).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		taskID := args[0]

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		contractID := tasksApproveContractIDFlag
		if contractID == "" {
			f.PrintText(fmt.Sprintf("Looking up contract for task %s...", taskID))
			foundContractID, _, err := client.FindTaskContract(cmd.Context(), taskID)
			if err != nil {
				f.PrintText("\nTo find the contract ID:")
				f.PrintText("  deel contracts list --json --items | jq '.[] | select(.type | contains(\"payg\")) | {id, title, worker_name}'")
				return HandleError(f, err, "finding task contract")
			}
			contractID = foundContractID
			f.PrintText(fmt.Sprintf("Found task in contract: %s", contractID))
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REVIEW",
			Resource:    "Task",
			Description: "Approve task",
			Details: map[string]string{
				"ContractID": contractID,
				"ID":         taskID,
				"Status":     "approved",
			},
		}); ok {
			return err
		}

		if err := client.ReviewTask(cmd.Context(), contractID, taskID, "approved"); err != nil {
			return HandleError(f, err, "approving task")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Task approved successfully.")
		}, map[string]any{
			"operation":   "REVIEW",
			"resource":    "Task",
			"status":      "approved",
			"task_id":     taskID,
			"contract_id": contractID,
		})
	},
}

var tasksRejectCmd = &cobra.Command{
	Use:   "reject <task-id>",
	Short: "Reject a task",
	Long: `Reject a task.

If --contract-id is not provided, the CLI will search across all active
task-based contracts to find the task (this may take a few seconds).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		taskID := args[0]

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		contractID := tasksRejectContractIDFlag
		if contractID == "" {
			f.PrintText(fmt.Sprintf("Looking up contract for task %s...", taskID))
			foundContractID, _, err := client.FindTaskContract(cmd.Context(), taskID)
			if err != nil {
				f.PrintText("\nTo find the contract ID:")
				f.PrintText("  deel contracts list --json --items | jq '.[] | select(.type | contains(\"payg\")) | {id, title, worker_name}'")
				return HandleError(f, err, "finding task contract")
			}
			contractID = foundContractID
			f.PrintText(fmt.Sprintf("Found task in contract: %s", contractID))
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REVIEW",
			Resource:    "Task",
			Description: "Reject task",
			Details: map[string]string{
				"ContractID": contractID,
				"ID":         taskID,
				"Status":     "rejected",
			},
		}); ok {
			return err
		}

		if err := client.ReviewTask(cmd.Context(), contractID, taskID, "rejected"); err != nil {
			return HandleError(f, err, "rejecting task")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Task rejected successfully.")
		}, map[string]any{
			"operation":   "REVIEW",
			"resource":    "Task",
			"status":      "rejected",
			"task_id":     taskID,
			"contract_id": contractID,
		})
	},
}

func init() {
	// List command flags
	tasksListCmd.Flags().StringVar(&tasksContractIDFlag, "contract-id", "", "Contract ID (required)")
	tasksListCmd.Flags().StringVar(&tasksStatusFlag, "status", "", "Filter by status")
	tasksListCmd.Flags().IntVar(&tasksLimitFlag, "limit", 100, "Maximum results")
	tasksListCmd.Flags().StringVar(&tasksCursorFlag, "cursor", "", "Pagination cursor")
	tasksListCmd.Flags().BoolVar(&tasksAllFlag, "all", false, "Fetch all pages")

	// Create command flags
	tasksCreateCmd.Flags().StringVar(&tasksContractIDFlag, "contract-id", "", "Contract ID (required)")
	tasksCreateCmd.Flags().StringVar(&tasksTitleFlag, "title", "", "Task title (required)")
	tasksCreateCmd.Flags().StringVar(&tasksDescriptionFlag, "description", "", "Task description")
	tasksCreateCmd.Flags().Float64Var(&tasksAmountFlag, "amount", 0, "Task amount (required)")
	tasksCreateCmd.Flags().StringVar(&tasksDateSubmittedFlag, "date-submitted", "", "Submission date (YYYY-MM-DD, defaults to today)")

	// Update command flags
	tasksUpdateCmd.Flags().StringVar(&tasksUpdateContractIDFlag, "contract-id", "", "Contract ID (required)")
	tasksUpdateCmd.Flags().StringVar(&tasksUpdateTitleFlag, "title", "", "Task title")
	tasksUpdateCmd.Flags().StringVar(&tasksUpdateDescriptionFlag, "description", "", "Task description")
	tasksUpdateCmd.Flags().Float64Var(&tasksUpdateAmountFlag, "amount", 0, "Task amount")

	// Get command flags
	tasksGetCmd.Flags().StringVar(&tasksGetContractIDFlag, "contract-id", "", "Contract ID (optional)")

	// Review-many command flags
	tasksReviewManyCmd.Flags().StringVar(&tasksReviewManyContractIDFlag, "contract-id", "", "Contract ID (optional)")
	tasksReviewManyCmd.Flags().StringVar(&tasksReviewManyStatusFlag, "status", "", "approve or reject")
	tasksReviewManyCmd.Flags().StringSliceVar(&tasksReviewManyIDsFlag, "ids", nil, "Task IDs (comma-separated or repeat)")

	// Delete command flags
	tasksDeleteCmd.Flags().StringVar(&tasksDeleteContractIDFlag, "contract-id", "", "Contract ID (optional)")
	tasksDeleteCmd.Flags().BoolVar(&tasksForceFlag, "force", false, "Confirm deletion")

	// Approve command flags
	tasksApproveCmd.Flags().StringVar(&tasksApproveContractIDFlag, "contract-id", "", "Contract ID (optional)")

	// Reject command flags
	tasksRejectCmd.Flags().StringVar(&tasksRejectContractIDFlag, "contract-id", "", "Contract ID (optional)")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksCreateCmd)
	tasksCmd.AddCommand(tasksGetCmd)
	tasksCmd.AddCommand(tasksUpdateCmd)
	tasksCmd.AddCommand(tasksDeleteCmd)
	tasksCmd.AddCommand(tasksApproveCmd)
	tasksCmd.AddCommand(tasksRejectCmd)
	tasksCmd.AddCommand(tasksReviewManyCmd)
}
