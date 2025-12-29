package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var contractsCmd = &cobra.Command{
	Use:   "contracts",
	Short: "Manage contracts",
	Long:  "List, view, and manage contracts in your Deel organization.",
}

var (
	contractsLimitFlag  int
	contractsCursorFlag string
	contractsStatusFlag string
	contractsTypeFlag   string
	contractsAllFlag    bool

	// Create command flags
	contractTitleFlag        string
	contractTypeFlag         string
	contractWorkerEmailFlag  string
	contractCurrencyFlag     string
	contractRateFlag         float64
	contractCountryFlag      string
	contractJobTitleFlag     string
	contractScopeFlag        string
	contractStartDateFlag    string
	contractEndDateFlag      string
	contractPaymentCycleFlag string

	// Terminate command flags
	terminateReasonFlag string
	terminateDateFlag   string
	terminateNotesFlag  string
)

var contractsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contracts",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "INVITE",
			Resource:    "Contract",
			Description: "Send invitation to worker",
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

		cursor := contractsCursorFlag
		var allContracts []api.Contract
		var next string

		for {
			resp, err := client.ListContracts(cmd.Context(), api.ContractsListParams{
				Limit:  contractsLimitFlag,
				Cursor: cursor,
				Status: contractsStatusFlag,
				Type:   contractsTypeFlag,
			})
			if err != nil {
				f.PrintError("Failed to list contracts: %v", err)
				return err
			}
			allContracts = append(allContracts, resp.Data...)
			next = resp.Page.Next
			if !contractsAllFlag || next == "" {
				if !contractsAllFlag {
					allContracts = resp.Data
				}
				break
			}
			cursor = next
		}

		response := api.ContractsListResponse{
			Data: allContracts,
		}
		response.Page.Next = ""

		return f.Output(func() {
			if len(allContracts) == 0 {
				f.PrintText("No contracts found.")
				return
			}
			table := f.NewTable("ID", "TITLE", "WORKER", "TYPE", "STATUS")
			for _, c := range allContracts {
				table.AddRow(c.ID, c.Title, c.WorkerName, c.Type, c.Status)
			}
			table.Render()
			if !contractsAllFlag && next != "" {
				f.PrintText("")
				f.PrintText("More results available. Use --cursor to paginate or --all to fetch everything.")
			}
		}, response)
	},
}

var contractsGetCmd = &cobra.Command{
	Use:   "get <contract-id>",
	Short: "Get contract details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		contract, err := client.GetContract(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get contract: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:           " + contract.ID)
			f.PrintText("Title:        " + contract.Title)
			f.PrintText("Type:         " + contract.Type)
			f.PrintText("Status:       " + contract.Status)
			f.PrintText("Worker:       " + contract.WorkerName)
			f.PrintText("Email:        " + contract.WorkerEmail)
			f.PrintText("Country:      " + contract.Country)
			f.PrintText(fmt.Sprintf("Compensation: %.2f %s", contract.CompensationAmount, contract.Currency))
			f.PrintText("Start Date:   " + contract.StartDate)
			if contract.EndDate != "" {
				f.PrintText("End Date:     " + contract.EndDate)
			}
		}, contract)
	},
}

var contractsAmendmentsCmd = &cobra.Command{
	Use:   "amendments <contract-id>",
	Short: "List contract amendments",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		amendments, err := client.ListContractAmendments(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to list amendments: %v", err)
			return err
		}

		return f.Output(func() {
			if len(amendments) == 0 {
				f.PrintText("No amendments found.")
				return
			}
			table := f.NewTable("ID", "TYPE", "STATUS", "CREATED")
			for _, a := range amendments {
				table.AddRow(a.ID, a.Type, a.Status, a.CreatedAt)
			}
			table.Render()
		}, amendments)
	},
}

var contractsPaymentDatesCmd = &cobra.Command{
	Use:   "payment-dates <contract-id>",
	Short: "Get contract payment dates",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		dates, err := client.GetContractPaymentDates(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get payment dates: %v", err)
			return err
		}

		return f.Output(func() {
			if len(dates) == 0 {
				f.PrintText("No payment dates found.")
				return
			}
			table := f.NewTable("DATE", "AMOUNT", "STATUS")
			for _, d := range dates {
				table.AddRow(d.Date, fmt.Sprintf("%.2f", d.Amount), d.Status)
			}
			table.Render()
		}, dates)
	},
}

var contractsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new contract",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Validate required fields
		if contractTitleFlag == "" {
			f.PrintError("--title is required")
			return fmt.Errorf("title is required")
		}
		if contractTypeFlag == "" {
			f.PrintError("--type is required (fixed_rate, pay_as_you_go, milestone, task_based)")
			return fmt.Errorf("type is required")
		}
		if contractWorkerEmailFlag == "" {
			f.PrintError("--worker-email is required")
			return fmt.Errorf("worker-email is required")
		}
		if contractCurrencyFlag == "" {
			f.PrintError("--currency is required")
			return fmt.Errorf("currency is required")
		}
		if contractCountryFlag == "" {
			f.PrintError("--country is required")
			return fmt.Errorf("country is required")
		}

		params := api.CreateContractParams{
			Title:        contractTitleFlag,
			Type:         contractTypeFlag,
			WorkerEmail:  contractWorkerEmailFlag,
			Currency:     contractCurrencyFlag,
			Rate:         contractRateFlag,
			Country:      contractCountryFlag,
			JobTitle:     contractJobTitleFlag,
			ScopeOfWork:  contractScopeFlag,
			StartDate:    contractStartDateFlag,
			EndDate:      contractEndDateFlag,
			PaymentCycle: contractPaymentCycleFlag,
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Contract",
			Description: "Create contract",
			Details: map[string]string{
				"Title":       contractTitleFlag,
				"Type":        contractTypeFlag,
				"WorkerEmail": contractWorkerEmailFlag,
				"Currency":    contractCurrencyFlag,
				"Rate":        fmt.Sprintf("%.2f", contractRateFlag),
				"Country":     contractCountryFlag,
				"JobTitle":    contractJobTitleFlag,
				"StartDate":   contractStartDateFlag,
				"EndDate":     contractEndDateFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		contract, err := client.CreateContract(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create contract: %v", err)
			return err
		}

		f.PrintSuccess("Contract created successfully")
		f.PrintText("Contract ID: " + contract.ID)
		f.PrintText("Status: " + contract.Status)
		f.PrintText("\nNext steps:")
		f.PrintText("  1. Sign the contract: deel contracts sign " + contract.ID)
		f.PrintText("  2. Invite worker: deel contracts invite " + contract.ID)
		return nil
	},
}

var contractsSignCmd = &cobra.Command{
	Use:   "sign <contract-id>",
	Short: "Sign a contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "SIGN",
			Resource:    "Contract",
			Description: "Sign contract",
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

		contract, err := client.SignContract(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to sign contract: %v", err)
			return err
		}

		f.PrintSuccess("Contract signed successfully")
		f.PrintText("Contract ID: " + contract.ID)
		f.PrintText("Status: " + contract.Status)
		return nil
	},
}

var contractsTerminateCmd = &cobra.Command{
	Use:   "terminate <contract-id>",
	Short: "Terminate a contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if terminateReasonFlag == "" {
			f.PrintError("--reason is required")
			f.PrintText("\nTo see available reasons, run:")
			f.PrintText("  deel contracts termination-reasons " + args[0])
			return fmt.Errorf("reason is required")
		}

		params := api.TerminateContractParams{
			Reason:        terminateReasonFlag,
			EffectiveDate: terminateDateFlag,
			Notes:         terminateNotesFlag,
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "TERMINATE",
			Resource:    "Contract",
			Description: "Terminate contract",
			Details: map[string]string{
				"ID":            args[0],
				"Reason":        terminateReasonFlag,
				"EffectiveDate": terminateDateFlag,
				"Notes":         terminateNotesFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		err = client.TerminateContract(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to terminate contract: %v", err)
			return err
		}

		f.PrintSuccess("Contract termination initiated successfully")
		f.PrintText("Contract ID: " + args[0])
		f.PrintText("Reason: " + terminateReasonFlag)
		if terminateDateFlag != "" {
			f.PrintText("Effective Date: " + terminateDateFlag)
		}
		return nil
	},
}

var contractsTerminationReasonsCmd = &cobra.Command{
	Use:   "termination-reasons <contract-id>",
	Short: "List available termination reasons",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		reasons, err := client.ListTerminationReasons(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to list termination reasons: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Available termination reasons:")
			for _, reason := range reasons {
				f.PrintText("  • " + reason)
			}
			f.PrintText("\nTo terminate this contract:")
			f.PrintText("  deel contracts terminate " + args[0] + " --reason \"<reason>\"")
		}, reasons)
	},
}

