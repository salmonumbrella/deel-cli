package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var calcCmd = &cobra.Command{
	Use:     "calc",
	Aliases: []string{"calculator"},
	Short:   "Salary and cost calculators",
	Long:    "Calculate employer costs, take-home pay, and view salary data.",
}

var (
	calcCountryFlag  string
	calcSalaryFlag   float64
	calcCurrencyFlag string
	calcRoleFlag     string
)

var calcCostCmd = &cobra.Command{
	Use:   "cost",
	Short: "Calculate employer cost",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if calcCountryFlag == "" || calcSalaryFlag == 0 || calcCurrencyFlag == "" {
			f.PrintError("Required: --country, --salary, --currency")
			return fmt.Errorf("missing required flags")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		result, err := client.CalculateCost(cmd.Context(), api.CalculateCostParams{
			Country:     calcCountryFlag,
			GrossSalary: calcSalaryFlag,
			Currency:    calcCurrencyFlag,
		})
		if err != nil {
			f.PrintError("Failed to calculate: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Country:      " + result.Country)
			f.PrintText(fmt.Sprintf("Gross Salary: %.2f %s", result.GrossSalary, result.Currency))
			f.PrintText(fmt.Sprintf("Employer Cost: %.2f %s", result.EmployerCost, result.Currency))
			f.PrintText(fmt.Sprintf("Taxes & Fees: %.2f %s", result.TaxesAndFees, result.Currency))
			f.PrintText(fmt.Sprintf("Total Cost:   %.2f %s", result.TotalCost, result.Currency))
		}, result)
	},
}

var calcTakeHomeCmd = &cobra.Command{
	Use:   "take-home",
	Short: "Calculate take-home pay",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if calcCountryFlag == "" || calcSalaryFlag == 0 || calcCurrencyFlag == "" {
			f.PrintError("Required: --country, --salary, --currency")
			return fmt.Errorf("missing required flags")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		result, err := client.CalculateTakeHome(cmd.Context(), api.CalculateTakeHomeParams{
			Country:     calcCountryFlag,
			GrossSalary: calcSalaryFlag,
			Currency:    calcCurrencyFlag,
		})
		if err != nil {
			f.PrintError("Failed to calculate: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Country:      " + result.Country)
			f.PrintText(fmt.Sprintf("Gross Salary: %.2f %s", result.GrossSalary, result.Currency))
			f.PrintText(fmt.Sprintf("Net Salary:   %.2f %s", result.NetSalary, result.Currency))
			f.PrintText(fmt.Sprintf("Tax Rate:     %.1f%%", result.TaxRate*100))
		}, result)
	},
}

var calcSalaryHistogramCmd = &cobra.Command{
	Use:   "salary-histogram",
	Short: "Get salary histogram for a role",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if calcRoleFlag == "" || calcCountryFlag == "" {
			f.PrintError("Required: --role, --country")
			return fmt.Errorf("missing required flags")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		result, err := client.GetSalaryHistogram(cmd.Context(), calcRoleFlag, calcCountryFlag)
		if err != nil {
			f.PrintError("Failed to get histogram: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Role:    " + result.Role)
			f.PrintText("Country: " + result.Country)
			f.PrintText(fmt.Sprintf("Min:     %.2f %s", result.Min, result.Currency))
			f.PrintText(fmt.Sprintf("25th:    %.2f %s", result.Percentile25, result.Currency))
			f.PrintText(fmt.Sprintf("Median:  %.2f %s", result.Median, result.Currency))
			f.PrintText(fmt.Sprintf("75th:    %.2f %s", result.Percentile75, result.Currency))
			f.PrintText(fmt.Sprintf("Max:     %.2f %s", result.Max, result.Currency))
		}, result)
	},
}

func init() {
	calcCostCmd.Flags().StringVar(&calcCountryFlag, "country", "", "Country code (required)")
	calcCostCmd.Flags().Float64Var(&calcSalaryFlag, "salary", 0, "Gross salary (required)")
	calcCostCmd.Flags().StringVar(&calcCurrencyFlag, "currency", "", "Currency code (required)")

	calcTakeHomeCmd.Flags().StringVar(&calcCountryFlag, "country", "", "Country code (required)")
	calcTakeHomeCmd.Flags().Float64Var(&calcSalaryFlag, "salary", 0, "Gross salary (required)")
	calcTakeHomeCmd.Flags().StringVar(&calcCurrencyFlag, "currency", "", "Currency code (required)")

	calcSalaryHistogramCmd.Flags().StringVar(&calcRoleFlag, "role", "", "Role/job title (required)")
	calcSalaryHistogramCmd.Flags().StringVar(&calcCountryFlag, "country", "", "Country code (required)")

	calcCmd.AddCommand(calcCostCmd)
	calcCmd.AddCommand(calcTakeHomeCmd)
	calcCmd.AddCommand(calcSalaryHistogramCmd)
}
