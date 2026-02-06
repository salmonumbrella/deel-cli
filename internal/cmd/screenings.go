package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var screeningsCmd = &cobra.Command{
	Use:   "screenings",
	Short: "Manage identity verification and screenings",
	Long:  "Manage KYC, AML, and identity verification through Veriff and external providers.",
}

var (
	screeningWorkerIDFlag   string
	screeningCallbackFlag   string
	screeningProviderFlag   string
	screeningVerifiedAtFlag string
	screeningDocTypeFlag    string
	screeningDocIDFlag      string
	screeningExpiryFlag     string
	screeningVerifiedByFlag string
	screeningNotesFlag      string
	screeningDocURLsFlag    []string
)

var screeningsVeriffCmd = &cobra.Command{
	Use:   "veriff",
	Short: "Create a Veriff verification session",
	Long:  "Create a new Veriff identity verification session for a worker. Requires --worker-id flag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if screeningWorkerIDFlag == "" {
			return failValidation(cmd, f, "--worker-id is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "VeriffSession",
			Description: "Create Veriff session",
			Details: map[string]string{
				"WorkerID": screeningWorkerIDFlag,
				"Callback": screeningCallbackFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		session, err := client.CreateVeriffSession(cmd.Context(), api.CreateVeriffSessionParams{
			WorkerID: screeningWorkerIDFlag,
			Callback: screeningCallbackFlag,
		})
		if err != nil {
			return HandleError(f, err, "create Veriff session")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Veriff session created successfully")
			f.PrintText("ID:       " + session.ID)
			f.PrintText("URL:      " + session.URL)
			f.PrintText("Status:   " + session.Status)
			f.PrintText("Expires:  " + session.ExpiresAt)
		}, session)
	},
}

var screeningsKYCCmd = &cobra.Command{
	Use:   "kyc <worker-id>",
	Short: "Get KYC verification details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		kyc, err := client.GetKYCDetails(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "get KYC details")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("Worker ID:  " + kyc.WorkerID)
			f.PrintText("Status:     " + kyc.Status)
			if kyc.VerifiedAt != "" {
				f.PrintText("Verified:   " + kyc.VerifiedAt)
			}
			if kyc.Provider != "" {
				f.PrintText("Provider:   " + kyc.Provider)
			}
			if kyc.Details.FirstName != "" {
				f.PrintText("Name:       " + kyc.Details.FirstName + " " + kyc.Details.LastName)
			}
			if kyc.Details.DateOfBirth != "" {
				f.PrintText("DOB:        " + kyc.Details.DateOfBirth)
			}
			if kyc.Details.Country != "" {
				f.PrintText("Country:    " + kyc.Details.Country)
			}
		}, kyc)
	},
}

var screeningsAMLCmd = &cobra.Command{
	Use:   "aml",
	Short: "Get AML screening data",
	Long:  "Retrieve anti-money laundering screening data for your organization.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		aml, err := client.GetAMLData(cmd.Context())
		if err != nil {
			return HandleError(f, err, "get AML data")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText(fmt.Sprintf("Total Matches: %d", aml.Summary.TotalMatches))
			f.PrintText("Highest Risk:  " + aml.Summary.HighestRisk)
			f.PrintText("")

			if len(aml.Results) == 0 {
				f.PrintText("No AML matches found.")
				return
			}

			table := f.NewTable("NAME", "COUNTRY", "MATCH TYPE", "RISK", "SCREENED")
			for _, r := range aml.Results {
				table.AddRow(r.Name, r.Country, r.MatchType, r.RiskLevel, r.ScreenedAt)
			}
			table.Render()
		}, aml)
	},
}

