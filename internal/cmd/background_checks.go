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

var (
	bgCheckCountryFlag  string
	bgCheckContractFlag string
	bgCheckLimitFlag    int
)

var bgCheckOptionsCmd = &cobra.Command{
	Use:   "options",
	Short: "List background check options",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if bgCheckCountryFlag == "" {
			return failValidation(cmd, f, "--country is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		options, err := client.ListBackgroundCheckOptions(cmd.Context(), bgCheckCountryFlag)
		if err != nil {
			return HandleError(f, err, "list options")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
			return failValidation(cmd, f, "--contract is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		checks, err := client.ListBackgroundChecksByContract(cmd.Context(), bgCheckContractFlag)
		if err != nil {
			return HandleError(f, err, "list checks")
		}

		// Apply client-side limit
		if bgCheckLimitFlag > 0 && len(checks) > bgCheckLimitFlag {
			checks = checks[:bgCheckLimitFlag]
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
	bgCheckListCmd.Flags().IntVar(&bgCheckLimitFlag, "limit", 100, "Maximum results")

	bgCheckCmd.AddCommand(bgCheckOptionsCmd)
	bgCheckCmd.AddCommand(bgCheckListCmd)
}
