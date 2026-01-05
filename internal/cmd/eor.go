package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var eorCmd = &cobra.Command{
	Use:   "eor",
	Short: "Manage EOR contracts and workers",
	Long:  "Create, view, and manage Employer of Record (EOR) contracts, workers, and related operations.",
}

// Flags for create command
var (
	eorCreateTitleFlag        string
	eorCreateWorkerEmailFlag  string
	eorCreateWorkerNameFlag   string
	eorCreateCountryFlag      string
	eorCreateStartDateFlag    string
	eorCreateSalaryFlag       string
	eorCreateCurrencyFlag     string
	eorCreatePayFrequencyFlag string
	eorCreateJobTitleFlag     string
	eorCreateSeniorityFlag    string
	eorCreateScopeFlag        string
)

var eorCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create EOR contract",
	Long:  "Create a new Employer of Record contract. Requires --title, --worker-email, --worker-name, --country, --start-date, --salary, --currency, --pay-frequency, and --job-title flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Validate required flags
		if eorCreateTitleFlag == "" {
			f.PrintError("--title flag is required")
			return fmt.Errorf("--title flag is required")
		}
		if eorCreateWorkerEmailFlag == "" {
			f.PrintError("--worker-email flag is required")
			return fmt.Errorf("--worker-email flag is required")
		}
		if eorCreateWorkerNameFlag == "" {
			f.PrintError("--worker-name flag is required")
			return fmt.Errorf("--worker-name flag is required")
		}
		if eorCreateCountryFlag == "" {
			f.PrintError("--country flag is required")
			return fmt.Errorf("--country flag is required")
		}
		if eorCreateStartDateFlag == "" {
			f.PrintError("--start-date flag is required")
			return fmt.Errorf("--start-date flag is required")
		}
		if eorCreateSalaryFlag == "" {
			f.PrintError("--salary flag is required")
			return fmt.Errorf("--salary flag is required")
		}
		if eorCreateCurrencyFlag == "" {
			f.PrintError("--currency flag is required")
			return fmt.Errorf("--currency flag is required")
		}
		if eorCreatePayFrequencyFlag == "" {
			f.PrintError("--pay-frequency flag is required")
			return fmt.Errorf("--pay-frequency flag is required")
		}
		if eorCreateJobTitleFlag == "" {
			f.PrintError("--job-title flag is required")
			return fmt.Errorf("--job-title flag is required")
		}

		// Parse salary
		salary, err := strconv.ParseFloat(eorCreateSalaryFlag, 64)
		if err != nil {
			f.PrintError("Invalid --salary value: %v", err)
			return fmt.Errorf("invalid --salary value: %w", err)
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "EORContract",
			Description: "Create EOR contract",
			Details: map[string]string{
				"Title":        eorCreateTitleFlag,
				"WorkerEmail":  eorCreateWorkerEmailFlag,
				"WorkerName":   eorCreateWorkerNameFlag,
				"Country":      eorCreateCountryFlag,
				"StartDate":    eorCreateStartDateFlag,
				"Salary":       fmt.Sprintf("%.2f %s", salary, eorCreateCurrencyFlag),
				"PayFrequency": eorCreatePayFrequencyFlag,
				"JobTitle":     eorCreateJobTitleFlag,
				"Seniority":    eorCreateSeniorityFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateEORContractParams{
			Title:          eorCreateTitleFlag,
			WorkerEmail:    eorCreateWorkerEmailFlag,
			WorkerName:     eorCreateWorkerNameFlag,
			Country:        eorCreateCountryFlag,
			StartDate:      eorCreateStartDateFlag,
			Salary:         salary,
			Currency:       eorCreateCurrencyFlag,
			PayFrequency:   eorCreatePayFrequencyFlag,
			JobTitle:       eorCreateJobTitleFlag,
			SeniorityLevel: eorCreateSeniorityFlag,
			Scope:          eorCreateScopeFlag,
		}

		contract, err := client.CreateEORContract(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create EOR contract: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("EOR contract created successfully")
			f.PrintText("ID:            " + contract.ID)
			f.PrintText("Title:         " + contract.Title)
			f.PrintText("Status:        " + contract.Status)
			f.PrintText("Worker Email:  " + contract.WorkerEmail)
			f.PrintText("Worker Name:   " + contract.WorkerName)
			f.PrintText("Country:       " + contract.Country)
			f.PrintText("Start Date:    " + contract.StartDate)
			f.PrintText(fmt.Sprintf("Salary:        %.2f %s", contract.Salary, contract.Currency))
			f.PrintText("Pay Frequency: " + contract.PayFrequency)
			f.PrintText("Job Title:     " + contract.JobTitle)
			if contract.SeniorityLevel != "" {
				f.PrintText("Seniority:     " + contract.SeniorityLevel)
			}
			f.PrintText("Created:       " + contract.CreatedAt)
		}, contract)
	},
}

var eorGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get EOR contract details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "SIGN",
			Resource:    "EORContract",
			Description: "Sign EOR contract",
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

		contract, err := client.GetEORContract(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get EOR contract: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:            " + contract.ID)
			f.PrintText("Title:         " + contract.Title)
			f.PrintText("Status:        " + contract.Status)
			f.PrintText("Worker ID:     " + contract.WorkerID)
			f.PrintText("Worker Email:  " + contract.WorkerEmail)
			f.PrintText("Worker Name:   " + contract.WorkerName)
			f.PrintText("Country:       " + contract.Country)
			f.PrintText("Start Date:    " + contract.StartDate)
			if contract.EndDate != "" {
				f.PrintText("End Date:      " + contract.EndDate)
			}
			f.PrintText(fmt.Sprintf("Salary:        %.2f %s", contract.Salary, contract.Currency))
			f.PrintText("Pay Frequency: " + contract.PayFrequency)
			f.PrintText("Job Title:     " + contract.JobTitle)
			if contract.SeniorityLevel != "" {
				f.PrintText("Seniority:     " + contract.SeniorityLevel)
			}
			if contract.Scope != "" {
				f.PrintText("Scope:         " + contract.Scope)
			}
			f.PrintText("Created:       " + contract.CreatedAt)
			if len(contract.Benefits) > 0 {
				f.PrintText("")
				f.PrintText("Benefits:")
				table := f.NewTable("  ID", "NAME", "DESCRIPTION", "AMOUNT")
				for _, benefit := range contract.Benefits {
					desc := benefit.Description
					if desc == "" {
						desc = "-"
					}
					amt := "-"
					if benefit.Amount > 0 {
						amt = fmt.Sprintf("%.2f", benefit.Amount)
					}
					table.AddRow("  "+benefit.ID, benefit.Name, desc, amt)
				}
				table.Render()
			}
		}, contract)
	},
}

var eorSignCmd = &cobra.Command{
	Use:   "sign <id>",
	Short: "Sign EOR contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		contract, err := client.SignEORContract(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to sign EOR contract: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("EOR contract signed successfully")
			f.PrintText("ID:     " + contract.ID)
			f.PrintText("Title:  " + contract.Title)
			f.PrintText("Status: " + contract.Status)
		}, contract)
	},
}

// Flags for cancel command
var eorCancelReasonFlag string

var eorCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel EOR contract",
	Long:  "Cancel an EOR contract. Requires --reason flag.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if eorCancelReasonFlag == "" {
			f.PrintError("--reason flag is required")
			return fmt.Errorf("--reason flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CANCEL",
			Resource:    "EORContract",
			Description: "Cancel EOR contract",
			Details: map[string]string{
				"ID":     args[0],
				"Reason": eorCancelReasonFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CancelEORContractParams{
			Reason: eorCancelReasonFlag,
		}

		contract, err := client.CancelEORContract(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to cancel EOR contract: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("EOR contract cancelled successfully")
			f.PrintText("ID:     " + contract.ID)
			f.PrintText("Title:  " + contract.Title)
			f.PrintText("Status: " + contract.Status)
		}, contract)
	},
}

