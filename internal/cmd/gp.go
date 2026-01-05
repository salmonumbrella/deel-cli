package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var gpCmd = &cobra.Command{
	Use:   "gp",
	Short: "Manage Global Payroll contracts and workers",
	Long:  "Create, view, and manage Global Payroll (GP) contracts, workers, shifts, and related operations.",
}

// Flags for create command
var (
	gpCreateWorkerEmailFlag  string
	gpCreateWorkerNameFlag   string
	gpCreateCountryFlag      string
	gpCreateStartDateFlag    string
	gpCreateJobTitleFlag     string
	gpCreateSalaryFlag       string
	gpCreateCurrencyFlag     string
	gpCreatePayFrequencyFlag string
)

var gpCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create GP contract",
	Long:  "Create a new Global Payroll contract. Requires --worker-email, --worker-name, --country, --start-date, --job-title, --salary, --currency, and --pay-frequency flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Validate required flags
		if gpCreateWorkerEmailFlag == "" {
			f.PrintError("--worker-email flag is required")
			return fmt.Errorf("--worker-email flag is required")
		}
		if gpCreateWorkerNameFlag == "" {
			f.PrintError("--worker-name flag is required")
			return fmt.Errorf("--worker-name flag is required")
		}
		if gpCreateCountryFlag == "" {
			f.PrintError("--country flag is required")
			return fmt.Errorf("--country flag is required")
		}
		if gpCreateStartDateFlag == "" {
			f.PrintError("--start-date flag is required")
			return fmt.Errorf("--start-date flag is required")
		}
		if gpCreateJobTitleFlag == "" {
			f.PrintError("--job-title flag is required")
			return fmt.Errorf("--job-title flag is required")
		}
		if gpCreateSalaryFlag == "" {
			f.PrintError("--salary flag is required")
			return fmt.Errorf("--salary flag is required")
		}
		if gpCreateCurrencyFlag == "" {
			f.PrintError("--currency flag is required")
			return fmt.Errorf("--currency flag is required")
		}
		if gpCreatePayFrequencyFlag == "" {
			f.PrintError("--pay-frequency flag is required")
			return fmt.Errorf("--pay-frequency flag is required")
		}

		// Parse salary
		salary, err := strconv.ParseFloat(gpCreateSalaryFlag, 64)
		if err != nil {
			f.PrintError("Invalid --salary value: %v", err)
			return fmt.Errorf("invalid --salary value: %w", err)
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "GPContract",
			Description: "Create GP contract",
			Details: map[string]string{
				"WorkerEmail":  gpCreateWorkerEmailFlag,
				"WorkerName":   gpCreateWorkerNameFlag,
				"Country":      gpCreateCountryFlag,
				"StartDate":    gpCreateStartDateFlag,
				"JobTitle":     gpCreateJobTitleFlag,
				"Salary":       fmt.Sprintf("%.2f %s", salary, gpCreateCurrencyFlag),
				"PayFrequency": gpCreatePayFrequencyFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateGPContractParams{
			WorkerEmail:  gpCreateWorkerEmailFlag,
			WorkerName:   gpCreateWorkerNameFlag,
			Country:      gpCreateCountryFlag,
			StartDate:    gpCreateStartDateFlag,
			JobTitle:     gpCreateJobTitleFlag,
			Salary:       salary,
			Currency:     gpCreateCurrencyFlag,
			PayFrequency: gpCreatePayFrequencyFlag,
		}

		contract, err := client.CreateGPContract(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create GP contract: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("GP contract created successfully")
			f.PrintText("ID:            " + contract.ID)
			f.PrintText("Worker ID:     " + contract.WorkerID)
			f.PrintText("Worker Email:  " + contract.WorkerEmail)
			f.PrintText("Worker Name:   " + contract.WorkerName)
			f.PrintText("Country:       " + contract.Country)
			f.PrintText("Start Date:    " + contract.StartDate)
			f.PrintText(fmt.Sprintf("Salary:        %.2f %s", contract.Salary, contract.Currency))
			f.PrintText("Pay Frequency: " + contract.PayFrequency)
			f.PrintText("Job Title:     " + contract.JobTitle)
			f.PrintText("Status:        " + contract.Status)
			f.PrintText("Created:       " + contract.CreatedAt)
		}, contract)
	},
}

