package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
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
	offCycleAllFlag      bool
)

var offCycleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List off-cycle payments",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, f, err := initClient("listing payments")
		if err != nil {
			return err
		}

		payments, _, hasMore, err := collectCursorItems(cmd.Context(), offCycleAllFlag, offCycleCursorFlag, offCycleLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.OffCyclePayment], error) {
			resp, err := client.ListOffCyclePayments(ctx, api.OffCyclePaymentsListParams{
				ContractID: offCycleContractFlag,
				Status:     offCycleStatusFlag,
				Limit:      limit,
				Cursor:     cursor,
			})
			if err != nil {
				return CursorListResult[api.OffCyclePayment]{}, err
			}
			return CursorListResult[api.OffCyclePayment]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing payments")
		}

		response := api.OffCyclePaymentsListResponse{
			Data: payments,
		}
		response.Page.Next = ""

		return outputList(cmd, f, payments, hasMore, "No off-cycle payments found.", []string{"ID", "WORKER", "TYPE", "AMOUNT", "STATUS", "DATE"}, func(p api.OffCyclePayment) []string {
			amount := fmt.Sprintf("%.2f %s", p.Amount, p.Currency)
			return []string{p.ID, p.WorkerName, p.Type, amount, p.Status, p.PaymentDate}
		}, response)
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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "OffCyclePayment",
			Description: "Create off-cycle payment",
			Details: map[string]string{
				"ContractID":  offCycleCreateContractFlag,
				"Amount":      fmt.Sprintf("%.2f %s", offCycleCreateAmountFlag, offCycleCreateCurrencyFlag),
				"Type":        offCycleCreateTypeFlag,
				"PaymentDate": offCycleCreateDateFlag,
				"Description": offCycleCreateDescriptionFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "creating payment")
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
			return HandleError(f, err, "creating payment")
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
			return HandleError(f, err, "getting payment")
		}

		breakdown, err := client.GetIndividualPaymentBreakdown(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting payment")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
	receiptsLimitFlag    int
	receiptsCursorFlag   string
	receiptsContractFlag string
	receiptsPaymentFlag  string
	receiptsAllFlag      bool
)

var receiptsCmd = &cobra.Command{
	Use:   "receipts",
	Short: "List payment receipts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, f, err := initClient("listing payments")
		if err != nil {
			return err
		}

		receipts, _, hasMore, err := collectCursorItems(cmd.Context(), receiptsAllFlag, receiptsCursorFlag, receiptsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.DetailedPaymentReceipt], error) {
			resp, err := client.ListDetailedPaymentReceipts(ctx, api.DetailedPaymentReceiptsListParams{
				Limit:      limit,
				Cursor:     cursor,
				ContractID: receiptsContractFlag,
				PaymentID:  receiptsPaymentFlag,
			})
			if err != nil {
				return CursorListResult[api.DetailedPaymentReceipt]{}, err
			}
			return CursorListResult[api.DetailedPaymentReceipt]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing payments")
		}

		response := api.DetailedPaymentReceiptsListResponse{
			Data: receipts,
		}
		response.Page.Next = ""

		return outputList(cmd, f, receipts, hasMore, "No payment receipts found.", []string{"ID", "PAYMENT ID", "WORKER", "AMOUNT", "ISSUE DATE"}, func(r api.DetailedPaymentReceipt) []string {
			amount := fmt.Sprintf("%.2f %s", r.Amount, r.Currency)
			return []string{r.ID, r.PaymentID, r.WorkerName, amount, r.IssueDate}
		}, response)
	},
}

func init() {
	offCycleListCmd.Flags().StringVar(&offCycleContractFlag, "contract", "", "Filter by contract ID")
	offCycleListCmd.Flags().StringVar(&offCycleStatusFlag, "status", "", "Filter by status")
	offCycleListCmd.Flags().IntVar(&offCycleLimitFlag, "limit", 100, "Maximum results")
	offCycleListCmd.Flags().StringVar(&offCycleCursorFlag, "cursor", "", "Pagination cursor")
	offCycleListCmd.Flags().BoolVar(&offCycleAllFlag, "all", false, "Fetch all pages")

	offCycleCreateCmd.Flags().StringVar(&offCycleCreateContractFlag, "contract", "", "Contract ID (required)")
	offCycleCreateCmd.Flags().Float64Var(&offCycleCreateAmountFlag, "amount", 0, "Payment amount (required)")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateCurrencyFlag, "currency", "", "Currency code (required)")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateTypeFlag, "type", "", "Payment type (required)")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateDescriptionFlag, "description", "", "Description")
	offCycleCreateCmd.Flags().StringVar(&offCycleCreateDateFlag, "date", "", "Payment date YYYY-MM-DD (required)")

	receiptsCmd.Flags().IntVar(&receiptsLimitFlag, "limit", 100, "Maximum results")
	receiptsCmd.Flags().StringVar(&receiptsCursorFlag, "cursor", "", "Pagination cursor")
	receiptsCmd.Flags().StringVar(&receiptsContractFlag, "contract", "", "Filter by contract ID")
	receiptsCmd.Flags().StringVar(&receiptsPaymentFlag, "payment", "", "Filter by payment ID")
	receiptsCmd.Flags().BoolVar(&receiptsAllFlag, "all", false, "Fetch all pages")

	offCycleCmd.AddCommand(offCycleListCmd)
	offCycleCmd.AddCommand(offCycleCreateCmd)

	paymentsCmd.AddCommand(offCycleCmd)
	paymentsCmd.AddCommand(breakdownCmd)
	paymentsCmd.AddCommand(receiptsCmd)
}