// Flags for amend command
var (
	eorAmendTypeFlag          string
	eorAmendEffectiveDateFlag string
	eorAmendReasonFlag        string
	eorAmendSalaryFlag        string
	eorAmendJobTitleFlag      string
	eorAmendSeniorityFlag     string
	eorAmendScopeFlag         string
)

var eorAmendCmd = &cobra.Command{
	Use:   "amend <id>",
	Short: "Create amendment for EOR contract",
	Long:  "Create an amendment for an EOR contract. Requires --type, --effective-date, and --reason flags. Additional changes via --salary, --job-title, --seniority, --scope flags.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if eorAmendTypeFlag == "" {
			f.PrintError("--type flag is required")
			return fmt.Errorf("--type flag is required")
		}
		if eorAmendEffectiveDateFlag == "" {
			f.PrintError("--effective-date flag is required")
			return fmt.Errorf("--effective-date flag is required")
		}
		if eorAmendReasonFlag == "" {
			f.PrintError("--reason flag is required")
			return fmt.Errorf("--reason flag is required")
		}

		// Build changes map
		changes := make(map[string]interface{})
		if eorAmendSalaryFlag != "" {
			salary, err := strconv.ParseFloat(eorAmendSalaryFlag, 64)
			if err != nil {
				f.PrintError("Invalid --salary value: %v", err)
				return fmt.Errorf("invalid --salary value: %w", err)
			}
			changes["salary"] = salary
		}
		if eorAmendJobTitleFlag != "" {
			changes["job_title"] = eorAmendJobTitleFlag
		}
		if eorAmendSeniorityFlag != "" {
			changes["seniority_level"] = eorAmendSeniorityFlag
		}
		if eorAmendScopeFlag != "" {
			changes["scope"] = eorAmendScopeFlag
		}

		if len(changes) == 0 {
			f.PrintError("At least one change flag (--salary, --job-title, --seniority, --scope) is required")
			return fmt.Errorf("at least one change flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "AMEND",
			Resource:    "EORContract",
			Description: "Create EOR amendment",
			Details: map[string]string{
				"ID":            args[0],
				"Type":          eorAmendTypeFlag,
				"EffectiveDate": eorAmendEffectiveDateFlag,
				"Reason":        eorAmendReasonFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateEORAmendmentParams{
			Type:          eorAmendTypeFlag,
			Changes:       changes,
			EffectiveDate: eorAmendEffectiveDateFlag,
			Reason:        eorAmendReasonFlag,
		}

		amendment, err := client.CreateEORAmendment(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to create EOR amendment: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("EOR amendment created successfully")
			f.PrintText("ID:             " + amendment.ID)
			f.PrintText("Contract ID:    " + amendment.ContractID)
			f.PrintText("Type:           " + amendment.Type)
			f.PrintText("Status:         " + amendment.Status)
			f.PrintText("Effective Date: " + amendment.EffectiveDate)
			f.PrintText("Reason:         " + amendment.Reason)
			f.PrintText("Created:        " + amendment.CreatedAt)
			if len(amendment.Changes) > 0 {
				f.PrintText("")
				f.PrintText("Changes:")
				for key, val := range amendment.Changes {
					f.PrintText(fmt.Sprintf("  %s: %v", key, val))
				}
			}
		}, amendment)
	},
}

// eorAmendmentsCmd is the parent command for amendment operations
var eorAmendmentsCmd = &cobra.Command{
	Use:   "amendments",
	Short: "Manage EOR contract amendments",
	Long:  "List, sign, and accept amendments for EOR contracts.",
}

var eorAmendmentsLimitFlag int

