package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var timeOffCmd = &cobra.Command{
	Use:     "time-off",
	Aliases: []string{"timeoff", "pto"},
	Short:   "Manage time off requests",
	Long:    "List, create, and manage time off requests.",
}

var (
	timeOffProfileFlag string
	timeOffStatusFlag  []string
	timeOffLimitFlag   int
	timeOffCursorFlag  string
	timeOffAllFlag     bool
)

var timeOffListCmd = &cobra.Command{
	Use:   "list",
	Short: "List time off requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		cursor := timeOffCursorFlag
		var allRequests []api.TimeOffRequest
		var next string

		for {
			resp, err := client.ListTimeOffRequests(cmd.Context(), api.TimeOffListParams{
				HRISProfileID: timeOffProfileFlag,
				Status:        timeOffStatusFlag,
				Limit:         timeOffLimitFlag,
				Cursor:        cursor,
			})
			if err != nil {
				f.PrintError("Failed to list time off: %v", err)
				return err
			}
			allRequests = append(allRequests, resp.Data...)
			next = resp.Page.Next
			if !timeOffAllFlag || next == "" {
				if !timeOffAllFlag {
					allRequests = resp.Data
				}
				break
			}
			cursor = next
		}

		response := api.TimeOffListResponse{
			Data: allRequests,
		}
		response.Page.Next = ""

		return f.Output(func() {
			if len(allRequests) == 0 {
				f.PrintText("No time off requests found.")
				return
			}
			table := f.NewTable("ID", "WORKER", "TYPE", "DATES", "DAYS", "STATUS")
			for _, t := range allRequests {
				dates := t.StartDate + " - " + t.EndDate
				table.AddRow(t.ID, t.WorkerName, t.Type, dates, fmt.Sprintf("%.1f", t.Days), t.Status)
			}
			table.Render()
			if !timeOffAllFlag && next != "" {
				f.PrintText("")
				f.PrintText("More results available. Use --cursor to paginate or --all to fetch everything.")
			}
		}, response)
	},
}

var timeOffPoliciesCmd = &cobra.Command{
	Use:   "policies",
	Short: "List time off policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		policies, err := client.ListTimeOffPolicies(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list policies: %v", err)
			return err
		}

		return f.Output(func() {
			if len(policies) == 0 {
				f.PrintText("No policies found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE")
			for _, p := range policies {
				table.AddRow(p.ID, p.Name, p.Type)
			}
			table.Render()
		}, policies)
	},
}

var (
	timeOffCreateProfileFlag string
	timeOffCreatePolicyFlag  string
	timeOffCreateStartFlag   string
	timeOffCreateEndFlag     string
	timeOffCreateReasonFlag  string
)

var timeOffCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a time off request",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if timeOffCreateProfileFlag == "" || timeOffCreatePolicyFlag == "" ||
			timeOffCreateStartFlag == "" || timeOffCreateEndFlag == "" {
			f.PrintError("Required: --profile, --policy, --start, --end")
			return fmt.Errorf("missing required flags")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "TimeOffRequest",
			Description: "Create time off request",
			Details: map[string]string{
				"ProfileID": timeOffCreateProfileFlag,
				"PolicyID":  timeOffCreatePolicyFlag,
				"StartDate": timeOffCreateStartFlag,
				"EndDate":   timeOffCreateEndFlag,
				"Reason":    timeOffCreateReasonFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		req, err := client.CreateTimeOffRequest(cmd.Context(), api.CreateTimeOffParams{
			HRISProfileID: timeOffCreateProfileFlag,
			PolicyID:      timeOffCreatePolicyFlag,
			StartDate:     timeOffCreateStartFlag,
			EndDate:       timeOffCreateEndFlag,
			Reason:        timeOffCreateReasonFlag,
		})
		if err != nil {
			f.PrintError("Failed to create request: %v", err)
			return err
		}

		f.PrintSuccess("Created time off request: %s", req.ID)
		return nil
	},
}

var timeOffCancelCmd = &cobra.Command{
	Use:   "cancel <request-id>",
	Short: "Cancel a time off request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CANCEL",
			Resource:    "TimeOffRequest",
			Description: "Cancel time off request",
			Details: map[string]string{
				"ID": args[0],
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		if err := client.CancelTimeOffRequest(cmd.Context(), args[0]); err != nil {
			f.PrintError("Failed to cancel: %v", err)
			return err
		}

		f.PrintSuccess("Cancelled time off request: %s", args[0])
		return nil
	},
}

