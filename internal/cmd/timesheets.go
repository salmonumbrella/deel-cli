package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var timesheetsCmd = &cobra.Command{
	Use:   "timesheets",
	Short: "Manage timesheets",
	Long:  "List, view, and manage timesheets and timesheet entries in your Deel organization.",
}

// Flags for list command
var (
	timesheetsListContractIDFlag string
	timesheetsListStatusFlag     string
)

var timesheetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List timesheets",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.TimesheetsListParams{
			ContractID: timesheetsListContractIDFlag,
			Status:     timesheetsListStatusFlag,
		}

		timesheets, err := client.ListTimesheets(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to list timesheets: %v", err)
			return err
		}

		return f.Output(func() {
			if len(timesheets) == 0 {
				f.PrintText("No timesheets found.")
				return
			}
			table := f.NewTable("ID", "CONTRACT ID", "STATUS", "PERIOD", "TOTAL HOURS", "CREATED")
			for _, ts := range timesheets {
				period := fmt.Sprintf("%s to %s", ts.PeriodStart, ts.PeriodEnd)
				table.AddRow(ts.ID, ts.ContractID, ts.Status, period, fmt.Sprintf("%.2f", ts.TotalHours), ts.CreatedAt)
			}
			table.Render()
		}, timesheets)
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
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		timesheet, err := client.GetTimesheet(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get timesheet: %v", err)
			return err
		}

		return f.Output(func() {
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
			f.PrintError("--timesheet-id flag is required")
			return fmt.Errorf("--timesheet-id flag is required")
		}
		if createEntryDateFlag == "" {
			f.PrintError("--date flag is required")
			return fmt.Errorf("--date flag is required")
		}
		if createEntryHoursFlag <= 0 {
			f.PrintError("--hours flag is required and must be greater than 0")
			return fmt.Errorf("--hours flag is required and must be greater than 0")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateTimesheetEntryParams{
			TimesheetID: createEntryTimesheetIDFlag,
			Date:        createEntryDateFlag,
			Hours:       createEntryHoursFlag,
			Description: createEntryDescriptionFlag,
		}

		entry, err := client.CreateTimesheetEntry(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create timesheet entry: %v", err)
			return err
		}

		return f.Output(func() {
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
			f.PrintError("Invalid action %q. Must be 'approve' or 'reject'", action)
			return fmt.Errorf("invalid action %q", action)
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.ReviewTimesheetParams{
			Status:  status,
			Comment: reviewCommentFlag,
		}

		timesheet, err := client.ReviewTimesheet(cmd.Context(), timesheetID, params)
		if err != nil {
			f.PrintError("Failed to review timesheet: %v", err)
			return err
		}

		return f.Output(func() {
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
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		presets, err := client.ListHourlyPresets(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list hourly presets: %v", err)
			return err
		}

		return f.Output(func() {
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
			f.PrintError("--name flag is required")
			return fmt.Errorf("--name flag is required")
		}
		if presetsCreateHoursPerDayFlag == "" {
			f.PrintError("--hours-per-day flag is required")
			return fmt.Errorf("--hours-per-day flag is required")
		}
		if presetsCreateHoursPerWeekFlag == "" {
			f.PrintError("--hours-per-week flag is required")
			return fmt.Errorf("--hours-per-week flag is required")
		}

		// Parse hours
		hoursPerDay, err := strconv.ParseFloat(presetsCreateHoursPerDayFlag, 64)
		if err != nil {
			f.PrintError("Invalid --hours-per-day value: %v", err)
			return fmt.Errorf("invalid --hours-per-day value: %w", err)
		}

		hoursPerWeek, err := strconv.ParseFloat(presetsCreateHoursPerWeekFlag, 64)
		if err != nil {
			f.PrintError("Invalid --hours-per-week value: %v", err)
			return fmt.Errorf("invalid --hours-per-week value: %w", err)
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateHourlyPresetParams{
			Name:         presetsCreateNameFlag,
			HoursPerDay:  hoursPerDay,
			HoursPerWeek: hoursPerWeek,
		}

		preset, err := client.CreateHourlyPreset(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create hourly preset: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Hourly preset created successfully")
			f.PrintText("ID:            " + preset.ID)
			f.PrintText("Name:          " + preset.Name)
			f.PrintText(fmt.Sprintf("Hours per day: %.2f", preset.HoursPerDay))
			f.PrintText(fmt.Sprintf("Hours per week: %.2f", preset.HoursPerWeek))
			f.PrintText("Created:       " + preset.CreatedAt)
		}, preset)
	},
}

var presetsDeleteCmd = &cobra.Command{
	Use:   "delete <preset-id>",
	Short: "Delete hourly preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		err = client.DeleteHourlyPreset(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to delete hourly preset: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Hourly preset deleted successfully")
		}, map[string]string{"status": "deleted", "id": args[0]})
	},
}

func init() {
	// List command flags
	timesheetsListCmd.Flags().StringVar(&timesheetsListContractIDFlag, "contract-id", "", "Filter by contract ID")
	timesheetsListCmd.Flags().StringVar(&timesheetsListStatusFlag, "status", "", "Filter by status (e.g., pending, approved, rejected)")

	// Create entry command flags
	timesheetsCreateEntryCmd.Flags().StringVar(&createEntryTimesheetIDFlag, "timesheet-id", "", "Timesheet ID (required)")
	timesheetsCreateEntryCmd.Flags().StringVar(&createEntryDateFlag, "date", "", "Date for the entry (YYYY-MM-DD format, required)")
	timesheetsCreateEntryCmd.Flags().Float64Var(&createEntryHoursFlag, "hours", 0, "Hours worked (required)")
	timesheetsCreateEntryCmd.Flags().StringVar(&createEntryDescriptionFlag, "description", "", "Description of work (optional)")

	// Review command flags
	timesheetsReviewCmd.Flags().StringVar(&reviewCommentFlag, "comment", "", "Comment for the review (optional)")

	// Presets create command flags
	presetsCreateCmd.Flags().StringVar(&presetsCreateNameFlag, "name", "", "Preset name (required)")
	presetsCreateCmd.Flags().StringVar(&presetsCreateHoursPerDayFlag, "hours-per-day", "", "Hours per day (required)")
	presetsCreateCmd.Flags().StringVar(&presetsCreateHoursPerWeekFlag, "hours-per-week", "", "Hours per week (required)")

	// Add subcommands to presets
	presetsCmd.AddCommand(presetsListCmd)
	presetsCmd.AddCommand(presetsCreateCmd)
	presetsCmd.AddCommand(presetsDeleteCmd)

	// Add subcommands to timesheets
	timesheetsCmd.AddCommand(timesheetsListCmd)
	timesheetsCmd.AddCommand(timesheetsGetCmd)
	timesheetsCmd.AddCommand(timesheetsCreateEntryCmd)
	timesheetsCmd.AddCommand(timesheetsReviewCmd)
	timesheetsCmd.AddCommand(presetsCmd)
}
