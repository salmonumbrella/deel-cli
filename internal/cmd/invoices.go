package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
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
	invoicesAllFlag      bool
)

var invoicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all invoices",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "listing invoices")
		}

		cursor := invoicesCursorFlag
		var allInvoices []api.Invoice
		var next string

		for {
			resp, err := client.ListInvoices(cmd.Context(), api.InvoicesListParams{
				Limit:      invoicesLimitFlag,
				Cursor:     cursor,
				Status:     invoicesStatusFlag,
				ContractID: invoicesContractFlag,
			})
			if err != nil {
				return HandleError(f, err, "listing invoices")
			}
			allInvoices = append(allInvoices, resp.Data...)
			next = resp.Page.Next
			if !invoicesAllFlag || next == "" {
				if !invoicesAllFlag {
					allInvoices = resp.Data
				}
				break
			}
			cursor = next
		}

		response := api.InvoicesListResponse{
			Data: allInvoices,
		}
		response.Page.Next = ""

		return f.Output(func() {
			if len(allInvoices) == 0 {
				f.PrintText("No invoices found.")
				return
			}
			table := f.NewTable("ID", "NUMBER", "WORKER", "AMOUNT", "STATUS", "DUE DATE")
			for _, inv := range allInvoices {
				amount := fmt.Sprintf("%.2f %s", float64(inv.Amount), inv.Currency)
				table.AddRow(inv.ID, inv.Number, inv.WorkerName, amount, inv.Status, inv.DueDate)
			}
			table.Render()
			if !invoicesAllFlag && next != "" {
				f.PrintText("")
				f.PrintText("More results available. Use --cursor to paginate or --all to fetch everything.")
			}
		}, response)
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
			return HandleError(f, err, "getting invoice")
		}

		invoice, err := client.GetInvoice(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting invoice")
		}

		return f.Output(func() {
			f.PrintText("ID:          " + invoice.ID)
			f.PrintText("Number:      " + invoice.Number)
			f.PrintText("Status:      " + invoice.Status)
			f.PrintText(fmt.Sprintf("Amount:      %.2f %s", float64(invoice.Amount), invoice.Currency))
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
	Use:   "adjustments [invoice-id]",
	Short: "List invoice adjustments",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if len(args) == 0 {
			f.PrintError("invoice-id is required")
			return fmt.Errorf("invoice-id is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "listing adjustments")
		}

		adjustments, err := client.ListInvoiceAdjustments(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "listing adjustments")
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

// Flags for invoice adjustment create
var (
	invoiceAdjustmentTypeFlag        string
	invoiceAdjustmentAmountFlag      float64
	invoiceAdjustmentDescriptionFlag string
)

var invoicesAdjustmentsCreateCmd = &cobra.Command{
	Use:   "create <invoice-id>",
	Short: "Create invoice adjustment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if invoiceAdjustmentTypeFlag == "" {
			f.PrintError("--type is required")
			return fmt.Errorf("--type is required")
		}
		if invoiceAdjustmentAmountFlag == 0 {
			f.PrintError("--amount is required")
			return fmt.Errorf("--amount is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "InvoiceAdjustment",
			Description: "Create invoice adjustment",
			Details: map[string]string{
				"InvoiceID":   args[0],
				"Type":        invoiceAdjustmentTypeFlag,
				"Amount":      fmt.Sprintf("%.2f", invoiceAdjustmentAmountFlag),
				"Description": invoiceAdjustmentDescriptionFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "creating adjustment")
		}

		params := api.CreateInvoiceAdjustmentParams{
			Type:        invoiceAdjustmentTypeFlag,
			Amount:      invoiceAdjustmentAmountFlag,
			Description: invoiceAdjustmentDescriptionFlag,
		}

		adjustment, err := client.CreateInvoiceAdjustment(cmd.Context(), args[0], params)
		if err != nil {
			return HandleError(f, err, "creating adjustment")
		}

		return f.Output(func() {
			f.PrintSuccess("Adjustment created successfully")
			f.PrintText("ID:     " + adjustment.ID)
			f.PrintText("Type:   " + adjustment.Type)
			f.PrintText(fmt.Sprintf("Amount: %.2f %s", adjustment.Amount, adjustment.Currency))
			if adjustment.Description != "" {
				f.PrintText("Description: " + adjustment.Description)
			}
		}, adjustment)
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
			return HandleError(f, err, "downloading invoice")
		}

		pdfBytes, err := client.GetInvoicePDF(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "downloading invoice")
		}

		outputPath := invoicesPDFOutputFlag
		if outputPath == "" {
			outputPath = fmt.Sprintf("invoice-%s.pdf", args[0])
		}

		if outputPath == "-" {
			if _, err := os.Stdout.Write(pdfBytes); err != nil {
				return HandleError(f, err, "writing PDF to stdout")
			}
			return nil
		}

		if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
			return HandleError(f, err, "saving PDF")
		}

		f.PrintSuccess("Saved invoice to %s", outputPath)
		return nil
	},
}

