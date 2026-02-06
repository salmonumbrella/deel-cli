package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var timesheetsCmd = &cobra.Command{
	Use:   "timesheets",
	Short: "Manage timesheets",
	Long: `Manage timesheets for hourly contractors to track hours worked.

Use this for:
  - Viewing/approving hours logged by hourly contractors
  - Creating/updating timesheet entries
  - Managing hourly presets

Note: For expenses, bonuses, or deductions, use 'deel invoices adjustments' instead.`,
}

// Flags for list command
var (
	timesheetsListContractIDFlag string
	timesheetsListStatusFlag     string
	timesheetsListLimitFlag      int
	timesheetsListCursorFlag     string
	timesheetsListAllFlag        bool
)

var timesheetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List timesheets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, f, err := initClient("listing timesheets")
		if err != nil {
			return err
		}

		timesheets, page, hasMore, err := collectCursorItems(cmd.Context(), timesheetsListAllFlag, timesheetsListCursorFlag, timesheetsListLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.Timesheet], error) {
			params := api.TimesheetsListParams{
				ContractID: timesheetsListContractIDFlag,
				Status:     timesheetsListStatusFlag,
				Limit:      limit,
				Cursor:     cursor,
			}

			resp, err := client.ListTimesheets(ctx, params)
			if err != nil {
				return CursorListResult[api.Timesheet]{}, err
			}

			return CursorListResult[api.Timesheet]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing timesheets")
		}

		response := makeListResponse(timesheets, page)

		return outputList(cmd, f, timesheets, hasMore, "No timesheets found.", []string{"ID", "CONTRACT ID", "STATUS", "PERIOD", "TOTAL HOURS", "CREATED"}, func(ts api.Timesheet) []string {
			period := fmt.Sprintf("%s to %s", ts.PeriodStart, ts.PeriodEnd)
			return []string{ts.ID, ts.ContractID, ts.Status, period, fmt.Sprintf("%.2f", ts.TotalHours), ts.CreatedAt}
		}, response)
	},
}

var timesheetsGetCmd = &cobra.Command{
	Use:   "get <timesheet-id>",
	Short: "Get timesheet details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		timesheet, err := client.GetTimesheet(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "get timesheet")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("ID:          " + timesheet.ID)
			f.PrintText("Contract ID: " + timesheet.ContractID)
			f.PrintText("Status:      " + timesheet.Status)
			f.PrintText("Period:      " + timesheet.PeriodStart + " to " + timesheet.PeriodEnd)
			f.PrintText(fmt.Sprintf("Total Hours: %.2f", timesheet.TotalHours))
			f.PrintText("Created:     " + timesheet.CreatedAt)
			if len(timesheet.Entries) > 0 {
				f.PrintText("")
				f.PrintText("Entries:")
				table := f.NewTable("  ID", "DATE", "HOURS", "DESCRIPTION")
				for _, entry := range timesheet.Entries {
					table.AddRow("  "+entry.ID, entry.Date, fmt.Sprintf("%.2f", entry.Hours), entry.Description)
				}
				table.Render()
			}
		}, timesheet)
	},
}

// Flags for create-entry command
var (
	createEntryTimesheetIDFlag string
	createEntryDateFlag        string
	createEntryHoursFlag       float64
	createEntryDescriptionFlag string
)