// Bank accounts subcommand
var gpBankAccountsCmd = &cobra.Command{
	Use:   "bank-accounts",
	Short: "Manage GP worker bank accounts",
	Long:  "List and add bank accounts for Global Payroll workers.",
}

// Flags for bank-accounts list command
var gpBankAccountsListWorkerIDFlag string
var gpBankAccountsLimitFlag int

var gpBankAccountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bank accounts for worker",
	Long:  "List all bank accounts for a Global Payroll worker. Requires --worker-id flag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if gpBankAccountsListWorkerIDFlag == "" {
			f.PrintError("--worker-id flag is required")
			return fmt.Errorf("--worker-id flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		accounts, err := client.ListGPBankAccounts(cmd.Context(), gpBankAccountsListWorkerIDFlag)
		if err != nil {
			f.PrintError("Failed to list bank accounts: %v", err)
			return err
		}

		// Apply client-side limit
		if gpBankAccountsLimitFlag > 0 && len(accounts) > gpBankAccountsLimitFlag {
			accounts = accounts[:gpBankAccountsLimitFlag]
		}

		return f.Output(func() {
			if len(accounts) == 0 {
				f.PrintText("No bank accounts found")
				return
			}
			table := f.NewTable("ID", "ACCOUNT HOLDER", "BANK NAME", "ACCOUNT NUMBER", "CURRENCY", "PRIMARY", "STATUS")
			for _, account := range accounts {
				primaryStr := "No"
				if account.IsPrimary {
					primaryStr = "Yes"
				}
				table.AddRow(
					account.ID,
					account.AccountHolder,
					account.BankName,
					account.AccountNumber,
					account.Currency,
					primaryStr,
					account.Status,
				)
			}
			table.Render()
		}, accounts)
	},
}

// Flags for bank-accounts add command
var (
	gpBankAccountAddWorkerIDFlag      string
	gpBankAccountAddAccountHolderFlag string
	gpBankAccountAddBankNameFlag      string
	gpBankAccountAddAccountNumberFlag string
	gpBankAccountAddCurrencyFlag      string
)

var gpBankAccountsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add bank account to GP worker",
	Long:  "Add a bank account to a Global Payroll worker. Requires --worker-id, --account-holder, --bank-name, --account-number, and --currency flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if gpBankAccountAddWorkerIDFlag == "" {
			f.PrintError("--worker-id flag is required")
			return fmt.Errorf("--worker-id flag is required")
		}
		if gpBankAccountAddAccountHolderFlag == "" {
			f.PrintError("--account-holder flag is required")
			return fmt.Errorf("--account-holder flag is required")
		}
		if gpBankAccountAddBankNameFlag == "" {
			f.PrintError("--bank-name flag is required")
			return fmt.Errorf("--bank-name flag is required")
		}
		if gpBankAccountAddAccountNumberFlag == "" {
			f.PrintError("--account-number flag is required")
			return fmt.Errorf("--account-number flag is required")
		}
		if gpBankAccountAddCurrencyFlag == "" {
			f.PrintError("--currency flag is required")
			return fmt.Errorf("--currency flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "GPBankAccount",
			Description: "Add GP bank account",
			Details: map[string]string{
				"WorkerID":      gpBankAccountAddWorkerIDFlag,
				"AccountHolder": gpBankAccountAddAccountHolderFlag,
				"BankName":      gpBankAccountAddBankNameFlag,
				"AccountNumber": gpBankAccountAddAccountNumberFlag,
				"Currency":      gpBankAccountAddCurrencyFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.AddGPBankAccountParams{
			WorkerID:      gpBankAccountAddWorkerIDFlag,
			AccountHolder: gpBankAccountAddAccountHolderFlag,
			BankName:      gpBankAccountAddBankNameFlag,
			AccountNumber: gpBankAccountAddAccountNumberFlag,
			Currency:      gpBankAccountAddCurrencyFlag,
		}

		bankAccount, err := client.AddGPBankAccount(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to add bank account: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Bank account added successfully")
			f.PrintText("ID:             " + bankAccount.ID)
			f.PrintText("Worker ID:      " + bankAccount.WorkerID)
			f.PrintText("Account Holder: " + bankAccount.AccountHolder)
			f.PrintText("Bank Name:      " + bankAccount.BankName)
			f.PrintText("Account Number: " + bankAccount.AccountNumber)
			f.PrintText("Currency:       " + bankAccount.Currency)
			f.PrintText(fmt.Sprintf("Is Primary:     %t", bankAccount.IsPrimary))
			f.PrintText("Status:         " + bankAccount.Status)
		}, bankAccount)
	},
}

