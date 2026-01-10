package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var benefitsCmd = &cobra.Command{
	Use:   "benefits",
	Short: "Manage employee benefits",
	Long:  "View available benefits by country and employee benefit enrollments.",
}

var (
	benefitsCountryFlag  string
	benefitsEmployeeFlag string
	benefitsLimitFlag    int
)

var benefitsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List benefits by country",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if benefitsCountryFlag == "" {
			f.PrintError("--country is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		benefits, err := client.ListBenefitsByCountry(cmd.Context(), benefitsCountryFlag)
		if err != nil {
			return HandleError(f, err, "list benefits")
		}

		// Apply client-side limit
		if benefitsLimitFlag > 0 && len(benefits) > benefitsLimitFlag {
			benefits = benefits[:benefitsLimitFlag]
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(benefits) == 0 {
				f.PrintText("No benefits found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "PROVIDER", "COST")
			for _, b := range benefits {
				cost := fmt.Sprintf("%.2f %s", b.Cost, b.Currency)
				table.AddRow(b.ID, b.Name, b.Type, b.Provider, cost)
			}
			table.Render()
		}, benefits)
	},
}

var benefitsEmployeeCmd = &cobra.Command{
	Use:   "employee",
	Short: "List benefits for an employee",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if benefitsEmployeeFlag == "" {
			f.PrintError("--employee is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		benefits, err := client.GetEmployeeBenefits(cmd.Context(), benefitsEmployeeFlag)
		if err != nil {
			return HandleError(f, err, "get benefits")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(benefits) == 0 {
				f.PrintText("No benefits found.")
				return
			}
			table := f.NewTable("ID", "BENEFIT", "STATUS", "ENROLLED", "COST")
			for _, b := range benefits {
				cost := fmt.Sprintf("%.2f %s", b.Cost, b.Currency)
				table.AddRow(b.ID, b.BenefitName, b.Status, b.EnrolledDate, cost)
			}
			table.Render()
		}, benefits)
	},
}

func init() {
	benefitsListCmd.Flags().StringVar(&benefitsCountryFlag, "country", "", "Country code (required)")
	benefitsListCmd.Flags().IntVar(&benefitsLimitFlag, "limit", 100, "Maximum results")
	benefitsEmployeeCmd.Flags().StringVar(&benefitsEmployeeFlag, "employee", "", "Employee ID (required)")

	benefitsCmd.AddCommand(benefitsListCmd)
	benefitsCmd.AddCommand(benefitsEmployeeCmd)
}