var screeningsExternalKYCCmd = &cobra.Command{
	Use:   "submit-kyc",
	Short: "Submit external KYC verification",
	Long:  "Submit KYC verification data from an external provider. Requires --worker-id, --provider, --verified-at, and --doc-type flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if screeningWorkerIDFlag == "" || screeningProviderFlag == "" ||
			screeningVerifiedAtFlag == "" || screeningDocTypeFlag == "" {
			return failValidation(cmd, f, "--worker-id, --provider, --verified-at, and --doc-type are required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "SUBMIT",
			Resource:    "ExternalKYC",
			Description: "Submit external KYC",
			Details: map[string]string{
				"WorkerID":     screeningWorkerIDFlag,
				"Provider":     screeningProviderFlag,
				"VerifiedAt":   screeningVerifiedAtFlag,
				"DocumentType": screeningDocTypeFlag,
				"DocumentID":   screeningDocIDFlag,
				"ExpiresAt":    screeningExpiryFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		submission, err := client.SubmitExternalKYC(cmd.Context(), api.SubmitExternalKYCParams{
			WorkerID:       screeningWorkerIDFlag,
			Provider:       screeningProviderFlag,
			VerifiedAt:     screeningVerifiedAtFlag,
			DocumentType:   screeningDocTypeFlag,
			DocumentID:     screeningDocIDFlag,
			ExpirationDate: screeningExpiryFlag,
		})
		if err != nil {
			return HandleError(f, err, "submit external KYC")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("External KYC submitted successfully")
			f.PrintText("ID:         " + submission.ID)
			f.PrintText("Worker ID:  " + submission.WorkerID)
			f.PrintText("Status:     " + submission.Status)
			f.PrintText("Submitted:  " + submission.SubmittedAt)
		}, submission)
	},
}

var screeningsManualVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Create manual verification record",
	Long:  "Create a manual verification record for a worker. Requires --worker-id and --verified-by flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if screeningWorkerIDFlag == "" || screeningVerifiedByFlag == "" {
			return failValidation(cmd, f, "--worker-id and --verified-by are required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "ManualVerification",
			Description: "Create manual verification",
			Details: map[string]string{
				"WorkerID":   screeningWorkerIDFlag,
				"VerifiedBy": screeningVerifiedByFlag,
				"Notes":      screeningNotesFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		verification, err := client.CreateManualVerification(cmd.Context(), api.CreateManualVerificationParams{
			WorkerID:     screeningWorkerIDFlag,
			VerifiedBy:   screeningVerifiedByFlag,
			Notes:        screeningNotesFlag,
			DocumentURLs: screeningDocURLsFlag,
		})
		if err != nil {
			return HandleError(f, err, "create manual verification")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Manual verification created successfully")
			f.PrintText("ID:          " + verification.ID)
			f.PrintText("Worker ID:   " + verification.WorkerID)
			f.PrintText("Verified By: " + verification.VerifiedBy)
			f.PrintText("Status:      " + verification.Status)
			if verification.Notes != "" {
				f.PrintText("Notes:       " + verification.Notes)
			}
			f.PrintText("Created:     " + verification.CreatedAt)
		}, verification)
	},
}

func init() {
	// Veriff command flags
	screeningsVeriffCmd.Flags().StringVar(&screeningWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	screeningsVeriffCmd.Flags().StringVar(&screeningCallbackFlag, "callback", "", "Callback URL (optional)")

	// External KYC command flags
	screeningsExternalKYCCmd.Flags().StringVar(&screeningWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	screeningsExternalKYCCmd.Flags().StringVar(&screeningProviderFlag, "provider", "", "Provider name (required)")
	screeningsExternalKYCCmd.Flags().StringVar(&screeningVerifiedAtFlag, "verified-at", "", "Verification date (required)")
	screeningsExternalKYCCmd.Flags().StringVar(&screeningDocTypeFlag, "doc-type", "", "Document type (required)")
	screeningsExternalKYCCmd.Flags().StringVar(&screeningDocIDFlag, "doc-id", "", "Document ID (optional)")
	screeningsExternalKYCCmd.Flags().StringVar(&screeningExpiryFlag, "expiry", "", "Document expiration date (optional)")

	// Manual verification command flags
	screeningsManualVerifyCmd.Flags().StringVar(&screeningWorkerIDFlag, "worker-id", "", "Worker ID (required)")
	screeningsManualVerifyCmd.Flags().StringVar(&screeningVerifiedByFlag, "verified-by", "", "Name of verifier (required)")
	screeningsManualVerifyCmd.Flags().StringVar(&screeningNotesFlag, "notes", "", "Verification notes (optional)")
	screeningsManualVerifyCmd.Flags().StringSliceVar(&screeningDocURLsFlag, "doc-urls", []string{}, "Document URLs (optional, can be specified multiple times)")

	// Add subcommands
	screeningsCmd.AddCommand(screeningsVeriffCmd)
	screeningsCmd.AddCommand(screeningsKYCCmd)
	screeningsCmd.AddCommand(screeningsAMLCmd)
	screeningsCmd.AddCommand(screeningsExternalKYCCmd)
	screeningsCmd.AddCommand(screeningsManualVerifyCmd)
}
