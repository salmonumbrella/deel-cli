package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var payoutsCmd = &cobra.Command{
	Use:   "payouts",
	Short: "Manage payouts",
	Long:  "Withdraw funds, manage auto-withdrawal settings, and view contractor balances.",
}

var (
	withdrawAmountFlag      float64
	withdrawCurrencyFlag    string
	withdrawDescriptionFlag string
)

var withdrawCmd = &cobra.Command{
	Use:   "withdraw",
	Short: "Withdraw funds",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if withdrawAmountFlag == 0 || withdrawCurrencyFlag == "" {
			f.PrintError("Required: --amount and --currency")
			return fmt.Errorf("missing required flags")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		withdrawal, err := client.WithdrawFunds(cmd.Context(), api.WithdrawFundsParams{
			Amount:      withdrawAmountFlag,
			Currency:    withdrawCurrencyFlag,
			Description: withdrawDescriptionFlag,
		})
		if err != nil {
			f.PrintError("Failed to withdraw funds: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Withdrawal initiated")
			f.PrintText("ID:          " + withdrawal.ID)
			f.PrintText(fmt.Sprintf("Amount:      %.2f %s", withdrawal.Amount, withdrawal.Currency))
			f.PrintText("Status:      " + withdrawal.Status)
			f.PrintText("Created:     " + withdrawal.CreatedAt)
			if withdrawal.Description != "" {
				f.PrintText("Description: " + withdrawal.Description)
			}
		}, withdrawal)
	},
}

var autoWithdrawalCmd = &cobra.Command{
	Use:   "auto-withdrawal",
	Short: "View auto-withdrawal settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		settings, err := client.GetAutoWithdrawal(cmd.Context())
		if err != nil {
			f.PrintError("Failed to get auto-withdrawal settings: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText(fmt.Sprintf("Enabled:    %t", settings.Enabled))
			if settings.Threshold > 0 {
				f.PrintText(fmt.Sprintf("Threshold:  %.2f %s", settings.Threshold, settings.Currency))
			}
			if settings.Schedule != "" {
				f.PrintText("Schedule:   " + settings.Schedule)
			}
		}, settings)
	},
}

var (
	autoWithdrawalSetEnabledFlag   bool
	autoWithdrawalSetThresholdFlag float64
	autoWithdrawalSetCurrencyFlag  string
	autoWithdrawalSetScheduleFlag  string
)

var autoWithdrawalSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update auto-withdrawal settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Check if at least one flag was provided
		if !cmd.Flags().Changed("enabled") && !cmd.Flags().Changed("threshold") &&
			!cmd.Flags().Changed("currency") && !cmd.Flags().Changed("schedule") {
			f.PrintError("At least one flag (--enabled, --threshold, --currency, or --schedule) must be provided")
			return fmt.Errorf("no update flags provided")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.SetAutoWithdrawalParams{}
		if cmd.Flags().Changed("enabled") {
			params.Enabled = autoWithdrawalSetEnabledFlag
		}
		if cmd.Flags().Changed("threshold") {
			params.Threshold = autoWithdrawalSetThresholdFlag
		}
		if cmd.Flags().Changed("currency") {
			params.Currency = autoWithdrawalSetCurrencyFlag
		}
		if cmd.Flags().Changed("schedule") {
			params.Schedule = autoWithdrawalSetScheduleFlag
		}

		settings, err := client.SetAutoWithdrawal(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to update auto-withdrawal settings: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Auto-withdrawal settings updated")
			f.PrintText(fmt.Sprintf("Enabled:    %t", settings.Enabled))
			if settings.Threshold > 0 {
				f.PrintText(fmt.Sprintf("Threshold:  %.2f %s", settings.Threshold, settings.Currency))
			}
			if settings.Schedule != "" {
				f.PrintText("Schedule:   " + settings.Schedule)
			}
		}, settings)
	},
}

var balancesCmd = &cobra.Command{
	Use:   "balances",
	Short: "List contractor balances",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		balances, err := client.ListContractorBalances(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list contractor balances: %v", err)
			return err
		}

		return f.Output(func() {
			if len(balances) == 0 {
				f.PrintText("No contractor balances found.")
				return
			}
			table := f.NewTable("CONTRACTOR ID", "NAME", "BALANCE", "PENDING", "UPDATED")
			for _, b := range balances {
				balance := fmt.Sprintf("%.2f %s", b.Balance, b.Currency)
				pending := ""
				if b.PendingAmount > 0 {
					pending = fmt.Sprintf("%.2f %s", b.PendingAmount, b.Currency)
				}
				table.AddRow(b.ContractorID, b.ContractorName, balance, pending, b.UpdatedAt)
			}
			table.Render()
		}, balances)
	},
}

func init() {
	withdrawCmd.Flags().Float64Var(&withdrawAmountFlag, "amount", 0, "Amount to withdraw (required)")
	withdrawCmd.Flags().StringVar(&withdrawCurrencyFlag, "currency", "", "Currency code (required)")
	withdrawCmd.Flags().StringVar(&withdrawDescriptionFlag, "description", "", "Description")

	autoWithdrawalSetCmd.Flags().BoolVar(&autoWithdrawalSetEnabledFlag, "enabled", false, "Enable auto-withdrawal")
	autoWithdrawalSetCmd.Flags().Float64Var(&autoWithdrawalSetThresholdFlag, "threshold", 0, "Withdrawal threshold")
	autoWithdrawalSetCmd.Flags().StringVar(&autoWithdrawalSetCurrencyFlag, "currency", "", "Currency code")
	autoWithdrawalSetCmd.Flags().StringVar(&autoWithdrawalSetScheduleFlag, "schedule", "", "Withdrawal schedule")

	autoWithdrawalCmd.AddCommand(autoWithdrawalSetCmd)

	payoutsCmd.AddCommand(withdrawCmd)
	payoutsCmd.AddCommand(autoWithdrawalCmd)
	payoutsCmd.AddCommand(balancesCmd)
}