// Flags for approve command
var timeOffApproveCommentFlag string

var timeOffApproveCmd = &cobra.Command{
	Use:   "approve <request-id>",
	Short: "Approve time off request",
	Long:  "Approve a time off request. Optional --comment flag to add approval notes.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		params := api.ApproveRejectParams{
			RequestID: args[0],
			Action:    "approve",
			Comment:   timeOffApproveCommentFlag,
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "APPROVE",
			Resource:    "TimeOffRequest",
			Description: "Approve time off request",
			Details: map[string]string{
				"ID":      args[0],
				"Comment": timeOffApproveCommentFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		approval, err := client.ApproveRejectTimeOff(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to approve time off request: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Time off request approved successfully")
			f.PrintText("Request ID: " + approval.RequestID)
			f.PrintText("Status:     " + approval.Status)
			if approval.Comment != "" {
				f.PrintText("Comment:    " + approval.Comment)
			}
		}, approval)
	},
}

// Flags for reject command
var timeOffRejectCommentFlag string

var timeOffRejectCmd = &cobra.Command{
	Use:   "reject <request-id>",
	Short: "Reject time off request",
	Long:  "Reject a time off request. Requires --comment flag to provide rejection reason.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if timeOffRejectCommentFlag == "" {
			f.PrintError("--comment flag is required for rejections")
			return fmt.Errorf("--comment flag is required")
		}

		params := api.ApproveRejectParams{
			RequestID: args[0],
			Action:    "reject",
			Comment:   timeOffRejectCommentFlag,
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REJECT",
			Resource:    "TimeOffRequest",
			Description: "Reject time off request",
			Details: map[string]string{
				"ID":      args[0],
				"Comment": timeOffRejectCommentFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		approval, err := client.ApproveRejectTimeOff(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to reject time off request: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Time off request rejected successfully")
			f.PrintText("Request ID: " + approval.RequestID)
			f.PrintText("Status:     " + approval.Status)
			f.PrintText("Comment:    " + approval.Comment)
		}, approval)
	},
}

// Flags for validate command
var (
	timeOffValidateProfileFlag   string
	timeOffValidateTypeFlag      string
	timeOffValidateStartDateFlag string
	timeOffValidateEndDateFlag   string
)

var timeOffValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate time off request",
	Long:  "Validate a time off request without creating it. Requires --profile-id, --type, --start-date, and --end-date flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Validate required flags
		if timeOffValidateProfileFlag == "" {
			f.PrintError("--profile-id flag is required")
			return fmt.Errorf("--profile-id flag is required")
		}
		if timeOffValidateTypeFlag == "" {
			f.PrintError("--type flag is required")
			return fmt.Errorf("--type flag is required")
		}
		if timeOffValidateStartDateFlag == "" {
			f.PrintError("--start-date flag is required")
			return fmt.Errorf("--start-date flag is required")
		}
		if timeOffValidateEndDateFlag == "" {
			f.PrintError("--end-date flag is required")
			return fmt.Errorf("--end-date flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.ValidateTimeOffParams{
			ProfileID: timeOffValidateProfileFlag,
			Type:      timeOffValidateTypeFlag,
			StartDate: timeOffValidateStartDateFlag,
			EndDate:   timeOffValidateEndDateFlag,
		}

		validation, err := client.ValidateTimeOffRequest(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to validate time off request: %v", err)
			return err
		}

		return f.Output(func() {
			if validation.Valid {
				f.PrintSuccess("Time off request is valid")
			} else {
				f.PrintError("Time off request validation failed")
			}
			f.PrintText(fmt.Sprintf("Valid: %t", validation.Valid))
			if len(validation.Errors) > 0 {
				f.PrintText("")
				f.PrintText("Validation Errors:")
				for _, errMsg := range validation.Errors {
					f.PrintText("  - " + errMsg)
				}
			}
			if len(validation.Warnings) > 0 {
				f.PrintText("")
				f.PrintText("Warnings:")
				for _, warn := range validation.Warnings {
					f.PrintText("  - " + warn)
				}
			}
		}, validation)
	},
}