// Reports subcommand
var gpReportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Manage GP reports",
	Long:  "Access and manage Global Payroll reports.",
}

// Flags for reports g2n command
var (
	gpReportsG2NWorkerIDFlag string
	gpReportsG2NPeriodFlag   string
)

var gpReportsG2NCmd = &cobra.Command{
	Use:   "g2n",
	Short: "List gross-to-net reports",
	Long:  "List gross-to-net reports for Global Payroll workers. Requires --worker-id flag. Optional --period flag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if gpReportsG2NWorkerIDFlag == "" {
			f.PrintError("--worker-id flag is required")
			return fmt.Errorf("--worker-id flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.ListG2NReportsParams{
			WorkerID: gpReportsG2NWorkerIDFlag,
			Period:   gpReportsG2NPeriodFlag,
		}

		reports, err := client.ListG2NReports(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to list gross-to-net reports: %v", err)
			return err
		}

		return f.Output(func() {
			if len(reports) == 0 {
				f.PrintText("No gross-to-net reports found")
				return
			}
			table := f.NewTable("ID", "WORKER NAME", "PERIOD", "GROSS", "NET", "DEDUCTIONS", "TAXES", "STATUS")
			for _, report := range reports {
				table.AddRow(
					report.ID,
					report.WorkerName,
					report.Period,
					fmt.Sprintf("%.2f %s", report.GrossAmount, report.Currency),
					fmt.Sprintf("%.2f %s", report.NetAmount, report.Currency),
					fmt.Sprintf("%.2f", report.Deductions),
					fmt.Sprintf("%.2f", report.Taxes),
					report.Status,
				)
			}
			table.Render()
		}, reports)
	},
}

// Flags for terminate command
var (
	gpTerminateWorkerIDFlag      string
	gpTerminateReasonFlag        string
	gpTerminateEffectiveDateFlag string
)

var gpTerminateCmd = &cobra.Command{
	Use:   "terminate",
	Short: "Request termination for GP worker",
	Long:  "Request termination for a Global Payroll worker. Requires --worker-id, --reason, and --effective-date flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if gpTerminateWorkerIDFlag == "" {
			f.PrintError("--worker-id flag is required")
			return fmt.Errorf("--worker-id flag is required")
		}
		if gpTerminateReasonFlag == "" {
			f.PrintError("--reason flag is required")
			return fmt.Errorf("--reason flag is required")
		}
		if gpTerminateEffectiveDateFlag == "" {
			f.PrintError("--effective-date flag is required")
			return fmt.Errorf("--effective-date flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "TERMINATE",
			Resource:    "GPWorker",
			Description: "Request GP termination",
			Details: map[string]string{
				"WorkerID":      gpTerminateWorkerIDFlag,
				"Reason":        gpTerminateReasonFlag,
				"EffectiveDate": gpTerminateEffectiveDateFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.RequestGPTerminationParams{
			WorkerID:      gpTerminateWorkerIDFlag,
			Reason:        gpTerminateReasonFlag,
			EffectiveDate: gpTerminateEffectiveDateFlag,
		}

		termination, err := client.RequestGPTermination(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to request GP termination: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("GP termination requested successfully")
			f.PrintText("ID:             " + termination.ID)
			f.PrintText("Worker ID:      " + termination.WorkerID)
			f.PrintText("Reason:         " + termination.Reason)
			f.PrintText("Effective Date: " + termination.EffectiveDate)
			f.PrintText("Status:         " + termination.Status)
			f.PrintText("Created:        " + termination.CreatedAt)
		}, termination)
	},
}

