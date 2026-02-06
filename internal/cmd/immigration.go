package cmd

import (
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var immigrationCmd = &cobra.Command{
	Use:     "immigration",
	Aliases: []string{"visa"},
	Short:   "Manage immigration and visas",
	Long:    "View immigration cases, documents, visa types, and requirements.",
}

var (
	immigrationCaseFlag     string
	immigrationCountryFlag  string
	immigrationFromFlag     string
	immigrationToFlag       string
	immigrationContractFlag string
	immigrationTypeFlag     string
	immigrationStartFlag    string
	immigrationDocNameFlag  string
	immigrationDocTypeFlag  string
	immigrationDocURLFlag   string
)

var immigrationCasesCmd = &cobra.Command{
	Use:   "cases <case-id>",
	Short: "Get immigration case details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		caseDetails, err := client.GetImmigrationCaseDetails(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "get case")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("ID:          " + caseDetails.ID)
			f.PrintText("Case Number: " + caseDetails.CaseNumber)
			f.PrintText("Worker:      " + caseDetails.WorkerName)
			f.PrintText("Type:        " + caseDetails.Type)
			f.PrintText("Status:      " + caseDetails.Status)
			f.PrintText("Country:     " + caseDetails.Country)
			f.PrintText("Start Date:  " + caseDetails.StartDate)
			if caseDetails.ExpiryDate != "" {
				f.PrintText("Expiry:      " + caseDetails.ExpiryDate)
			}
		}, caseDetails)
	},
}

var immigrationDocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "List immigration documents for a case",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if immigrationCaseFlag == "" {
			return failValidation(cmd, f, "--case is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		docs, err := client.ListImmigrationDocs(cmd.Context(), immigrationCaseFlag)
		if err != nil {
			return HandleError(f, err, "list documents")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(docs) == 0 {
				f.PrintText("No documents found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "STATUS", "EXPIRES")
			for _, d := range docs {
				table.AddRow(d.ID, d.Name, d.Type, d.Status, d.ExpiresAt)
			}
			table.Render()
		}, docs)
	},
}

var immigrationVisaTypesCmd = &cobra.Command{
	Use:   "visa-types",
	Short: "List visa types for a country",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if immigrationCountryFlag == "" {
			return failValidation(cmd, f, "--country is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		types, err := client.ListVisaTypes(cmd.Context(), immigrationCountryFlag)
		if err != nil {
			return HandleError(f, err, "list visa types")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(types) == 0 {
				f.PrintText("No visa types found.")
				return
			}
			table := f.NewTable("ID", "NAME", "CATEGORY", "DURATION")
			for _, t := range types {
				table.AddRow(t.ID, t.Name, t.Category, t.Duration)
			}
			table.Render()
		}, types)
	},
}

var immigrationCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check visa requirements between countries",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if immigrationFromFlag == "" || immigrationToFlag == "" {
			return failValidation(cmd, f, "--from and --to are required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		req, err := client.CheckVisaRequirement(cmd.Context(), immigrationFromFlag, immigrationToFlag)
		if err != nil {
			return HandleError(f, err, "check")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			required := "No"
			if req.Required {
				required = "Yes"
			}
			f.PrintText("Visa Required: " + required)
			if req.Type != "" {
				f.PrintText("Suggested Type: " + req.Type)
			}
			f.PrintText("Max Stay:       " + req.Duration)
			if req.Notes != "" {
				f.PrintText("Notes:          " + req.Notes)
			}
		}, req)
	},
}

var immigrationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new immigration case",
	Long:  "Create a new immigration case for a contract. Requires --contract-id, --type, --country, and --start-date flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if immigrationContractFlag == "" || immigrationTypeFlag == "" ||
			immigrationCountryFlag == "" || immigrationStartFlag == "" {
			return failValidation(cmd, f, "--contract-id, --type, --country, and --start-date are required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "ImmigrationCase",
			Description: "Create immigration case",
			Details: map[string]string{
				"ContractID": immigrationContractFlag,
				"Type":       immigrationTypeFlag,
				"Country":    immigrationCountryFlag,
				"StartDate":  immigrationStartFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		caseDetails, err := client.CreateImmigrationCase(cmd.Context(), api.CreateImmigrationCaseParams{
			ContractID: immigrationContractFlag,
			Type:       immigrationTypeFlag,
			Country:    immigrationCountryFlag,
			StartDate:  immigrationStartFlag,
		})
		if err != nil {
			return HandleError(f, err, "create immigration case")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Immigration case created successfully")
			f.PrintText("ID:          " + caseDetails.ID)
			f.PrintText("Case Number: " + caseDetails.CaseNumber)
			f.PrintText("Worker:      " + caseDetails.WorkerName)
			f.PrintText("Type:        " + caseDetails.Type)
			f.PrintText("Status:      " + caseDetails.Status)
			f.PrintText("Country:     " + caseDetails.Country)
			f.PrintText("Start Date:  " + caseDetails.StartDate)
			if caseDetails.ExpiryDate != "" {
				f.PrintText("Expiry:      " + caseDetails.ExpiryDate)
			}
		}, caseDetails)
	},
}

var immigrationUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a document to an immigration case",
	Long:  "Upload a document to an existing immigration case. Requires --case, --name, --type, and --doc-url flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if immigrationCaseFlag == "" || immigrationDocNameFlag == "" ||
			immigrationDocTypeFlag == "" || immigrationDocURLFlag == "" {
			return failValidation(cmd, f, "--case, --name, --type, and --doc-url are required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPLOAD",
			Resource:    "ImmigrationDocument",
			Description: "Upload immigration document",
			Details: map[string]string{
				"CaseID": immigrationCaseFlag,
				"Name":   immigrationDocNameFlag,
				"Type":   immigrationDocTypeFlag,
				"DocURL": immigrationDocURLFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		doc, err := client.UploadImmigrationDocument(cmd.Context(), immigrationCaseFlag, api.UploadImmigrationDocumentParams{
			Name:        immigrationDocNameFlag,
			Type:        immigrationDocTypeFlag,
			DocumentURL: immigrationDocURLFlag,
		})
		if err != nil {
			return HandleError(f, err, "upload document")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Document uploaded successfully")
			f.PrintText("ID:       " + doc.ID)
			f.PrintText("Name:     " + doc.Name)
			f.PrintText("Type:     " + doc.Type)
			f.PrintText("Status:   " + doc.Status)
			if doc.ExpiresAt != "" {
				f.PrintText("Expires:  " + doc.ExpiresAt)
			}
		}, doc)
	},
}

func init() {
	immigrationDocsCmd.Flags().StringVar(&immigrationCaseFlag, "case", "", "Case ID (required)")
	immigrationVisaTypesCmd.Flags().StringVar(&immigrationCountryFlag, "country", "", "Country code (required)")
	immigrationCheckCmd.Flags().StringVar(&immigrationFromFlag, "from", "", "Origin country code (required)")
	immigrationCheckCmd.Flags().StringVar(&immigrationToFlag, "to", "", "Destination country code (required)")

	// Create command flags
	immigrationCreateCmd.Flags().StringVar(&immigrationContractFlag, "contract-id", "", "Contract ID (required)")
	immigrationCreateCmd.Flags().StringVar(&immigrationTypeFlag, "type", "", "Immigration case type (required)")
	immigrationCreateCmd.Flags().StringVar(&immigrationCountryFlag, "country", "", "Country code (required)")
	immigrationCreateCmd.Flags().StringVar(&immigrationStartFlag, "start-date", "", "Start date (required)")

	// Upload command flags
	immigrationUploadCmd.Flags().StringVar(&immigrationCaseFlag, "case", "", "Case ID (required)")
	immigrationUploadCmd.Flags().StringVar(&immigrationDocNameFlag, "name", "", "Document name (required)")
	immigrationUploadCmd.Flags().StringVar(&immigrationDocTypeFlag, "type", "", "Document type (required)")
	immigrationUploadCmd.Flags().StringVar(&immigrationDocURLFlag, "doc-url", "", "Document URL (required)")

	immigrationCmd.AddCommand(immigrationCasesCmd)
	immigrationCmd.AddCommand(immigrationDocsCmd)
	immigrationCmd.AddCommand(immigrationVisaTypesCmd)
	immigrationCmd.AddCommand(immigrationCheckCmd)
	immigrationCmd.AddCommand(immigrationCreateCmd)
	immigrationCmd.AddCommand(immigrationUploadCmd)
}