var timesheetsCreateEntryCmd = &cobra.Command{
	Use:   "create-entry",
	Short: "Create a timesheet entry",
	Long:  "Create a new timesheet entry. Requires --timesheet-id, --date, and --hours flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if createEntryTimesheetIDFlag == "" {
			return failValidation(cmd, f, "--timesheet-id flag is required")
		}
		if createEntryDateFlag == "" {
			return failValidation(cmd, f, "--date flag is required")
		}
		if createEntryHoursFlag <= 0 {
			return failValidation(cmd, f, "--hours flag is required and must be greater than 0")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "TimesheetEntry",
			Description: "Create timesheet entry",
			Details: map[string]string{
				"TimesheetID": createEntryTimesheetIDFlag,
				"Date":        createEntryDateFlag,
				"Hours":       fmt.Sprintf("%.2f", createEntryHoursFlag),
				"Description": createEntryDescriptionFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.CreateTimesheetEntryParams{
			TimesheetID: createEntryTimesheetIDFlag,
			Date:        createEntryDateFlag,
			Hours:       createEntryHoursFlag,
			Description: createEntryDescriptionFlag,
		}

		entry, err := client.CreateTimesheetEntry(cmd.Context(), params)
		if err != nil {
			return HandleError(f, err, "create timesheet entry")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Timesheet entry created successfully")
			f.PrintText("ID:          " + entry.ID)
			f.PrintText("Timesheet:   " + entry.TimesheetID)
			f.PrintText("Date:        " + entry.Date)
			f.PrintText(fmt.Sprintf("Hours:       %.2f", entry.Hours))
			if entry.Description != "" {
				f.PrintText("Description: " + entry.Description)
			}
		}, entry)
	},
}

// Flags for update-entry command
var (
	updateEntryHoursFlag       float64
	updateEntryDescriptionFlag string
)

var timesheetsUpdateEntryCmd = &cobra.Command{
	Use:   "update-entry <entry-id>",
	Short: "Update a timesheet entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		hoursSet := cmd.Flags().Changed("hours")
		if !hoursSet && updateEntryDescriptionFlag == "" {
			return failValidation(cmd, f, "At least one of --hours or --description is required")
		}
		if hoursSet && updateEntryHoursFlag <= 0 {
			return failValidation(cmd, f, "--hours must be greater than 0")
		}

		details := map[string]string{
			"ID": args[0],
		}
		if hoursSet {
			details["Hours"] = fmt.Sprintf("%.2f", updateEntryHoursFlag)
		}
		if updateEntryDescriptionFlag != "" {
			details["Description"] = updateEntryDescriptionFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "TimesheetEntry",
			Description: "Update timesheet entry",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.UpdateTimesheetEntryParams{
			Hours:       updateEntryHoursFlag,
			Description: updateEntryDescriptionFlag,
		}

		entry, err := client.UpdateTimesheetEntry(cmd.Context(), args[0], params)
		if err != nil {
			return HandleError(f, err, "update timesheet entry")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Timesheet entry updated successfully")
			f.PrintText("ID:          " + entry.ID)
			f.PrintText("Timesheet:   " + entry.TimesheetID)
			f.PrintText("Date:        " + entry.Date)
			f.PrintText(fmt.Sprintf("Hours:       %.2f", entry.Hours))
			if entry.Description != "" {
				f.PrintText("Description: " + entry.Description)
			}
		}, entry)
	},
}

// Flags for delete-entry command
var timesheetsDeleteEntryForceFlag bool

var timesheetsDeleteEntryCmd = &cobra.Command{
	Use:   "delete-entry <entry-id>",
	Short: "Delete a timesheet entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "TimesheetEntry",
			Description: "Delete timesheet entry",
			Details: map[string]string{
				"ID": args[0],
			},
		}); ok {
			return err
		}

		if ok, err := requireForce(cmd, f, timesheetsDeleteEntryForceFlag, "delete", "timesheet entry", args[0], "deel timesheets delete-entry "+args[0]+" --force"); !ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		if err := client.DeleteTimesheetEntry(cmd.Context(), args[0]); err != nil {
			return HandleError(f, err, "delete timesheet entry")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Timesheet entry deleted successfully.")
		}, map[string]any{
			"deleted":   true,
			"entry_id":  args[0],
			"resource":  "TimesheetEntry",
			"operation": "DELETE",
		})
	},
}

// Flags for review command
var reviewCommentFlag string

