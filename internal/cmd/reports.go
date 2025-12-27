package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Generate reports",
	Long:  "Generate various reports for payments, contracts, and more.",
}

var (
	paymentsReportStartDateFlag  string
	paymentsReportEndDateFlag    string
	paymentsReportContractFlag   string
	paymentsReportStatusFlag     string
)

var paymentsReportCmd = &cobra.Command{
	Use:   "payments",
	Short: "Get detailed payments report",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		report, err := client.GetDetailedPaymentsReport(cmd.Context(), api.DetailedPaymentsReportParams{
			StartDate:  paymentsReportStartDateFlag,
			EndDate:    paymentsReportEndDateFlag,
			ContractID: paymentsReportContractFlag,
			Status:     paymentsReportStatusFlag,
		})
		if err != nil {
			f.PrintError("Failed to get payments report: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Report ID:     " + report.ReportID)
			f.PrintText("Generated:     " + report.GeneratedAt)
			f.PrintText("Period:        " + report.StartDate + " to " + report.EndDate)
			f.PrintText(fmt.Sprintf("Total Amount:  %.2f %s", report.TotalAmount, report.Currency))
			f.PrintText("")
			if len(report.Payments) == 0 {
				f.PrintText("No payments found in this period.")
				return
			}
			f.PrintText(fmt.Sprintf("Payments: %d", len(report.Payments)))
			table := f.NewTable("PAYMENT ID", "WORKER", "AMOUNT", "TYPE", "STATUS", "DATE")
			for _, p := range report.Payments {
				amount := fmt.Sprintf("%.2f %s", p.Amount, p.Currency)
				table.AddRow(p.PaymentID, p.WorkerName, amount, p.Type, p.Status, p.PaymentDate)
			}
			table.Render()
		}, report)
	},
}

func init() {
	paymentsReportCmd.Flags().StringVar(&paymentsReportStartDateFlag, "start-date", "", "Start date (YYYY-MM-DD)")
	paymentsReportCmd.Flags().StringVar(&paymentsReportEndDateFlag, "end-date", "", "End date (YYYY-MM-DD)")
	paymentsReportCmd.Flags().StringVar(&paymentsReportContractFlag, "contract", "", "Filter by contract ID")
	paymentsReportCmd.Flags().StringVar(&paymentsReportStatusFlag, "status", "", "Filter by status")

	reportsCmd.AddCommand(paymentsReportCmd)
}
