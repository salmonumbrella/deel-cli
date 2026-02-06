package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var payrollCmd = &cobra.Command{
	Use:   "payroll",
	Short: "Manage payroll and payslips",
	Long:  "View payslips, payment breakdowns, and receipts.",
}

var (
	payrollWorkerFlag string
	payrollGPFlag     bool
	payrollCycleFlag  string
	payrollLimitFlag  int
)

var payrollPayslipsCmd = &cobra.Command{
	Use:   "payslips",
	Short: "List payslips for a worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if payrollWorkerFlag == "" {
			return failValidation(cmd, f, "--worker is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		if payrollGPFlag {
			payslips, err := client.GetGPWorkerPayslips(cmd.Context(), payrollWorkerFlag)
			if err != nil {
				return HandleError(f, err, "get payslips")
			}
			if len(payslips) == 0 {
				f.PrintWarning("Hint: No GP payslips found. If this is an EOR employee, try without --gp flag.")
			}
			return f.OutputFiltered(cmd.Context(), func() {
				if len(payslips) == 0 {
					f.PrintText("No payslips found.")
					return
				}
				f.PrintText(fmt.Sprintf("Found %d payslips for worker %s\n", len(payslips), payrollWorkerFlag))
				table := f.NewTable("ID", "FROM", "TO", "STATUS")
				for _, ps := range payslips {
					table.AddRow(ps.ID, ps.From, ps.To, ps.Status)
				}
				table.Render()
			}, payslips)
		}

		payslips, err := client.GetEORWorkerPayslips(cmd.Context(), payrollWorkerFlag)
		if err != nil {
			return HandleError(f, err, "get payslips")
		}
		if len(payslips) == 0 {
			f.PrintWarning("Hint: No EOR payslips found. If this is a Global Payroll employee, try with --gp flag.")
		}
		return f.OutputFiltered(cmd.Context(), func() {
			if len(payslips) == 0 {
				f.PrintText("No payslips found.")
				return
			}
			f.PrintText(fmt.Sprintf("Found %d payslips for worker %s\n", len(payslips), payrollWorkerFlag))
			table := f.NewTable("ID", "FROM", "TO", "STATUS")
			for _, ps := range payslips {
				table.AddRow(ps.ID, ps.From, ps.To, ps.Status)
			}
			table.Render()
		}, payslips)
	},
}

var payrollPaymentsCmd = &cobra.Command{
	Use:   "payments",
	Short: "Get payment breakdown for a cycle",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if payrollCycleFlag == "" {
			return failValidation(cmd, f, "--cycle is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		breakdown, err := client.GetPaymentBreakdown(cmd.Context(), payrollCycleFlag)
		if err != nil {
			return HandleError(f, err, "get breakdown")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("Cycle:   " + breakdown.CycleID)
			f.PrintText(fmt.Sprintf("Total:   %.2f %s", breakdown.TotalAmount, breakdown.Currency))
			f.PrintText(fmt.Sprintf("Workers: %d", breakdown.Workers))
			f.PrintText("Status:  " + breakdown.Status)
		}, breakdown)
	},
}

var payrollReceiptsCmd = &cobra.Command{
	Use:   "receipts",
	Short: "List payment receipts",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		receipts, err := client.ListPaymentReceipts(cmd.Context(), payrollLimitFlag)
		if err != nil {
			return HandleError(f, err, "list receipts")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(receipts) == 0 {
				f.PrintText("No receipts found.")
				return
			}
			table := f.NewTable("ID", "AMOUNT", "DATE", "REFERENCE")
			for _, r := range receipts {
				table.AddRow(r.ID, fmt.Sprintf("%.2f %s", r.Amount, r.Currency), r.Date, r.Reference)
			}
			table.Render()
		}, receipts)
	},
}

var (
	payrollDownloadWorkerFlag  string
	payrollDownloadPayslipFlag string
)

var payrollDownloadCmd = &cobra.Command{
	Use:   "download-pdf",
	Short: "Get download URL for a GP payslip PDF",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if payrollDownloadWorkerFlag == "" {
			return failValidation(cmd, f, "--worker is required")
		}
		if payrollDownloadPayslipFlag == "" {
			return failValidation(cmd, f, "--payslip is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		url, err := client.GetGPPayslipDownloadURL(cmd.Context(), payrollDownloadWorkerFlag, payrollDownloadPayslipFlag)
		if err != nil {
			return HandleError(f, err, "get download URL")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText(url)
		}, map[string]string{"url": url})
	},
}

func init() {
	payrollPayslipsCmd.Flags().StringVar(&payrollWorkerFlag, "worker", "", "Worker ID (required)")
	payrollPayslipsCmd.Flags().BoolVar(&payrollGPFlag, "gp", false, "Use Global Payroll API")

	payrollPaymentsCmd.Flags().StringVar(&payrollCycleFlag, "cycle", "", "Payment cycle ID (required)")

	payrollReceiptsCmd.Flags().IntVar(&payrollLimitFlag, "limit", 100, "Maximum results")

	payrollDownloadCmd.Flags().StringVar(&payrollDownloadWorkerFlag, "worker", "", "Worker ID (required)")
	payrollDownloadCmd.Flags().StringVar(&payrollDownloadPayslipFlag, "payslip", "", "Payslip ID (required)")

	payrollCmd.AddCommand(payrollPayslipsCmd)
	payrollCmd.AddCommand(payrollPaymentsCmd)
	payrollCmd.AddCommand(payrollReceiptsCmd)
	payrollCmd.AddCommand(payrollDownloadCmd)
}