var eorAmendmentsListCmd = &cobra.Command{
	Use:   "list <contract-id>",
	Short: "List amendments for an EOR contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "listing amendments")
		}

		amendments, err := client.ListEORAmendments(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "listing amendments")
		}

		// Apply client-side limit
		if eorAmendmentsLimitFlag > 0 && len(amendments) > eorAmendmentsLimitFlag {
			amendments = amendments[:eorAmendmentsLimitFlag]
		}

		return f.Output(func() {
			if len(amendments) == 0 {
				f.PrintText("No amendments found.")
				return
			}
			table := f.NewTable("ID", "TYPE", "STATUS", "EFFECTIVE DATE", "CREATED")
			for _, a := range amendments {
				table.AddRow(a.ID, a.Type, a.Status, a.EffectiveDate, a.CreatedAt)
			}
			table.Render()
		}, amendments)
	},
}

var eorAmendmentsSignCmd = &cobra.Command{
	Use:   "sign <amendment-id>",
	Short: "Sign an EOR amendment",
	Long:  "Sign an EOR contract amendment. This is typically done after the amendment has been accepted.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "SIGN",
			Resource:    "EORAmendment",
			Description: "Sign EOR amendment",
			Details: map[string]string{
				"AmendmentID": args[0],
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "signing amendment")
		}

		amendment, err := client.SignEORAmendment(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "signing amendment")
		}

		return f.Output(func() {
			f.PrintSuccess("Amendment signed successfully")
			f.PrintText("ID:             " + amendment.ID)
			f.PrintText("Contract ID:    " + amendment.ContractID)
			f.PrintText("Type:           " + amendment.Type)
			f.PrintText("Status:         " + amendment.Status)
			f.PrintText("Effective Date: " + amendment.EffectiveDate)
			if amendment.SignedAt != "" {
				f.PrintText("Signed At:      " + amendment.SignedAt)
			}
		}, amendment)
	},
}

var eorAmendmentsAcceptCmd = &cobra.Command{
	Use:   "accept <amendment-id>",
	Short: "Accept an EOR amendment",
	Long:  "Accept an EOR contract amendment. This is typically done before signing.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "ACCEPT",
			Resource:    "EORAmendment",
			Description: "Accept EOR amendment",
			Details: map[string]string{
				"AmendmentID": args[0],
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "accepting amendment")
		}

		amendment, err := client.AcceptEORAmendment(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "accepting amendment")
		}

		return f.Output(func() {
			f.PrintSuccess("Amendment accepted successfully")
			f.PrintText("ID:             " + amendment.ID)
			f.PrintText("Contract ID:    " + amendment.ContractID)
			f.PrintText("Type:           " + amendment.Type)
			f.PrintText("Status:         " + amendment.Status)
			f.PrintText("Effective Date: " + amendment.EffectiveDate)
			if amendment.AcceptedAt != "" {
				f.PrintText("Accepted At:    " + amendment.AcceptedAt)
			}
		}, amendment)
	},
}

// Flags for terminate command
var (
	eorTerminateReasonFlag       string
	eorTerminateReasonDetailFlag string
	eorTerminateNotifiedFlag     bool
	eorTerminateSensitiveFlag    bool
	eorTerminateSeveranceFlag    string
	eorTerminatePTOFlag          int
	eorTerminateUnpaidFlag       int
	eorTerminateSickFlag         int
)