var timeOffEntitlementsCmd = &cobra.Command{
	Use:   "entitlements <profile-id>",
	Short: "Show entitlements for profile",
	Long:  "Display time off entitlements for a specific HRIS profile.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		entitlements, err := client.GetEntitlements(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get entitlements: %v", err)
			return err
		}

		return f.Output(func() {
			if len(entitlements) == 0 {
				f.PrintText("No entitlements found for profile: " + args[0])
				return
			}
			f.PrintText("Time Off Entitlements for Profile: " + args[0])
			f.PrintText("")
			table := f.NewTable("ID", "TYPE", "YEAR", "TOTAL", "USED", "PENDING", "BALANCE")
			for _, ent := range entitlements {
				table.AddRow(
					ent.ID,
					ent.Type,
					fmt.Sprintf("%d", ent.Year),
					fmt.Sprintf("%.1f", ent.TotalDays),
					fmt.Sprintf("%.1f", ent.UsedDays),
					fmt.Sprintf("%.1f", ent.PendingDays),
					fmt.Sprintf("%.1f", ent.Balance),
				)
			}
			table.Render()
		}, entitlements)
	},
}

var timeOffScheduleCmd = &cobra.Command{
	Use:   "schedule <profile-id>",
	Short: "Show work schedule for profile",
	Long:  "Display work schedule for a specific HRIS profile.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		schedule, err := client.GetWorkSchedule(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get work schedule: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Work Schedule for Profile: " + schedule.ProfileID)
			f.PrintText("")
			f.PrintText("Timezone:      " + schedule.Timezone)
			f.PrintText(fmt.Sprintf("Hours Per Day: %.1f", schedule.HoursPerDay))
			f.PrintText("Start Time:    " + schedule.StartTime)
			f.PrintText("End Time:      " + schedule.EndTime)
			f.PrintText("")
			f.PrintText("Working Days:")
			for _, day := range schedule.WorkDays {
				f.PrintText("  - " + day)
			}
		}, schedule)
	},
}

func init() {
	timeOffListCmd.Flags().StringVar(&timeOffProfileFlag, "profile", "", "HRIS profile ID")
	timeOffListCmd.Flags().StringSliceVar(&timeOffStatusFlag, "status", nil, "Filter by status")
	timeOffListCmd.Flags().IntVar(&timeOffLimitFlag, "limit", 50, "Maximum results")
	timeOffListCmd.Flags().StringVar(&timeOffCursorFlag, "cursor", "", "Pagination cursor")
	timeOffListCmd.Flags().BoolVar(&timeOffAllFlag, "all", false, "Fetch all pages")

	timeOffCreateCmd.Flags().StringVar(&timeOffCreateProfileFlag, "profile", "", "HRIS profile ID (required)")
	timeOffCreateCmd.Flags().StringVar(&timeOffCreatePolicyFlag, "policy", "", "Policy ID (required)")
	timeOffCreateCmd.Flags().StringVar(&timeOffCreateStartFlag, "start", "", "Start date YYYY-MM-DD (required)")
	timeOffCreateCmd.Flags().StringVar(&timeOffCreateEndFlag, "end", "", "End date YYYY-MM-DD (required)")
	timeOffCreateCmd.Flags().StringVar(&timeOffCreateReasonFlag, "reason", "", "Reason for time off")

	// Approve command flags
	timeOffApproveCmd.Flags().StringVar(&timeOffApproveCommentFlag, "comment", "", "Optional approval comment")

	// Reject command flags
	timeOffRejectCmd.Flags().StringVar(&timeOffRejectCommentFlag, "comment", "", "Rejection reason (required)")

	// Validate command flags
	timeOffValidateCmd.Flags().StringVar(&timeOffValidateProfileFlag, "profile-id", "", "HRIS profile ID (required)")
	timeOffValidateCmd.Flags().StringVar(&timeOffValidateTypeFlag, "type", "", "Time off type (required)")
	timeOffValidateCmd.Flags().StringVar(&timeOffValidateStartDateFlag, "start-date", "", "Start date YYYY-MM-DD (required)")
	timeOffValidateCmd.Flags().StringVar(&timeOffValidateEndDateFlag, "end-date", "", "End date YYYY-MM-DD (required)")

	timeOffCmd.AddCommand(timeOffListCmd)
	timeOffCmd.AddCommand(timeOffPoliciesCmd)
	timeOffCmd.AddCommand(timeOffCreateCmd)
	timeOffCmd.AddCommand(timeOffCancelCmd)
	timeOffCmd.AddCommand(timeOffApproveCmd)
	timeOffCmd.AddCommand(timeOffRejectCmd)
	timeOffCmd.AddCommand(timeOffValidateCmd)
	timeOffCmd.AddCommand(timeOffEntitlementsCmd)
	timeOffCmd.AddCommand(timeOffScheduleCmd)
}