var (
	deelInvoicesLimitFlag  int
	deelInvoicesCursorFlag string
	deelInvoicesStatusFlag string
	invoicesPDFOutputFlag  string
	deelInvoicesAllFlag    bool
)

var deelInvoicesCmd = &cobra.Command{
	Use:   "deel-invoices",
	Short: "List Deel-issued invoices",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "listing invoices")
		}

		cursor := deelInvoicesCursorFlag
		var allDeelInvoices []api.DeelInvoice
		var next string

		for {
			resp, err := client.ListDeelInvoices(cmd.Context(), api.DeelInvoicesListParams{
				Limit:  deelInvoicesLimitFlag,
				Cursor: cursor,
				Status: deelInvoicesStatusFlag,
			})
			if err != nil {
				return HandleError(f, err, "listing invoices")
			}
			allDeelInvoices = append(allDeelInvoices, resp.Data...)
			next = resp.Page.Next
			if !deelInvoicesAllFlag || next == "" {
				if !deelInvoicesAllFlag {
					allDeelInvoices = resp.Data
				}
				break
			}
			cursor = next
		}

		response := api.DeelInvoicesListResponse{
			Data: allDeelInvoices,
		}
		response.Page.Next = ""

		return f.Output(func() {
			if len(allDeelInvoices) == 0 {
				f.PrintText("No Deel invoices found.")
				return
			}
			table := f.NewTable("ID", "NUMBER", "AMOUNT", "STATUS", "ISSUE DATE", "DUE DATE")
			for _, inv := range allDeelInvoices {
				amount := fmt.Sprintf("%.2f %s", inv.Amount, inv.Currency)
				table.AddRow(inv.ID, inv.Number, amount, inv.Status, inv.IssueDate, inv.DueDate)
			}
			table.Render()
			if !deelInvoicesAllFlag && next != "" {
				f.PrintText("")
				f.PrintText("More results available. Use --cursor to paginate or --all to fetch everything.")
			}
		}, response)
	},
}

func init() {
	invoicesListCmd.Flags().IntVar(&invoicesLimitFlag, "limit", 50, "Maximum results")
	invoicesListCmd.Flags().StringVar(&invoicesCursorFlag, "cursor", "", "Pagination cursor")
	invoicesListCmd.Flags().StringVar(&invoicesStatusFlag, "status", "", "Filter by status")
	invoicesListCmd.Flags().StringVar(&invoicesContractFlag, "contract", "", "Filter by contract ID")
	invoicesListCmd.Flags().BoolVar(&invoicesAllFlag, "all", false, "Fetch all pages")

	deelInvoicesCmd.Flags().IntVar(&deelInvoicesLimitFlag, "limit", 50, "Maximum results")
	deelInvoicesCmd.Flags().StringVar(&deelInvoicesCursorFlag, "cursor", "", "Pagination cursor")
	deelInvoicesCmd.Flags().StringVar(&deelInvoicesStatusFlag, "status", "", "Filter by status")
	deelInvoicesCmd.Flags().BoolVar(&deelInvoicesAllFlag, "all", false, "Fetch all pages")

	invoicesAdjustmentsCreateCmd.Flags().StringVar(&invoiceAdjustmentTypeFlag, "type", "", "Adjustment type (required)")
	invoicesAdjustmentsCreateCmd.Flags().Float64Var(&invoiceAdjustmentAmountFlag, "amount", 0, "Adjustment amount (required)")
	invoicesAdjustmentsCreateCmd.Flags().StringVar(&invoiceAdjustmentDescriptionFlag, "description", "", "Adjustment description")

	invoicesCmd.AddCommand(invoicesListCmd)
	invoicesCmd.AddCommand(invoicesGetCmd)
	invoicesCmd.AddCommand(invoicesAdjustmentsCmd)
	invoicesCmd.AddCommand(invoicesPDFCmd)
	invoicesCmd.AddCommand(deelInvoicesCmd)

	invoicesPDFCmd.Flags().StringVar(&invoicesPDFOutputFlag, "output", "", "Output path for PDF ('-' for stdout)")
	invoicesAdjustmentsCmd.AddCommand(invoicesAdjustmentsCreateCmd)
}
