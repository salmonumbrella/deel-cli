package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var complianceCmd = &cobra.Command{
	Use:   "compliance",
	Short: "Manage compliance documents",
	Long:  "View compliance documents, templates, and validation status.",
}

var (
	complianceContractFlag string
	complianceCountryFlag  string
)

var complianceDocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "List compliance documents for a contract",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if complianceContractFlag == "" {
			f.PrintError("--contract is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		docs, err := client.ListComplianceDocs(cmd.Context(), complianceContractFlag)
		if err != nil {
			f.PrintError("Failed to list documents: %v", err)
			return err
		}

		return f.Output(func() {
			if len(docs) == 0 {
				f.PrintText("No compliance documents found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "STATUS", "REQUIRED", "EXPIRES")
			for _, d := range docs {
				req := "No"
				if d.Required {
					req = "Yes"
				}
				table.AddRow(d.ID, d.Name, d.Type, d.Status, req, d.ExpiresAt)
			}
			table.Render()
		}, docs)
	},
}

var complianceTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "List compliance templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if complianceCountryFlag == "" {
			f.PrintError("--country is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		templates, err := client.ListComplianceTemplates(cmd.Context(), complianceCountryFlag)
		if err != nil {
			f.PrintError("Failed to list templates: %v", err)
			return err
		}

		return f.Output(func() {
			if len(templates) == 0 {
				f.PrintText("No templates found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "COUNTRY")
			for _, t := range templates {
				table.AddRow(t.ID, t.Name, t.Type, t.Country)
			}
			table.Render()
		}, templates)
	},
}

var complianceValidationsCmd = &cobra.Command{
	Use:   "validations",
	Short: "Get compliance validation status",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if complianceContractFlag == "" {
			f.PrintError("--contract is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		validation, err := client.GetComplianceValidations(cmd.Context(), complianceContractFlag)
		if err != nil {
			f.PrintError("Failed to get validations: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Contract:   " + validation.ContractID)
			f.PrintText("Status:     " + validation.Status)
			f.PrintText(fmt.Sprintf("Issues:     %d", validation.Issues))
			f.PrintText("Last Check: " + validation.LastCheck)
		}, validation)
	},
}

func init() {
	complianceDocsCmd.Flags().StringVar(&complianceContractFlag, "contract", "", "Contract ID (required)")
	complianceTemplatesCmd.Flags().StringVar(&complianceCountryFlag, "country", "", "Country code (required)")
	complianceValidationsCmd.Flags().StringVar(&complianceContractFlag, "contract", "", "Contract ID (required)")

	complianceCmd.AddCommand(complianceDocsCmd)
	complianceCmd.AddCommand(complianceTemplatesCmd)
	complianceCmd.AddCommand(complianceValidationsCmd)
}