var timesheetsReviewCmd = &cobra.Command{
	Use:   "review <timesheet-id> <approve|reject>",
	Short: "Review a timesheet",
	Long:  "Approve or reject a timesheet. Optionally provide a comment with --comment flag.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		timesheetID := args[0]
		action := args[1]

		// Validate action
		var status string
		switch action {
		case "approve":
			status = "approved"
		case "reject":
			status = "rejected"
		default:
			return failValidation(cmd, f, fmt.Sprintf("invalid action %q (must be 'approve' or 'reject')", action))
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "REVIEW",
			Resource:    "Timesheet",
			Description: "Review timesheet",
			Details: map[string]string{
				"ID":      timesheetID,
				"Status":  status,
				"Comment": reviewCommentFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.ReviewTimesheetParams{
			Status:  status,
			Comment: reviewCommentFlag,
		}

		timesheet, err := client.ReviewTimesheet(cmd.Context(), timesheetID, params)
		if err != nil {
			return HandleError(f, err, "review timesheet")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Timesheet %s successfully", action+"d")
			f.PrintText("ID:          " + timesheet.ID)
			f.PrintText("Contract ID: " + timesheet.ContractID)
			f.PrintText("Status:      " + timesheet.Status)
			f.PrintText("Period:      " + timesheet.PeriodStart + " to " + timesheet.PeriodEnd)
			f.PrintText(fmt.Sprintf("Total Hours: %.2f", timesheet.TotalHours))
		}, timesheet)
	},
}

// Presets subcommand
var presetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "Manage hourly presets",
	Long:  "List, create, and delete hourly presets for timesheets.",
}

var presetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List hourly presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		presets, err := client.ListHourlyPresets(cmd.Context())
		if err != nil {
			return HandleError(f, err, "list hourly presets")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(presets) == 0 {
				f.PrintText("No hourly presets found.")
				return
			}
			table := f.NewTable("ID", "NAME", "HOURS/DAY", "HOURS/WEEK", "CREATED")
			for _, p := range presets {
				table.AddRow(p.ID, p.Name, fmt.Sprintf("%.2f", p.HoursPerDay), fmt.Sprintf("%.2f", p.HoursPerWeek), p.CreatedAt)
			}
			table.Render()
		}, presets)
	},
}

// Flags for presets create command
var (
	presetsCreateNameFlag         string
	presetsCreateHoursPerDayFlag  string
	presetsCreateHoursPerWeekFlag string
)

var presetsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create hourly preset",
	Long:  "Create a new hourly preset. Requires --name, --hours-per-day, and --hours-per-week flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if presetsCreateNameFlag == "" {
			return failValidation(cmd, f, "--name flag is required")
		}
		if presetsCreateHoursPerDayFlag == "" {
			return failValidation(cmd, f, "--hours-per-day flag is required")
		}
		if presetsCreateHoursPerWeekFlag == "" {
			return failValidation(cmd, f, "--hours-per-week flag is required")
		}

		// Parse hours
		hoursPerDay, err := strconv.ParseFloat(presetsCreateHoursPerDayFlag, 64)
		if err != nil {
			return failValidation(cmd, f, fmt.Sprintf("invalid --hours-per-day value: %v", err))
		}

		hoursPerWeek, err := strconv.ParseFloat(presetsCreateHoursPerWeekFlag, 64)
		if err != nil {
			return failValidation(cmd, f, fmt.Sprintf("invalid --hours-per-week value: %v", err))
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "HourlyPreset",
			Description: "Create hourly preset",
			Details: map[string]string{
				"Name":         presetsCreateNameFlag,
				"HoursPerDay":  fmt.Sprintf("%.2f", hoursPerDay),
				"HoursPerWeek": fmt.Sprintf("%.2f", hoursPerWeek),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.CreateHourlyPresetParams{
			Name:         presetsCreateNameFlag,
			HoursPerDay:  hoursPerDay,
			HoursPerWeek: hoursPerWeek,
		}

		preset, err := client.CreateHourlyPreset(cmd.Context(), params)
		if err != nil {
			return HandleError(f, err, "create hourly preset")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Hourly preset created successfully")
			f.PrintText("ID:            " + preset.ID)
			f.PrintText("Name:          " + preset.Name)
			f.PrintText(fmt.Sprintf("Hours per day: %.2f", preset.HoursPerDay))
			f.PrintText(fmt.Sprintf("Hours per week: %.2f", preset.HoursPerWeek))
			f.PrintText("Created:       " + preset.CreatedAt)
		}, preset)
	},
}