// Shifts subcommand
var gpShiftsCmd = &cobra.Command{
	Use:   "shifts",
	Short: "Manage GP shifts",
	Long:  "Create and manage Global Payroll shifts.",
}

// Flags for shifts create command
var (
	gpShiftsCreateWorkerIDFlag     string
	gpShiftsCreateDateFlag         string
	gpShiftsCreateStartTimeFlag    string
	gpShiftsCreateEndTimeFlag      string
	gpShiftsCreateBreakMinutesFlag int
)

var gpShiftsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create shift",
	Long:  "Create a new Global Payroll shift. Requires --worker-id, --date, --start-time, --end-time, and --break-minutes flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if gpShiftsCreateWorkerIDFlag == "" {
			f.PrintError("--worker-id flag is required")
			return fmt.Errorf("--worker-id flag is required")
		}
		if gpShiftsCreateDateFlag == "" {
			f.PrintError("--date flag is required")
			return fmt.Errorf("--date flag is required")
		}
		if gpShiftsCreateStartTimeFlag == "" {
			f.PrintError("--start-time flag is required")
			return fmt.Errorf("--start-time flag is required")
		}
		if gpShiftsCreateEndTimeFlag == "" {
			f.PrintError("--end-time flag is required")
			return fmt.Errorf("--end-time flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "GPShift",
			Description: "Create GP shift",
			Details: map[string]string{
				"WorkerID":     gpShiftsCreateWorkerIDFlag,
				"Date":         gpShiftsCreateDateFlag,
				"StartTime":    gpShiftsCreateStartTimeFlag,
				"EndTime":      gpShiftsCreateEndTimeFlag,
				"BreakMinutes": fmt.Sprintf("%d", gpShiftsCreateBreakMinutesFlag),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateGPShiftParams{
			WorkerID:     gpShiftsCreateWorkerIDFlag,
			Date:         gpShiftsCreateDateFlag,
			StartTime:    gpShiftsCreateStartTimeFlag,
			EndTime:      gpShiftsCreateEndTimeFlag,
			BreakMinutes: gpShiftsCreateBreakMinutesFlag,
		}

		shift, err := client.CreateGPShift(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create GP shift: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("GP shift created successfully")
			f.PrintText("ID:            " + shift.ID)
			f.PrintText("Worker ID:     " + shift.WorkerID)
			f.PrintText("Date:          " + shift.Date)
			f.PrintText("Start Time:    " + shift.StartTime)
			f.PrintText("End Time:      " + shift.EndTime)
			f.PrintText(fmt.Sprintf("Break Minutes: %d", shift.BreakMinutes))
			f.PrintText(fmt.Sprintf("Total Hours:   %.2f", shift.TotalHours))
			f.PrintText("Status:        " + shift.Status)
			f.PrintText("Created:       " + shift.CreatedAt)
		}, shift)
	},
}

// Rates subcommand
var gpRatesCmd = &cobra.Command{
	Use:   "rates",
	Short: "Manage GP shift rates",
	Long:  "List and create Global Payroll shift rates.",
}

var gpRatesLimitFlag int

var gpRatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List shift rates",
	Long:  "List all Global Payroll shift rates.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		rates, err := client.ListGPShiftRates(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list shift rates: %v", err)
			return err
		}

		// Apply client-side limit
		if gpRatesLimitFlag > 0 && len(rates) > gpRatesLimitFlag {
			rates = rates[:gpRatesLimitFlag]
		}

		return f.Output(func() {
			if len(rates) == 0 {
				f.PrintText("No shift rates found")
				return
			}
			table := f.NewTable("ID", "NAME", "RATE", "CURRENCY", "TYPE", "STATUS")
			for _, rate := range rates {
				table.AddRow(
					rate.ID,
					rate.Name,
					fmt.Sprintf("%.2f", rate.Rate),
					rate.Currency,
					rate.Type,
					rate.Status,
				)
			}
			table.Render()
		}, rates)
	},
}

