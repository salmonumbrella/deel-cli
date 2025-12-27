package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var invoicesCmd = &cobra.Command{
	Use:   "invoices",
	Short: "Manage invoices",
	Long:  "List, view, and manage invoices and adjustments.",
}

var (
	invoicesLimitFlag    int
	invoicesCursorFlag   string
	invoicesStatusFlag   string
	invoicesContractFlag string
)

var invoicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all invoices",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListInvoices(cmd.Context(), api.InvoicesListParams{
			Limit:      invoicesLimitFlag,
			Cursor:     invoicesCursorFlag,
			Status:     invoicesStatusFlag,
			ContractID: invoicesContractFlag,
		})
		if err != nil {
			f.PrintError("Failed to list invoices: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No invoices found.")
				return
			}
			table := f.NewTable("ID", "NUMBER", "WORKER", "AMOUNT", "STATUS", "DUE DATE")
			for _, inv := range resp.Data {
				amount := fmt.Sprintf("%.2f %s", inv.Amount, inv.Currency)
				table.AddRow(inv.ID, inv.Number, inv.WorkerName, amount, inv.Status, inv.DueDate)
			}
			table.Render()
		}, resp)
	},
}

var invoicesGetCmd = &cobra.Command{
	Use:   "get <invoice-id>",
	Short: "Get invoice details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		invoice, err := client.GetInvoice(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get invoice: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:          " + invoice.ID)
			f.PrintText("Number:      " + invoice.Number)
			f.PrintText("Status:      " + invoice.Status)
			f.PrintText(fmt.Sprintf("Amount:      %.2f %s", invoice.Amount, invoice.Currency))
			f.PrintText("Worker:      " + invoice.WorkerName)
			f.PrintText("Contract:    " + invoice.ContractID)
			f.PrintText("Due Date:    " + invoice.DueDate)
			if invoice.PaidDate != "" {
				f.PrintText("Paid Date:   " + invoice.PaidDate)
			}
			if invoice.Description != "" {
				f.PrintText("Description: " + invoice.Description)
			}
		}, invoice)
	},
}

var invoicesAdjustmentsCmd = &cobra.Command{
	Use:   "adjustments <invoice-id>",
	Short: "List invoice adjustments",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		adjustments, err := client.ListInvoiceAdjustments(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to list adjustments: %v", err)
			return err
		}

		return f.Output(func() {
			if len(adjustments) == 0 {
				f.PrintText("No adjustments found.")
				return
			}
			table := f.NewTable("ID", "TYPE", "AMOUNT", "STATUS", "CREATED")
			for _, a := range adjustments {
				amount := fmt.Sprintf("%.2f %s", a.Amount, a.Currency)
				table.AddRow(a.ID, a.Type, amount, a.Status, a.CreatedAt)
			}
			table.Render()
		}, adjustments)
	},
}

var invoicesPDFCmd = &cobra.Command{
	Use:   "pdf <invoice-id>",
	Short: "Download invoice PDF",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		pdfBytes, err := client.GetInvoicePDF(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get invoice PDF: %v", err)
			return err
		}

		filename := fmt.Sprintf("invoice-%s.pdf", args[0])
		if err := os.WriteFile(filename, pdfBytes, 0644); err != nil {
			f.PrintError("Failed to save PDF: %v", err)
			return err
		}

		f.PrintSuccess("Saved invoice to %s", filename)
		return nil
	},
}

var (
	deelInvoicesLimitFlag  int
	deelInvoicesCursorFlag string
	deelInvoicesStatusFlag string
)

var deelInvoicesCmd = &cobra.Command{
	Use:   "deel-invoices",
	Short: "List Deel-issued invoices",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListDeelInvoices(cmd.Context(), api.DeelInvoicesListParams{
			Limit:  deelInvoicesLimitFlag,
			Cursor: deelInvoicesCursorFlag,
			Status: deelInvoicesStatusFlag,
		})
		if err != nil {
			f.PrintError("Failed to list Deel invoices: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No Deel invoices found.")
				return
			}
			table := f.NewTable("ID", "NUMBER", "AMOUNT", "STATUS", "ISSUE DATE", "DUE DATE")
			for _, inv := range resp.Data {
				amount := fmt.Sprintf("%.2f %s", inv.Amount, inv.Currency)
				table.AddRow(inv.ID, inv.Number, amount, inv.Status, inv.IssueDate, inv.DueDate)
			}
			table.Render()
		}, resp)
	},
}

func init() {
	invoicesListCmd.Flags().IntVar(&invoicesLimitFlag, "limit", 50, "Maximum results")
	invoicesListCmd.Flags().StringVar(&invoicesCursorFlag, "cursor", "", "Pagination cursor")
	invoicesListCmd.Flags().StringVar(&invoicesStatusFlag, "status", "", "Filter by status")
	invoicesListCmd.Flags().StringVar(&invoicesContractFlag, "contract", "", "Filter by contract ID")

	deelInvoicesCmd.Flags().IntVar(&deelInvoicesLimitFlag, "limit", 50, "Maximum results")
	deelInvoicesCmd.Flags().StringVar(&deelInvoicesCursorFlag, "cursor", "", "Pagination cursor")
	deelInvoicesCmd.Flags().StringVar(&deelInvoicesStatusFlag, "status", "", "Filter by status")

	invoicesCmd.AddCommand(invoicesListCmd)
	invoicesCmd.AddCommand(invoicesGetCmd)
	invoicesCmd.AddCommand(invoicesAdjustmentsCmd)
	invoicesCmd.AddCommand(invoicesPDFCmd)
	invoicesCmd.AddCommand(deelInvoicesCmd)
}