// Flags for presets update command
var (
	presetsUpdateNameFlag         string
	presetsUpdateDescriptionFlag  string
	presetsUpdateHoursPerDayFlag  string
	presetsUpdateHoursPerWeekFlag string
	presetsUpdateRateFlag         string
	presetsUpdateCurrencyFlag     string
)

var presetsUpdateCmd = &cobra.Command{
	Use:   "update <preset-id>",
	Short: "Update hourly preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		var hoursPerDay float64
		var hoursPerWeek float64
		var rate float64

		if presetsUpdateHoursPerDayFlag != "" {
			parsed, err := strconv.ParseFloat(presetsUpdateHoursPerDayFlag, 64)
			if err != nil {
				return failValidation(cmd, f, fmt.Sprintf("invalid --hours-per-day value: %v", err))
			}
			hoursPerDay = parsed
		}
		if presetsUpdateHoursPerWeekFlag != "" {
			parsed, err := strconv.ParseFloat(presetsUpdateHoursPerWeekFlag, 64)
			if err != nil {
				return failValidation(cmd, f, fmt.Sprintf("invalid --hours-per-week value: %v", err))
			}
			hoursPerWeek = parsed
		}
		if presetsUpdateRateFlag != "" {
			parsed, err := strconv.ParseFloat(presetsUpdateRateFlag, 64)
			if err != nil {
				return failValidation(cmd, f, fmt.Sprintf("invalid --rate value: %v", err))
			}
			rate = parsed
		}

		if presetsUpdateNameFlag == "" &&
			presetsUpdateDescriptionFlag == "" &&
			presetsUpdateHoursPerDayFlag == "" &&
			presetsUpdateHoursPerWeekFlag == "" &&
			presetsUpdateRateFlag == "" &&
			presetsUpdateCurrencyFlag == "" {
			return failValidation(cmd, f, "at least one update field is required")
		}

		details := map[string]string{
			"ID": args[0],
		}
		if presetsUpdateNameFlag != "" {
			details["Name"] = presetsUpdateNameFlag
		}
		if presetsUpdateDescriptionFlag != "" {
			details["Description"] = presetsUpdateDescriptionFlag
		}
		if presetsUpdateHoursPerDayFlag != "" {
			details["HoursPerDay"] = presetsUpdateHoursPerDayFlag
		}
		if presetsUpdateHoursPerWeekFlag != "" {
			details["HoursPerWeek"] = presetsUpdateHoursPerWeekFlag
		}
		if presetsUpdateRateFlag != "" {
			details["Rate"] = presetsUpdateRateFlag
		}
		if presetsUpdateCurrencyFlag != "" {
			details["Currency"] = presetsUpdateCurrencyFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "HourlyPreset",
			Description: "Update hourly preset",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.UpdateHourlyPresetParams{
			Name:         presetsUpdateNameFlag,
			Description:  presetsUpdateDescriptionFlag,
			HoursPerDay:  hoursPerDay,
			HoursPerWeek: hoursPerWeek,
			Rate:         rate,
			Currency:     presetsUpdateCurrencyFlag,
		}

		preset, err := client.UpdateHourlyPreset(cmd.Context(), args[0], params)
		if err != nil {
			return HandleError(f, err, "update hourly preset")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Hourly preset updated successfully")
			f.PrintText("ID:            " + preset.ID)
			f.PrintText("Name:          " + preset.Name)
			f.PrintText(fmt.Sprintf("Hours per day: %.2f", preset.HoursPerDay))
			f.PrintText(fmt.Sprintf("Hours per week: %.2f", preset.HoursPerWeek))
			if preset.Currency != "" {
				f.PrintText(fmt.Sprintf("Rate:          %.2f %s", preset.Rate, preset.Currency))
			}
		}, preset)
	},
}

var presetsDeleteCmd = &cobra.Command{
	Use:   "delete <preset-id>",
	Short: "Delete hourly preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "HourlyPreset",
			Description: "Delete hourly preset",
			Details: map[string]string{
				"ID": args[0],
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		err = client.DeleteHourlyPreset(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "delete hourly preset")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Hourly preset deleted successfully")
		}, map[string]string{"status": "deleted", "id": args[0]})
	},
}

