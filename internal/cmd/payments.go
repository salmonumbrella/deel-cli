package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var paymentsCmd = &cobra.Command{
	Use:   "payments",
	Short: "Manage payments",
	Long:  "Manage off-cycle payments and other payment operations.",
}

var offCycleCmd = &cobra.Command{
	Use:   "off-cycle",
	Short: "Manage off-cycle payments",
}

var (
	offCycleContractFlag string
	offCycleStatusFlag   string
	offCycleLimitFlag    int
	offCycleCursorFlag   string
)

var offCycleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List off-cycle payments",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListOffCyclePayments(cmd.Context(), api.OffCyclePaymentsListParams{
			ContractID: offCycleContractFlag,
			Status:     offCycleStatusFlag,
			Limit:      offCycleLimitFlag,
			Cursor:     offCycleCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list payments: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No off-cycle payments found.")
				return
			}
			table := f.NewTable("ID", "WORKER", "TYPE", "AMOUNT", "STATUS", "DATE")
			for _, p := range resp.Data {
				amount := fmt.Sprintf("%.2f %s", p.Amount, p.Currency)
				table.AddRow(p.ID, p.WorkerName, p.Type, amount, p.Status, p.PaymentDate)
			}
			table.Render()
		}, resp)
	},
}

var (
	offCycleCreateContractFlag    string
	offCycleCreateAmountFlag      float64
	offCycleCreateCurrencyFlag    string
	offCycleCreateTypeFlag        string
	offCycleCreateDescriptionFlag string
	offCycleCreateDateFlag        string
)

var offCycleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an off-cycle payment",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if offCycleCreateContractFlag == "" || offCycleCreateAmountFlag == 0 ||
			offCycleCreateCurrencyFlag == "" || offCycleCreateTypeFlag == "" ||
			offCycleCreateDateFlag == "" {
			f.PrintError("Required: --contract, --amount, --currency, --type, --date")
			return fmt.Errorf("missing required flags")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		payment, err := client.CreateOffCyclePayment(cmd.Context(), api.CreateOffCyclePaymentParams{
			ContractID:  offCycleCreateContractFlag,
			Amount:      offCycleCreateAmountFlag,
			Currency:    offCycleCreateCurrencyFlag,
			Type:        offCycleCreateTypeFlag,
			Description: offCycleCreateDescriptionFlag,
			PaymentDate: offCycleCreateDateFlag,
		})
		if err != nil {
			f.PrintError("Failed to create payment: %v", err)
			return err
		}

		f.PrintSuccess("Created off-cycle payment: %s", payment.ID)
		return nil
	},
}

var breakdownCmd = &cobra.Command{
	Use:   "breakdown <payment-id>",
	Short: "Get payment breakdown",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		breakdown, err := client.GetIndividualPaymentBreakdown(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get payment breakdown: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Payment ID:        " + breakdown.PaymentID)
			f.PrintText(fmt.Sprintf("Gross Amount:      %.2f %s", breakdown.GrossAmount, breakdown.Currency))
			f.PrintText(fmt.Sprintf("Net Amount:        %.2f %s", breakdown.NetAmount, breakdown.Currency))
			f.PrintText(fmt.Sprintf("Deel Fee:          %.2f %s", breakdown.DeelFee, breakdown.Currency))
			f.PrintText(fmt.Sprintf("Withholding Tax:   %.2f %s", breakdown.WithholdingTax, breakdown.Currency))
			f.PrintText(fmt.Sprintf("Other Deductions:  %.2f %s", breakdown.OtherDeductions, breakdown.Currency))
			f.PrintText(fmt.Sprintf("Reimbursements:    %.2f %s", breakdown.Reimbursements, breakdown.Currency))
			if len(breakdown.LineItems) > 0 {
				f.PrintText("\nLine Items:")
				table := f.NewTable("TYPE", "DESCRIPTION", "AMOUNT")
				for _, item := range breakdown.LineItems {
					amount := fmt.Sprintf("%.2f %s", item.Amount, breakdown.Currency)
					table.AddRow(item.Type, item.Description, amount)
				}
				table.Render()
			}
		}, breakdown)
	},
}

var (
	receiptsLimitFlag     int
	receiptsCursorFlag    string
	receiptsContractFlag  string
	receiptsPaymentFlag   string
)

var receiptsCmd = &cobra.Command{
	Use:   "receipts",
	Short: "List payment receipts",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListDetailedPaymentReceipts(cmd.Context(), api.DetailedPaymentReceiptsListParams{
			Limit:      receiptsLimitFlag,
			Cursor:     receiptsCursorFlag,
			ContractID: receiptsContractFlag,
			PaymentID:  receiptsPaymentFlag,
		})
		if err != nil {
			f.PrintError("Failed to list payment receipts: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No payment receipts found.")
				return
			}
			table := f.NewTable("ID", "PAYMENT ID", "WORKER", "AMOUNT", "ISSUE DATE")
			for _, r := range resp.Data {
				amount := fmt.Sprintf("%.2f %s", r.Amount, r.Currency)
				table.AddRow(r.ID, r.PaymentID, r.WorkerName, amount, r.IssueDate)
			}
			table.Render()
		}, resp)
	},
}

func init() {
	offCycleListCmd.Flags().StringVar(&offCycleContractFlag, "contract", "", "Filter by contract ID")
	offCycleListCmd.Flags().StringVar(&offCycleStatusFlag, "status", "", "Filter by status")
	offCycleListCmd.Flags().IntVar(&offCycleLimitFlag, "limit", 50, "Maximum results")
	offCycleListCmd.Flags().StringVar(&offCycleCursorFlag, "cursor", "", "Pagination cursor")

	offCycleCreateCmd.Flags().StringVar(&offCycleCreateContractFlag, "contract", "", "Contract ID (required)")
	offCycleCreateCmd.Flags().Float64Var(&offCycleCreateAmountFlag, "amount", 0, "Payment amount (required)")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateCurrencyFlag, "currency", "", "Currency code (required)")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateTypeFlag, "type", "", "Payment type (required)")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateDescriptionFlag, "description", "", "Description")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateDateFlag, "date", "", "Payment date YYYY-MM-DD (required)")

	receiptsCmd.Flags().IntVar(&receiptsLimitFlag, "limit", 50, "Maximum results")
	receiptsCmd.Flags().StringVar(&receiptsCursorFlag, "cursor", "", "Pagination cursor")
	receiptsCmd.Flags().StringVar(&receiptsContractFlag, "contract", "", "Filter by contract ID")
	receiptsCmd.Flags().StringVar(&receiptsPaymentFlag, "payment", "", "Filter by payment ID")

	offCycleCmd.AddCommand(offCycleListCmd)
	offCycleCmd.AddCommand(offCycleCreateCmd)

	paymentsCmd.AddCommand(offCycleCmd)
	paymentsCmd.AddCommand(breakdownCmd)
	paymentsCmd.AddCommand(receiptsCmd)
}
