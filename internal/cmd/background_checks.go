package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bgCheckCmd = &cobra.Command{
	Use:     "background-checks",
	Aliases: []string{"bgcheck", "checks"},
	Short:   "Manage background checks",
	Long:    "View background check options and status.",
}

var bgCheckCountryFlag string
var bgCheckContractFlag string

var bgCheckOptionsCmd = &cobra.Command{
	Use:   "options",
	Short: "List background check options",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if bgCheckCountryFlag == "" {
			f.PrintError("--country is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		options, err := client.ListBackgroundCheckOptions(cmd.Context(), bgCheckCountryFlag)
		if err != nil {
			f.PrintError("Failed to list options: %v", err)
			return err
		}

		return f.Output(func() {
			if len(options) == 0 {
				f.PrintText("No options found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "PROVIDER", "COST", "DURATION")
			for _, o := range options {
				cost := fmt.Sprintf("%.2f %s", o.Cost, o.Currency)
				table.AddRow(o.ID, o.Name, o.Type, o.Provider, cost, o.Duration)
			}
			table.Render()
		}, options)
	},
}

var bgCheckListCmd = &cobra.Command{
	Use:   "list",
	Short: "List background checks for a contract",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if bgCheckContractFlag == "" {
			f.PrintError("--contract is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		checks, err := client.ListBackgroundChecksByContract(cmd.Context(), bgCheckContractFlag)
		if err != nil {
			f.PrintError("Failed to list checks: %v", err)
			return err
		}

		return f.Output(func() {
			if len(checks) == 0 {
				f.PrintText("No background checks found.")
				return
			}
			table := f.NewTable("ID", "WORKER", "TYPE", "STATUS", "RESULT")
			for _, c := range checks {
				table.AddRow(c.ID, c.WorkerName, c.Type, c.Status, c.Result)
			}
			table.Render()
		}, checks)
	},
}

func init() {
	bgCheckOptionsCmd.Flags().StringVar(&bgCheckCountryFlag, "country", "", "Country code (required)")
	bgCheckListCmd.Flags().StringVar(&bgCheckContractFlag, "contract", "", "Contract ID (required)")

	bgCheckCmd.AddCommand(bgCheckOptionsCmd)
	bgCheckCmd.AddCommand(bgCheckListCmd)
}