var eorTerminateCmd = &cobra.Command{
	Use:   "terminate <oid>",
	Short: "Request termination for EOR contract",
	Long: `Request termination for an EOR contract (employer-initiated).

Requires --reason and --reason-detail flags.

Available reasons:
  TERMINATION, FOR_CAUSE, PERFORMANCE, PERFORMANCE_ISSUES, ATTENDANCE_ISSUES,
  MISCONDUCT, FALSIFYING, HARASSMENT, VIOLENCE, STEALING, POSITION_ELIMINATION,
  FORCE_REDUCTION, REORGANIZATION_DOWNSIZING_BUDGET_OR_REDUCTION_OF_WORKFORCE,
  ROLE_BECAME_REDUNDANT_OR_ROLE_CHANGED, NON_RENEWAL, PROBATION, and more.

Example:
  deel eor terminate abc123 --reason TERMINATION --reason-detail "Position eliminated due to restructuring" --notified`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if eorTerminateReasonFlag == "" {
			f.PrintError("--reason flag is required")
			return fmt.Errorf("--reason flag is required")
		}
		if eorTerminateReasonDetailFlag == "" {
			f.PrintError("--reason-detail flag is required (100-5000 chars)")
			return fmt.Errorf("--reason-detail flag is required")
		}
		if len(eorTerminateReasonDetailFlag) < 100 {
			f.PrintError("--reason-detail must be at least 100 characters")
			return fmt.Errorf("--reason-detail must be at least 100 characters")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "TERMINATE",
			Resource:    "EORContract",
			Description: "Request EOR termination",
			Details: map[string]string{
				"OID":              args[0],
				"Reason":           eorTerminateReasonFlag,
				"ReasonDetail":     eorTerminateReasonDetailFlag[:min(50, len(eorTerminateReasonDetailFlag))] + "...",
				"EmployeeNotified": fmt.Sprintf("%t", eorTerminateNotifiedFlag),
				"Sensitive":        fmt.Sprintf("%t", eorTerminateSensitiveFlag),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.EORTerminationParams{
			Reason:             eorTerminateReasonFlag,
			ReasonDetail:       eorTerminateReasonDetailFlag,
			IsEmployeeNotified: eorTerminateNotifiedFlag,
			IsSensitive:        eorTerminateSensitiveFlag,
			SeveranceType:      eorTerminateSeveranceFlag,
			UsedTimeOff: api.EORUsedTimeOff{
				PaidTimeOff:   eorTerminatePTOFlag,
				UnpaidTimeOff: eorTerminateUnpaidFlag,
				SickLeave:     eorTerminateSickFlag,
			},
		}

		termination, err := client.RequestEORTermination(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to request EOR termination: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("EOR termination requested successfully")
			f.PrintText("ID:                " + termination.ID)
			f.PrintText("Contract ID:       " + termination.ContractID)
			f.PrintText("Type:              " + termination.Type)
			f.PrintText("Status:            " + termination.Status)
			f.PrintText("Reason:            " + termination.Reason)
			f.PrintText("Effective Date:    " + termination.EffectiveDate)
			f.PrintText("Last Working Day:  " + termination.LastWorkingDay)
			f.PrintText(fmt.Sprintf("Notice Period:     %d days", termination.NoticePeriodDays))
			if termination.SeveranceAmount > 0 {
				f.PrintText(fmt.Sprintf("Severance:         %.2f %s", termination.SeveranceAmount, termination.Currency))
			}
			f.PrintText("Created:           " + termination.CreatedAt)
		}, termination)
	},
}

// Workers subcommand
var workersCmd = &cobra.Command{
	Use:   "workers",
	Short: "Manage EOR workers",
	Long:  "Create and manage EOR workers.",
}

// Flags for workers create command
var (
	workersCreateEmailFlag     string
	workersCreateFirstNameFlag string
	workersCreateLastNameFlag  string
	workersCreateCountryFlag   string
	workersCreateDOBFlag       string
	workersCreatePhoneFlag     string
)

var workersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create EOR worker",
	Long:  "Create a new EOR worker. Requires --email, --first-name, --last-name, and --country flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if workersCreateEmailFlag == "" {
			f.PrintError("--email flag is required")
			return fmt.Errorf("--email flag is required")
		}
		if workersCreateFirstNameFlag == "" {
			f.PrintError("--first-name flag is required")
			return fmt.Errorf("--first-name flag is required")
		}
		if workersCreateLastNameFlag == "" {
			f.PrintError("--last-name flag is required")
			return fmt.Errorf("--last-name flag is required")
		}
		if workersCreateCountryFlag == "" {
			f.PrintError("--country flag is required")
			return fmt.Errorf("--country flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "EORWorker",
			Description: "Create EOR worker",
			Details: map[string]string{
				"Email":     workersCreateEmailFlag,
				"FirstName": workersCreateFirstNameFlag,
				"LastName":  workersCreateLastNameFlag,
				"Country":   workersCreateCountryFlag,
				"DOB":       workersCreateDOBFlag,
				"Phone":     workersCreatePhoneFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateEORWorkerParams{
			Email:       workersCreateEmailFlag,
			FirstName:   workersCreateFirstNameFlag,
			LastName:    workersCreateLastNameFlag,
			Country:     workersCreateCountryFlag,
			DateOfBirth: workersCreateDOBFlag,
			Phone:       workersCreatePhoneFlag,
		}

		worker, err := client.CreateEORWorker(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create EOR worker: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("EOR worker created successfully")
			f.PrintText("ID:         " + worker.ID)
			f.PrintText("Email:      " + worker.Email)
			f.PrintText("First Name: " + worker.FirstName)
			f.PrintText("Last Name:  " + worker.LastName)
			f.PrintText("Country:    " + worker.Country)
			if worker.DateOfBirth != "" {
				f.PrintText("DOB:        " + worker.DateOfBirth)
			}
			if worker.Phone != "" {
				f.PrintText("Phone:      " + worker.Phone)
			}
			f.PrintText("Status:     " + worker.Status)
			f.PrintText("Created:    " + worker.CreatedAt)
		}, worker)
	},
}

// Bank accounts subcommand
var bankAccountsCmd = &cobra.Command{
	Use:   "bank-accounts",
	Short: "Manage EOR worker bank accounts",
	Long:  "Add and manage bank accounts for EOR workers.",
}

// Flags for bank-accounts add command
var (
	bankAccountAddAccountHolderFlag string
	bankAccountAddBankNameFlag      string
	bankAccountAddAccountNumberFlag string
	bankAccountAddRoutingNumberFlag string
	bankAccountAddIBANFlag          string
	bankAccountAddSwiftFlag         string
	bankAccountAddCurrencyFlag      string
	bankAccountAddIsPrimaryFlag     bool
)

var bankAccountsAddCmd = &cobra.Command{
	Use:   "add <worker-id>",
	Short: "Add bank account to EOR worker",
	Long:  "Add a bank account to an EOR worker. Requires --account-holder, --bank-name, --account-number, and --currency flags.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if bankAccountAddAccountHolderFlag == "" {
			f.PrintError("--account-holder flag is required")
			return fmt.Errorf("--account-holder flag is required")
		}
		if bankAccountAddBankNameFlag == "" {
			f.PrintError("--bank-name flag is required")
			return fmt.Errorf("--bank-name flag is required")
		}
		if bankAccountAddAccountNumberFlag == "" {
			f.PrintError("--account-number flag is required")
			return fmt.Errorf("--account-number flag is required")
		}
		if bankAccountAddCurrencyFlag == "" {
			f.PrintError("--currency flag is required")
			return fmt.Errorf("--currency flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "EORBankAccount",
			Description: "Add EOR bank account",
			Details: map[string]string{
				"WorkerID":      args[0],
				"AccountHolder": bankAccountAddAccountHolderFlag,
				"BankName":      bankAccountAddBankNameFlag,
				"AccountNumber": bankAccountAddAccountNumberFlag,
				"RoutingNumber": bankAccountAddRoutingNumberFlag,
				"IBAN":          bankAccountAddIBANFlag,
				"SWIFT":         bankAccountAddSwiftFlag,
				"Currency":      bankAccountAddCurrencyFlag,
				"Primary":       fmt.Sprintf("%t", bankAccountAddIsPrimaryFlag),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.AddBankAccountParams{
			AccountHolder: bankAccountAddAccountHolderFlag,
			BankName:      bankAccountAddBankNameFlag,
			AccountNumber: bankAccountAddAccountNumberFlag,
			RoutingNumber: bankAccountAddRoutingNumberFlag,
			IBAN:          bankAccountAddIBANFlag,
			Swift:         bankAccountAddSwiftFlag,
			Currency:      bankAccountAddCurrencyFlag,
			IsPrimary:     bankAccountAddIsPrimaryFlag,
		}

		bankAccount, err := client.AddEORWorkerBankAccount(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to add bank account: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Bank account added successfully")
			f.PrintText("ID:             " + bankAccount.ID)
			f.PrintText("Account Holder: " + bankAccount.AccountHolder)
			f.PrintText("Bank Name:      " + bankAccount.BankName)
			f.PrintText("Account Number: " + bankAccount.AccountNumber)
			if bankAccount.RoutingNumber != "" {
				f.PrintText("Routing Number: " + bankAccount.RoutingNumber)
			}
			if bankAccount.IBAN != "" {
				f.PrintText("IBAN:           " + bankAccount.IBAN)
			}
			if bankAccount.Swift != "" {
				f.PrintText("SWIFT:          " + bankAccount.Swift)
			}
			f.PrintText("Currency:       " + bankAccount.Currency)
			f.PrintText(fmt.Sprintf("Is Primary:     %t", bankAccount.IsPrimary))
		}, bankAccount)
	},
}