// Flags for rates create command
var (
	gpRatesCreateNameFlag     string
	gpRatesCreateRateFlag     string
	gpRatesCreateCurrencyFlag string
	gpRatesCreateTypeFlag     string
)

var gpRatesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create rate",
	Long:  "Create a new Global Payroll shift rate. Requires --name, --rate, --currency, and --type flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if gpRatesCreateNameFlag == "" {
			f.PrintError("--name flag is required")
			return fmt.Errorf("--name flag is required")
		}
		if gpRatesCreateRateFlag == "" {
			f.PrintError("--rate flag is required")
			return fmt.Errorf("--rate flag is required")
		}
		if gpRatesCreateCurrencyFlag == "" {
			f.PrintError("--currency flag is required")
			return fmt.Errorf("--currency flag is required")
		}
		if gpRatesCreateTypeFlag == "" {
			f.PrintError("--type flag is required")
			return fmt.Errorf("--type flag is required")
		}

		// Parse rate
		rate, err := strconv.ParseFloat(gpRatesCreateRateFlag, 64)
		if err != nil {
			f.PrintError("Invalid --rate value: %v", err)
			return fmt.Errorf("invalid --rate value: %w", err)
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "GPShiftRate",
			Description: "Create GP shift rate",
			Details: map[string]string{
				"Name":     gpRatesCreateNameFlag,
				"Rate":     fmt.Sprintf("%.2f %s", rate, gpRatesCreateCurrencyFlag),
				"Type":     gpRatesCreateTypeFlag,
				"Currency": gpRatesCreateCurrencyFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.GPCreateShiftRateParams{
			Name:     gpRatesCreateNameFlag,
			Rate:     rate,
			Currency: gpRatesCreateCurrencyFlag,
			Type:     gpRatesCreateTypeFlag,
		}

		shiftRate, err := client.CreateGPShiftRate(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create shift rate: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Shift rate created successfully")
			f.PrintText("ID:       " + shiftRate.ID)
			f.PrintText("Name:     " + shiftRate.Name)
			f.PrintText(fmt.Sprintf("Rate:     %.2f %s", shiftRate.Rate, shiftRate.Currency))
			f.PrintText("Type:     " + shiftRate.Type)
			f.PrintText("Status:   " + shiftRate.Status)
			f.PrintText("Created:  " + shiftRate.CreatedAt)
		}, shiftRate)
	},
}