var contractsPDFCmd = &cobra.Command{
	Use:   "pdf <contract-id>",
	Short: "Get contract PDF download URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		url, err := client.GetContractPDF(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get PDF URL: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("PDF Download URL:")
			f.PrintText(url)
		}, map[string]string{"url": url})
	},
}

var contractsInviteCmd = &cobra.Command{
	Use:   "invite <contract-id>",
	Short: "Send invitation email to worker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		err = client.InviteWorker(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to send invitation: %v", err)
			return err
		}

		f.PrintSuccess("Invitation email sent successfully")
		f.PrintText("Contract ID: " + args[0])
		return nil
	},
}

var contractsInviteLinkCmd = &cobra.Command{
	Use:   "invite-link <contract-id>",
	Short: "Get invite link for worker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		url, err := client.GetInviteLink(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get invite link: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Invite Link:")
			f.PrintText(url)
		}, map[string]string{"url": url})
	},
}

var contractsTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "List available contract templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		templates, err := client.ListContractTemplates(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list templates: %v", err)
			return err
		}

		return f.Output(func() {
			if len(templates) == 0 {
				f.PrintText("No templates found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "DESCRIPTION")
			for _, t := range templates {
				table.AddRow(t.ID, t.Name, t.Type, t.Description)
			}
			table.Render()
		}, templates)
	},
}

func init() {
	// List command flags
	contractsListCmd.Flags().IntVar(&contractsLimitFlag, "limit", 50, "Maximum results")
	contractsListCmd.Flags().StringVar(&contractsCursorFlag, "cursor", "", "Pagination cursor")
	contractsListCmd.Flags().StringVar(&contractsStatusFlag, "status", "", "Filter by status")
	contractsListCmd.Flags().StringVar(&contractsTypeFlag, "type", "", "Filter by type")
	contractsListCmd.Flags().BoolVar(&contractsAllFlag, "all", false, "Fetch all pages")

	// Create command flags
	contractsCreateCmd.Flags().StringVar(&contractTitleFlag, "title", "", "Contract title (required)")
	contractsCreateCmd.Flags().StringVar(&contractTypeFlag, "type", "", "Contract type: fixed_rate, pay_as_you_go, milestone, task_based (required)")
	contractsCreateCmd.Flags().StringVar(&contractWorkerEmailFlag, "worker-email", "", "Worker email address (required)")
	contractsCreateCmd.Flags().StringVar(&contractCurrencyFlag, "currency", "", "Currency code (e.g., USD, EUR) (required)")
	contractsCreateCmd.Flags().Float64Var(&contractRateFlag, "rate", 0, "Compensation rate")
	contractsCreateCmd.Flags().StringVar(&contractCountryFlag, "country", "", "Country code (required)")
	contractsCreateCmd.Flags().StringVar(&contractJobTitleFlag, "job-title", "", "Job title")
	contractsCreateCmd.Flags().StringVar(&contractScopeFlag, "scope", "", "Scope of work")
	contractsCreateCmd.Flags().StringVar(&contractStartDateFlag, "start-date", "", "Start date (YYYY-MM-DD)")
	contractsCreateCmd.Flags().StringVar(&contractEndDateFlag, "end-date", "", "End date (YYYY-MM-DD)")
	contractsCreateCmd.Flags().StringVar(&contractPaymentCycleFlag, "payment-cycle", "", "Payment cycle: weekly, bi_weekly, monthly")

	// Terminate command flags
	contractsTerminateCmd.Flags().StringVar(&terminateReasonFlag, "reason", "", "Termination reason (required)")
	contractsTerminateCmd.Flags().StringVar(&terminateDateFlag, "date", "", "Effective termination date (YYYY-MM-DD)")
	contractsTerminateCmd.Flags().StringVar(&terminateNotesFlag, "notes", "", "Additional notes")

	// Add all commands
	contractsCmd.AddCommand(contractsListCmd)
	contractsCmd.AddCommand(contractsGetCmd)
	contractsCmd.AddCommand(contractsAmendmentsCmd)
	contractsCmd.AddCommand(contractsPaymentDatesCmd)
	contractsCmd.AddCommand(contractsCreateCmd)
	contractsCmd.AddCommand(contractsSignCmd)
	contractsCmd.AddCommand(contractsTerminateCmd)
	contractsCmd.AddCommand(contractsTerminationReasonsCmd)
	contractsCmd.AddCommand(contractsPDFCmd)
	contractsCmd.AddCommand(contractsInviteCmd)
	contractsCmd.AddCommand(contractsInviteLinkCmd)
	contractsCmd.AddCommand(contractsTemplatesCmd)
}