func init() {
	// Create command flags
	eorCreateCmd.Flags().StringVar(&eorCreateTitleFlag, "title", "", "Contract title (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateWorkerEmailFlag, "worker-email", "", "Worker email (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateWorkerNameFlag, "worker-name", "", "Worker name (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateCountryFlag, "country", "", "Country code (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateStartDateFlag, "start-date", "", "Start date YYYY-MM-DD (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateSalaryFlag, "salary", "", "Annual salary (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateCurrencyFlag, "currency", "", "Currency code (required)")
	eorCreateCmd.Flags().StringVar(&eorCreatePayFrequencyFlag, "pay-frequency", "", "Pay frequency (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateJobTitleFlag, "job-title", "", "Job title (required)")
	eorCreateCmd.Flags().StringVar(&eorCreateSeniorityFlag, "seniority", "", "Seniority level (optional)")
	eorCreateCmd.Flags().StringVar(&eorCreateScopeFlag, "scope", "", "Scope of work (optional)")

	// Cancel command flags
	eorCancelCmd.Flags().StringVar(&eorCancelReasonFlag, "reason", "", "Cancellation reason (required)")

	// Amend command flags
	eorAmendCmd.Flags().StringVar(&eorAmendTypeFlag, "type", "", "Amendment type (required)")
	eorAmendCmd.Flags().StringVar(&eorAmendEffectiveDateFlag, "effective-date", "", "Effective date YYYY-MM-DD (required)")
	eorAmendCmd.Flags().StringVar(&eorAmendReasonFlag, "reason", "", "Amendment reason (required)")
	eorAmendCmd.Flags().StringVar(&eorAmendSalaryFlag, "salary", "", "New salary (optional)")
	eorAmendCmd.Flags().StringVar(&eorAmendJobTitleFlag, "job-title", "", "New job title (optional)")
	eorAmendCmd.Flags().StringVar(&eorAmendSeniorityFlag, "seniority", "", "New seniority level (optional)")
	eorAmendCmd.Flags().StringVar(&eorAmendScopeFlag, "scope", "", "New scope (optional)")

	// Terminate command flags
	eorTerminateCmd.Flags().StringVar(&eorTerminateReasonFlag, "reason", "", "Termination reason enum (required): TERMINATION, FOR_CAUSE, PERFORMANCE, etc.")
	eorTerminateCmd.Flags().StringVar(&eorTerminateReasonDetailFlag, "reason-detail", "", "Detailed reason description, 100-5000 chars (required)")
	eorTerminateCmd.Flags().BoolVar(&eorTerminateNotifiedFlag, "notified", false, "Has the employee been notified")
	eorTerminateCmd.Flags().BoolVar(&eorTerminateSensitiveFlag, "sensitive", false, "Mark as sensitive termination")
	eorTerminateCmd.Flags().StringVar(&eorTerminateSeveranceFlag, "severance", "", "Severance type: DAYS, WEEKS, MONTHS, or CASH")
	eorTerminateCmd.Flags().IntVar(&eorTerminatePTOFlag, "pto-days", 0, "Paid time off days used")
	eorTerminateCmd.Flags().IntVar(&eorTerminateUnpaidFlag, "unpaid-days", 0, "Unpaid time off days used")
	eorTerminateCmd.Flags().IntVar(&eorTerminateSickFlag, "sick-days", 0, "Sick leave days used")

	// Workers create command flags
	workersCreateCmd.Flags().StringVar(&workersCreateEmailFlag, "email", "", "Worker email (required)")
	workersCreateCmd.Flags().StringVar(&workersCreateFirstNameFlag, "first-name", "", "First name (required)")
	workersCreateCmd.Flags().StringVar(&workersCreateLastNameFlag, "last-name", "", "Last name (required)")
	workersCreateCmd.Flags().StringVar(&workersCreateCountryFlag, "country", "", "Country code (required)")
	workersCreateCmd.Flags().StringVar(&workersCreateDOBFlag, "date-of-birth", "", "Date of birth YYYY-MM-DD (optional)")
	workersCreateCmd.Flags().StringVar(&workersCreatePhoneFlag, "phone", "", "Phone number (optional)")

	// Bank accounts add command flags
	bankAccountsAddCmd.Flags().StringVar(&bankAccountAddAccountHolderFlag, "account-holder", "", "Account holder name (required)")
	bankAccountsAddCmd.Flags().StringVar(&bankAccountAddBankNameFlag, "bank-name", "", "Bank name (required)")
	bankAccountsAddCmd.Flags().StringVar(&bankAccountAddAccountNumberFlag, "account-number", "", "Account number (required)")
	bankAccountsAddCmd.Flags().StringVar(&bankAccountAddRoutingNumberFlag, "routing-number", "", "Routing number (optional)")
	bankAccountsAddCmd.Flags().StringVar(&bankAccountAddIBANFlag, "iban", "", "IBAN (optional)")
	bankAccountsAddCmd.Flags().StringVar(&bankAccountAddSwiftFlag, "swift", "", "SWIFT code (optional)")
	bankAccountsAddCmd.Flags().StringVar(&bankAccountAddCurrencyFlag, "currency", "", "Currency code (required)")
	bankAccountsAddCmd.Flags().BoolVar(&bankAccountAddIsPrimaryFlag, "is-primary", false, "Set as primary account (optional)")

	// Add subcommands to workers
	workersCmd.AddCommand(workersCreateCmd)

	// Add subcommands to bank-accounts
	bankAccountsCmd.AddCommand(bankAccountsAddCmd)

	// Amendments list command flags
	eorAmendmentsListCmd.Flags().IntVar(&eorAmendmentsLimitFlag, "limit", 100, "Maximum results")

	// Add subcommands to amendments
	eorAmendmentsCmd.AddCommand(eorAmendmentsListCmd)
	eorAmendmentsCmd.AddCommand(eorAmendmentsSignCmd)
	eorAmendmentsCmd.AddCommand(eorAmendmentsAcceptCmd)

	// Add subcommands to eor
	eorCmd.AddCommand(eorCreateCmd)
	eorCmd.AddCommand(eorGetCmd)
	eorCmd.AddCommand(eorSignCmd)
	eorCmd.AddCommand(eorCancelCmd)
	eorCmd.AddCommand(eorAmendCmd)
	eorCmd.AddCommand(eorAmendmentsCmd)
	eorCmd.AddCommand(eorTerminateCmd)
	eorCmd.AddCommand(workersCmd)
	eorCmd.AddCommand(bankAccountsCmd)
}