func init() {
	// List command flags
	timesheetsListCmd.Flags().StringVar(&timesheetsListContractIDFlag, "contract-id", "", "Filter by contract ID")
	timesheetsListCmd.Flags().StringVar(&timesheetsListStatusFlag, "status", "", "Filter by status (e.g., pending, approved, rejected)")
	timesheetsListCmd.Flags().IntVar(&timesheetsListLimitFlag, "limit", 100, "Maximum results")
	timesheetsListCmd.Flags().StringVar(&timesheetsListCursorFlag, "cursor", "", "Pagination cursor")
	timesheetsListCmd.Flags().BoolVar(&timesheetsListAllFlag, "all", false, "Fetch all pages")

	// Create entry command flags
	timesheetsCreateEntryCmd.Flags().StringVar(&createEntryTimesheetIDFlag, "timesheet-id", "", "Timesheet ID (required)")
	timesheetsCreateEntryCmd.Flags().StringVar(&createEntryDateFlag, "date", "", "Date for the entry (YYYY-MM-DD format, required)")
	timesheetsCreateEntryCmd.Flags().Float64Var(&createEntryHoursFlag, "hours", 0, "Hours worked (required)")
	timesheetsCreateEntryCmd.Flags().StringVar(&createEntryDescriptionFlag, "description", "", "Description of work (optional)")

	// Update entry command flags
	timesheetsUpdateEntryCmd.Flags().Float64Var(&updateEntryHoursFlag, "hours", 0, "Hours worked")
	timesheetsUpdateEntryCmd.Flags().StringVar(&updateEntryDescriptionFlag, "description", "", "Description of work")

	// Delete entry command flags
	timesheetsDeleteEntryCmd.Flags().BoolVar(&timesheetsDeleteEntryForceFlag, "force", false, "Confirm deletion")

	// Review command flags
	timesheetsReviewCmd.Flags().StringVar(&reviewCommentFlag, "comment", "", "Comment for the review (optional)")

	// Presets create command flags
	presetsCreateCmd.Flags().StringVar(&presetsCreateNameFlag, "name", "", "Preset name (required)")
	presetsCreateCmd.Flags().StringVar(&presetsCreateHoursPerDayFlag, "hours-per-day", "", "Hours per day (required)")
	presetsCreateCmd.Flags().StringVar(&presetsCreateHoursPerWeekFlag, "hours-per-week", "", "Hours per week (required)")

	// Presets update command flags
	presetsUpdateCmd.Flags().StringVar(&presetsUpdateNameFlag, "name", "", "Preset name")
	presetsUpdateCmd.Flags().StringVar(&presetsUpdateDescriptionFlag, "description", "", "Preset description")
	presetsUpdateCmd.Flags().StringVar(&presetsUpdateHoursPerDayFlag, "hours-per-day", "", "Hours per day")
	presetsUpdateCmd.Flags().StringVar(&presetsUpdateHoursPerWeekFlag, "hours-per-week", "", "Hours per week")
	presetsUpdateCmd.Flags().StringVar(&presetsUpdateRateFlag, "rate", "", "Rate")
	presetsUpdateCmd.Flags().StringVar(&presetsUpdateCurrencyFlag, "currency", "", "Currency code")

	// Add subcommands to presets
	presetsCmd.AddCommand(presetsListCmd)
	presetsCmd.AddCommand(presetsCreateCmd)
	presetsCmd.AddCommand(presetsUpdateCmd)
	presetsCmd.AddCommand(presetsDeleteCmd)

	// Add subcommands to timesheets
	timesheetsCmd.AddCommand(timesheetsListCmd)
	timesheetsCmd.AddCommand(timesheetsGetCmd)
	timesheetsCmd.AddCommand(timesheetsCreateEntryCmd)
	timesheetsCmd.AddCommand(timesheetsUpdateEntryCmd)
	timesheetsCmd.AddCommand(timesheetsDeleteEntryCmd)
	timesheetsCmd.AddCommand(timesheetsReviewCmd)
	timesheetsCmd.AddCommand(presetsCmd)
}
