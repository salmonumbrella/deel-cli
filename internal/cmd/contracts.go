package cmd

import (
	"fmt"
	"strings"

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
	contractsLimitFlag    int
	contractsCursorFlag   string
	contractsStatusFlag   string
	contractsTypeFlag     string
	contractsAllFlag      bool
	contractsEntityIDFlag string
	contractsCountryFlag  string

	// Create command flags
	contractTitleFlag         string
	contractTypeFlag          string
	contractWorkerEmailFlag   string
	contractWorkerFirstFlag   string
	contractWorkerLastFlag    string
	contractCurrencyFlag      string
	contractRateFlag          float64
	contractCountryFlag       string
	contractJobTitleFlag      string
	contractScopeFlag         string
	contractStartDateFlag     string
	contractEndDateFlag       string
	contractPaymentCycleFlag  string
	contractSeniorityFlag     string
	contractSpecialClauseFlag string

	// Extended create command flags
	contractTemplateFlag     string
	contractLegalEntityFlag  string
	contractGroupFlag        string
	contractCycleEndFlag     int
	contractCycleEndTypeFlag string
	contractFrequencyFlag    string
	contractManagerFlag      string

	// Terminate command flags
	terminateReasonFlag    string
	terminateDateFlag      string
	terminateNotesFlag     string
	terminateImmediateFlag bool
	terminateTypeFlag      string
	terminateRehireFlag    string

	// Sign command flags
	signSignerFlag string

	// Invite command flags
	inviteEmailFlag   string
	inviteLocaleFlag  string
	inviteMessageFlag string
)

var contractsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List contracts (default: active)",
	Long:    "List contracts in your organization. Defaults to active contracts; use --status to query other statuses and --entity-id or --country to filter.",
	Example: "  deel contracts list --query '.data[].id' -o json\n  deel contracts list --entity-id le-123 --all\n  deel contracts list --country TW --all",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
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
				return HandleError(f, err, "listing contracts")
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

		if contractsEntityIDFlag != "" {
			filtered := make([]api.Contract, 0, len(allContracts))
			hasEntityIDs := false
			for _, c := range allContracts {
				if c.EntityID == "" {
					continue
				}
				hasEntityIDs = true
				if c.EntityID == contractsEntityIDFlag {
					filtered = append(filtered, c)
				}
			}
			if hasEntityIDs {
				allContracts = filtered
			} else {
				entities, err := client.ListLegalEntities(cmd.Context())
				if err != nil {
					return HandleError(f, err, "resolving legal entity")
				}
				var entityNameFilter string
				for _, e := range entities {
					if e.ID == contractsEntityIDFlag {
						entityNameFilter = e.Name
						break
					}
				}
				if entityNameFilter == "" {
					return fmt.Errorf("legal entity %s not found", contractsEntityIDFlag)
				}
				filtered = filtered[:0]
				for _, c := range allContracts {
					if c.Entity == entityNameFilter {
						filtered = append(filtered, c)
					}
				}
				allContracts = filtered
			}
		}

		if contractsCountryFlag != "" {
			filtered := make([]api.Contract, 0, len(allContracts))
			for _, c := range allContracts {
				if strings.EqualFold(c.Country, contractsCountryFlag) {
					filtered = append(filtered, c)
				}
			}
			allContracts = filtered
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
			table := f.NewTable("ID", "TITLE", "WORKER", "ENTITY", "ENTITY ID", "TYPE", "STATUS")
			for _, c := range allContracts {
				entityID := c.EntityID
				if entityID == "" {
					entityID = "-"
				}
				table.AddRow(c.ID, c.Title, c.WorkerName, c.Entity, entityID, c.Type, c.Status)
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
			return HandleError(f, err, "initializing client")
		}

		contract, err := client.GetContract(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting contract")
		}

		return f.Output(func() {
			f.PrintText("ID:           " + contract.ID)
			f.PrintText("Title:        " + contract.Title)
			f.PrintText("Type:         " + contract.Type)
			f.PrintText("Status:       " + contract.Status)
			f.PrintText("Worker:       " + contract.WorkerName)
			f.PrintText("Email:        " + contract.WorkerEmail)
			f.PrintText("Entity:       " + contract.Entity)
			if contract.EntityID != "" {
				f.PrintText("Entity ID:    " + contract.EntityID)
			}
			f.PrintText("Country:      " + contract.Country)
			f.PrintText(fmt.Sprintf("Compensation: %.2f %s", contract.CompensationAmount, contract.Currency))
			f.PrintText("Start Date:   " + contract.StartDate)
			if contract.EndDate != "" {
				f.PrintText("End Date:     " + contract.EndDate)
			}
			f.PrintText("URL:          https://app.deel.com/contract/" + contract.ID + "/contracts")
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
			return HandleError(f, err, "initializing client")
		}

		amendments, err := client.ListContractAmendments(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "listing contract amendments")
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
			return HandleError(f, err, "initializing client")
		}

		dates, err := client.GetContractPaymentDates(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting payment dates")
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
			Title:          contractTitleFlag,
			Type:           contractTypeFlag,
			WorkerEmail:    contractWorkerEmailFlag,
			WorkerFirst:    contractWorkerFirstFlag,
			WorkerLast:     contractWorkerLastFlag,
			Currency:       contractCurrencyFlag,
			Rate:           contractRateFlag,
			Country:        contractCountryFlag,
			JobTitle:       contractJobTitleFlag,
			ScopeOfWork:    contractScopeFlag,
			StartDate:      contractStartDateFlag,
			EndDate:        contractEndDateFlag,
			PaymentCycle:   contractPaymentCycleFlag,
			SeniorityLevel: contractSeniorityFlag,
			SpecialClause:  contractSpecialClauseFlag,
			TemplateID:     contractTemplateFlag,
			LegalEntityID:  contractLegalEntityFlag,
			GroupID:        contractGroupFlag,
			CycleEnd:       contractCycleEndFlag,
			CycleEndType:   contractCycleEndTypeFlag,
			Frequency:      contractFrequencyFlag,
			ManagerID:      contractManagerFlag,
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Contract",
			Description: "Create contract",
			Details: map[string]string{
				"Title":        contractTitleFlag,
				"Type":         contractTypeFlag,
				"WorkerEmail":  contractWorkerEmailFlag,
				"Currency":     contractCurrencyFlag,
				"Rate":         fmt.Sprintf("%.2f", contractRateFlag),
				"Country":      contractCountryFlag,
				"JobTitle":     contractJobTitleFlag,
				"StartDate":    contractStartDateFlag,
				"EndDate":      contractEndDateFlag,
				"Template":     contractTemplateFlag,
				"LegalEntity":  contractLegalEntityFlag,
				"Group":        contractGroupFlag,
				"Manager":      contractManagerFlag,
				"CycleEnd":     fmt.Sprintf("%d", contractCycleEndFlag),
				"CycleEndType": contractCycleEndTypeFlag,
				"Frequency":    contractFrequencyFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		contract, err := client.CreateContract(cmd.Context(), params)
		if err != nil {
			return HandleError(f, err, "creating contract")
		}

		f.PrintSuccess("Contract created successfully")
		f.PrintText("Contract ID: " + contract.ID)
		f.PrintText("Status: " + contract.Status)
		f.PrintText("URL: https://app.deel.com/contract/" + contract.ID + "/contracts")
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

		if signSignerFlag == "" {
			return HandleError(f, fmt.Errorf("--signer is required"), "validating input")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "SIGN",
			Resource:    "Contract",
			Description: "Sign contract",
			Details: map[string]string{
				"ID":     args[0],
				"Signer": signSignerFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		contract, err := client.SignContract(cmd.Context(), args[0], signSignerFlag)
		if err != nil {
			return HandleError(f, err, "signing contract")
		}

		f.PrintSuccess("Contract signed successfully")
		f.PrintText("Contract ID: " + contract.ID)
		f.PrintText("Status: " + contract.Status)
		return nil
	},
}

var contractsTerminateCmd = &cobra.Command{
	Use:   "terminate <contract-id>",
	Short: "Terminate or cancel a contract",
	Long: `Terminate or cancel a contract.

Use --immediate to cancel unsigned contracts (waiting_for_client_sign or waiting_for_contractor_sign).
Use --date for scheduled termination of active contracts.

Examples:
  # Cancel an unsigned contract
  deel contracts terminate abc123 --reason "Project has come to an end" --immediate

  # Schedule termination for an active contract
  deel contracts terminate abc123 --reason "Project has come to an end" --date 2026-02-01`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if terminateReasonFlag == "" {
			f.PrintError("--reason is required")
			f.PrintText("\nTo see available reasons, run:")
			f.PrintText("  deel contracts termination-reasons")
			return fmt.Errorf("reason is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		// Look up reason ID from name
		reasons, err := client.ListTerminationReasons(cmd.Context())
		if err != nil {
			return HandleError(f, err, "listing termination reasons")
		}

		var reasonID string
		reasonLower := strings.ToLower(terminateReasonFlag)
		for _, r := range reasons {
			if strings.ToLower(r.Name) == reasonLower || r.ID == terminateReasonFlag {
				reasonID = r.ID
				break
			}
		}
		if reasonID == "" {
			f.PrintError("Unknown termination reason: %s", terminateReasonFlag)
			f.PrintText("\nAvailable reasons:")
			for _, r := range reasons {
				f.PrintText("  • " + r.Name)
			}
			return fmt.Errorf("unknown reason")
		}

		params := api.TerminateContractParams{
			TerminationReasonID:          reasonID,
			TerminationReasonDescription: terminateNotesFlag,
			CompletionDate:               terminateDateFlag,
			TerminateNow:                 terminateImmediateFlag,
			TerminationType:              terminateTypeFlag,
			EligibleForRehire:            terminateRehireFlag,
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

		err = client.TerminateContract(cmd.Context(), args[0], params)
		if err != nil {
			return HandleError(f, err, "terminating contract")
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
	Use:   "termination-reasons",
	Short: "List available termination reasons",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		reasons, err := client.ListTerminationReasons(cmd.Context())
		if err != nil {
			return HandleError(f, err, "listing termination reasons")
		}

		return f.Output(func() {
			f.PrintText("Available termination reasons:")
			for _, reason := range reasons {
				if reason.Description != "" {
					f.PrintText("  • " + reason.Name + " - " + reason.Description)
				} else {
					f.PrintText("  • " + reason.Name)
				}
			}
			f.PrintText("\nTo terminate a contract:")
			f.PrintText("  deel contracts terminate <contract-id> --reason \"<reason-name>\"")
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
			return HandleError(f, err, "initializing client")
		}

		url, err := client.GetContractPDF(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting contract PDF")
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

		if inviteEmailFlag == "" {
			return HandleError(f, fmt.Errorf("--email is required"), "validating input")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.InviteWorkerParams{
			Email:   inviteEmailFlag,
			Locale:  inviteLocaleFlag,
			Message: inviteMessageFlag,
		}

		err = client.InviteWorker(cmd.Context(), args[0], params)
		if err != nil {
			return HandleError(f, err, "sending invitation")
		}

		f.PrintSuccess("Invitation email sent successfully")
		f.PrintText("Contract ID: " + args[0])
		f.PrintText("Sent to: " + inviteEmailFlag)
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
			return HandleError(f, err, "initializing client")
		}

		url, err := client.GetInviteLink(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting invite link")
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
			return HandleError(f, err, "initializing client")
		}

		templates, err := client.ListContractTemplates(cmd.Context())
		if err != nil {
			return HandleError(f, err, "listing contract templates")
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
	contractsListCmd.Flags().IntVar(&contractsLimitFlag, "limit", 100, "Maximum results")
	contractsListCmd.Flags().StringVar(&contractsCursorFlag, "cursor", "", "Pagination cursor")
	contractsListCmd.Flags().StringVar(&contractsStatusFlag, "status", "active", "Filter by status (default: active)")
	contractsListCmd.Flags().StringVar(&contractsTypeFlag, "type", "", "Filter by type")
	contractsListCmd.Flags().BoolVar(&contractsAllFlag, "all", false, "Fetch all pages")
	contractsListCmd.Flags().StringVar(&contractsEntityIDFlag, "entity-id", "", "Filter by legal entity ID (client-side)")
	contractsListCmd.Flags().StringVar(&contractsCountryFlag, "country", "", "Filter by worker country code (client-side)")

	// Create command flags
	contractsCreateCmd.Flags().StringVar(&contractTitleFlag, "title", "", "Contract title (required)")
	contractsCreateCmd.Flags().StringVar(&contractTypeFlag, "type", "", "Contract type: fixed_rate, pay_as_you_go, milestone, task_based (required)")
	contractsCreateCmd.Flags().StringVar(&contractWorkerEmailFlag, "worker-email", "", "Worker email address (required)")
	contractsCreateCmd.Flags().StringVar(&contractWorkerFirstFlag, "worker-first", "", "Worker first name")
	contractsCreateCmd.Flags().StringVar(&contractWorkerLastFlag, "worker-last", "", "Worker last name")
	contractsCreateCmd.Flags().StringVar(&contractCurrencyFlag, "currency", "", "Currency code (e.g., USD, EUR) (required)")
	contractsCreateCmd.Flags().Float64Var(&contractRateFlag, "rate", 0, "Compensation rate")
	contractsCreateCmd.Flags().StringVar(&contractCountryFlag, "country", "", "Country code (required)")
	contractsCreateCmd.Flags().StringVar(&contractJobTitleFlag, "job-title", "", "Job title")
	contractsCreateCmd.Flags().StringVar(&contractScopeFlag, "scope", "", "Scope of work")
	contractsCreateCmd.Flags().StringVar(&contractStartDateFlag, "start-date", "", "Start date (YYYY-MM-DD)")
	contractsCreateCmd.Flags().StringVar(&contractEndDateFlag, "end-date", "", "End date (YYYY-MM-DD)")
	contractsCreateCmd.Flags().StringVar(&contractPaymentCycleFlag, "payment-cycle", "", "Payment cycle: weekly, bi_weekly, monthly")

	// Extended create command flags
	contractsCreateCmd.Flags().StringVar(&contractTemplateFlag, "template", "", "Contract template ID")
	contractsCreateCmd.Flags().StringVar(&contractLegalEntityFlag, "legal-entity", "", "Legal entity ID")
	contractsCreateCmd.Flags().StringVar(&contractGroupFlag, "group", "", "Group/team ID")
	contractsCreateCmd.Flags().IntVar(&contractCycleEndFlag, "cycle-end", 0, "Payment cycle end day (e.g., 5 for 5th of month)")
	contractsCreateCmd.Flags().StringVar(&contractCycleEndTypeFlag, "cycle-end-type", "", "Payment cycle end type: DAY_OF_MONTH, DAY_OF_WEEK, DAY_OF_LAST_WEEK")
	contractsCreateCmd.Flags().StringVar(&contractFrequencyFlag, "frequency", "", "Payment frequency: monthly, weekly, biweekly, semimonthly")
	contractsCreateCmd.Flags().StringVar(&contractSeniorityFlag, "seniority", "", "Seniority level ID (e.g., junior, mid, senior)")
	contractsCreateCmd.Flags().StringVar(&contractSpecialClauseFlag, "special-clause", "", "Special clause text for contract")
	contractsCreateCmd.Flags().StringVar(&contractManagerFlag, "manager", "", "Manager ID for workplace information")

	// Sign command flags
	contractsSignCmd.Flags().StringVar(&signSignerFlag, "signer", "", "Full name of person signing on behalf of client (required)")

	// Invite command flags
	contractsInviteCmd.Flags().StringVar(&inviteEmailFlag, "email", "", "Worker email address (required)")
	contractsInviteCmd.Flags().StringVar(&inviteLocaleFlag, "locale", "en", "Locale for invitation (default: en)")
	contractsInviteCmd.Flags().StringVar(&inviteMessageFlag, "message", "", "Custom message for invitation")

	// Terminate command flags
	contractsTerminateCmd.Flags().StringVar(&terminateReasonFlag, "reason", "", "Termination reason (required)")
	contractsTerminateCmd.Flags().StringVar(&terminateDateFlag, "date", "", "Completion date for scheduled termination (YYYY-MM-DD)")
	contractsTerminateCmd.Flags().StringVar(&terminateNotesFlag, "notes", "", "Additional notes/description for the termination")
	contractsTerminateCmd.Flags().BoolVar(&terminateImmediateFlag, "immediate", false, "Terminate immediately (overrides --date)")
	contractsTerminateCmd.Flags().StringVar(&terminateTypeFlag, "type", "TERMINATION", "Termination type: RESIGNATION or TERMINATION")
	contractsTerminateCmd.Flags().StringVar(&terminateRehireFlag, "rehire", "", "Eligible for rehire: YES, NO, or DONT_KNOW")

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