func init() {
	// Create command flags
	gpCreateCmd.Flags().StringVar(&gpCreateWorkerEmailFlag, "worker-email", "", "Worker email (required)")
	gpCreateCmd.Flags().StringVar(&gpCreateWorkerNameFlag, "worker-name", "", "Worker name (required)")
	gpCreateCmd.Flags().StringVar(&gpCreateCountryFlag, "country", "", "Country code (required)")
	gpCreateCmd.Flags().StringVar(&gpCreateStartDateFlag, "start-date", "", "Start date YYYY-MM-DD (required)")
	gpCreateCmd.Flags().StringVar(&gpCreateJobTitleFlag, "job-title", "", "Job title (required)")
	gpCreateCmd.Flags().StringVar(&gpCreateSalaryFlag, "salary", "", "Annual salary (required)")
	gpCreateCmd.Flags().StringVar(&gpCreateCurrencyFlag, "currency", "", "Currency code (required)")
	gpCreateCmd.Flags().StringVar(&gpCreatePayFrequencyFlag, "pay-frequency", "", "Pay frequency (required)")

	// Bank accounts list command flags
	gpBankAccountsListCmd.Flags().StringVar(&gpBankAccountsListWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	gpBankAccountsListCmd.Flags().IntVar(&gpBankAccountsLimitFlag, "limit", 100, "Maximum results")

	// Bank accounts add command flags
	gpBankAccountsAddCmd.Flags().StringVar(&gpBankAccountAddWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	gpBankAccountsAddCmd.Flags().StringVar(&gpBankAccountAddAccountHolderFlag, "account-holder", "", "Account holder name (required)")
	gpBankAccountsAddCmd.Flags().StringVar(&gpBankAccountAddBankNameFlag, "bank-name", "", "Bank name (required)")
	gpBankAccountsAddCmd.Flags().StringVar(&gpBankAccountAddAccountNumberFlag, "account-number", "", "Account number (required)")
	gpBankAccountsAddCmd.Flags().StringVar(&gpBankAccountAddCurrencyFlag, "currency", "", "Currency code (required)")

	// Reports g2n command flags
	gpReportsG2NCmd.Flags().StringVar(&gpReportsG2NWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	gpReportsG2NCmd.Flags().StringVar(&gpReportsG2NPeriodFlag, "period", "", "Period (optional)")

	// Terminate command flags
	gpTerminateCmd.Flags().StringVar(&gpTerminateWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	gpTerminateCmd.Flags().StringVar(&gpTerminateReasonFlag, "reason", "", "Termination reason (required)")
	gpTerminateCmd.Flags().StringVar(&gpTerminateEffectiveDateFlag, "effective-date", "", "Effective date YYYY-MM-DD (required)")

	// Shifts create command flags
	gpShiftsCreateCmd.Flags().StringVar(&gpShiftsCreateWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	gpShiftsCreateCmd.Flags().StringVar(&gpShiftsCreateDateFlag, "date", "", "Date YYYY-MM-DD (required)")
	gpShiftsCreateCmd.Flags().StringVar(&gpShiftsCreateStartTimeFlag, "start-time", "", "Start time HH:MM (required)")
	gpShiftsCreateCmd.Flags().StringVar(&gpShiftsCreateEndTimeFlag, "end-time", "", "End time HH:MM (required)")
	gpShiftsCreateCmd.Flags().IntVar(&gpShiftsCreateBreakMinutesFlag, "break-minutes", 0, "Break minutes (required)")

	// Rates list command flags
	gpRatesListCmd.Flags().IntVar(&gpRatesLimitFlag, "limit", 100, "Maximum results")

	// Rates create command flags
	gpRatesCreateCmd.Flags().StringVar(&gpRatesCreateNameFlag, "name", "", "Rate name (required)")
	gpRatesCreateCmd.Flags().StringVar(&gpRatesCreateRateFlag, "rate", "", "Rate amount (required)")
	gpRatesCreateCmd.Flags().StringVar(&gpRatesCreateCurrencyFlag, "currency", "", "Currency code (required)")
	gpRatesCreateCmd.Flags().StringVar(&gpRatesCreateTypeFlag, "type", "", "Rate type: hourly, daily, or flat (required)")

	// Add subcommands to bank-accounts
	gpBankAccountsCmd.AddCommand(gpBankAccountsListCmd)
	gpBankAccountsCmd.AddCommand(gpBankAccountsAddCmd)

	// Add subcommands to reports
	gpReportsCmd.AddCommand(gpReportsG2NCmd)

	// Add subcommands to shifts
	gpShiftsCmd.AddCommand(gpShiftsCreateCmd)

	// Add subcommands to rates
	gpRatesCmd.AddCommand(gpRatesListCmd)
	gpRatesCmd.AddCommand(gpRatesCreateCmd)

	// Add subcommands to gp
	gpCmd.AddCommand(gpCreateCmd)
	gpCmd.AddCommand(gpBankAccountsCmd)
	gpCmd.AddCommand(gpReportsCmd)
	gpCmd.AddCommand(gpTerminateCmd)
	gpCmd.AddCommand(gpShiftsCmd)
	gpCmd.AddCommand(gpRatesCmd)
}
