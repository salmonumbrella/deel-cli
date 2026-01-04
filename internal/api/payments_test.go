package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIndividualPaymentBreakdown(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"payment_id":       "pay123",
			"gross_amount":     1000.00,
			"net_amount":       850.00,
			"currency":         "USD",
			"deel_fee":         50.00,
			"withholding_tax":  100.00,
			"other_deductions": 0.00,
			"reimbursements":   0.00,
			"line_items": []map[string]any{
				{
					"description": "Base salary",
					"amount":      1000.00,
					"type":        "salary",
				},
				{
					"description": "Deel processing fee",
					"amount":      -50.00,
					"type":        "fee",
				},
				{
					"description": "Tax withholding",
					"amount":      -100.00,
					"type":        "tax",
				},
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/payments/pay123/breakdown", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetIndividualPaymentBreakdown(context.Background(), "pay123")

	require.NoError(t, err)
	assert.Equal(t, "pay123", result.PaymentID)
	assert.Equal(t, 1000.00, result.GrossAmount)
	assert.Equal(t, 850.00, result.NetAmount)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, 50.00, result.DeelFee)
	assert.Equal(t, 100.00, result.WithholdingTax)
	assert.Len(t, result.LineItems, 3)
	assert.Equal(t, "Base salary", result.LineItems[0].Description)
}

func TestGetIndividualPaymentBreakdown_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/payments/invalid/breakdown", http.StatusNotFound, map[string]string{
		"error": "payment not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetIndividualPaymentBreakdown(context.Background(), "invalid")

	require.Error(t, err)
}

func TestListDetailedPaymentReceipts(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "rcpt1",
				"payment_id":  "pay123",
				"receipt_url": "https://deel.com/receipts/rcpt1.pdf",
				"amount":      850.00,
				"currency":    "USD",
				"issue_date":  "2024-01-15",
				"contract_id": "c123",
				"worker_name": "John Doe",
				"description": "Payment receipt",
			},
			{
				"id":          "rcpt2",
				"payment_id":  "pay456",
				"receipt_url": "https://deel.com/receipts/rcpt2.pdf",
				"amount":      1200.00,
				"currency":    "EUR",
				"issue_date":  "2024-02-15",
				"contract_id": "c124",
				"worker_name": "Jane Smith",
				"description": "Payment receipt",
			},
		},
		"page": map[string]any{
			"next":  "cursor-xyz",
			"total": 2,
		},
	}
	server := mockServer(t, "GET", "/rest/v2/payment-receipts", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListDetailedPaymentReceipts(context.Background(), DetailedPaymentReceiptsListParams{})

	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "rcpt1", result.Data[0].ID)
	assert.Equal(t, "pay123", result.Data[0].PaymentID)
	assert.Equal(t, "https://deel.com/receipts/rcpt1.pdf", result.Data[0].ReceiptURL)
	assert.Equal(t, 850.00, result.Data[0].Amount)
	assert.Equal(t, "USD", result.Data[0].Currency)
	assert.Equal(t, "cursor-xyz", result.Page.Next)
	assert.Equal(t, 2, result.Page.Total)
}

func TestListDetailedPaymentReceipts_WithParams(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":         "rcpt1",
				"payment_id": "pay123",
			},
		},
		"page": map[string]any{},
	}
	server := mockServerWithQuery(t, "GET", "/rest/v2/payment-receipts", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "20", query["limit"])
		assert.Equal(t, "c123", query["contract_id"])
		assert.Equal(t, "pay123", query["payment_id"])
		assert.Equal(t, "cursor-def", query["cursor"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListDetailedPaymentReceipts(context.Background(), DetailedPaymentReceiptsListParams{
		Limit:      20,
		ContractID: "c123",
		PaymentID:  "pay123",
		Cursor:     "cursor-def",
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
}

func TestGetDetailedPaymentsReport(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"report_id":    "rpt123",
			"generated_at": "2024-03-15T10:00:00Z",
			"start_date":   "2024-01-01",
			"end_date":     "2024-03-31",
			"total_amount": 5000.00,
			"currency":     "USD",
			"payments": []map[string]any{
				{
					"payment_id":   "pay1",
					"contract_id":  "c123",
					"worker_name":  "John Doe",
					"amount":       1000.00,
					"currency":     "USD",
					"payment_date": "2024-01-15",
					"status":       "paid",
					"type":         "salary",
				},
				{
					"payment_id":   "pay2",
					"contract_id":  "c124",
					"worker_name":  "Jane Smith",
					"amount":       1500.00,
					"currency":     "USD",
					"payment_date": "2024-02-15",
					"status":       "paid",
					"type":         "salary",
				},
				{
					"payment_id":   "pay3",
					"contract_id":  "c125",
					"worker_name":  "Bob Johnson",
					"amount":       2500.00,
					"currency":     "USD",
					"payment_date": "2024-03-15",
					"status":       "processing",
					"type":         "bonus",
				},
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/reports/payments/detailed", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetDetailedPaymentsReport(context.Background(), DetailedPaymentsReportParams{})

	require.NoError(t, err)
	assert.Equal(t, "rpt123", result.ReportID)
	assert.Equal(t, "2024-03-15T10:00:00Z", result.GeneratedAt)
	assert.Equal(t, "2024-01-01", result.StartDate)
	assert.Equal(t, "2024-03-31", result.EndDate)
	assert.Equal(t, 5000.00, result.TotalAmount)
	assert.Equal(t, "USD", result.Currency)
	assert.Len(t, result.Payments, 3)
	assert.Equal(t, "pay1", result.Payments[0].PaymentID)
	assert.Equal(t, "John Doe", result.Payments[0].WorkerName)
	assert.Equal(t, "paid", result.Payments[0].Status)
}

func TestGetDetailedPaymentsReport_WithParams(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"report_id": "rpt456",
			"payments":  []map[string]any{},
		},
	}
	server := mockServerWithQuery(t, "GET", "/rest/v2/reports/payments/detailed", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "2024-01-01", query["start_date"])
		assert.Equal(t, "2024-01-31", query["end_date"])
		assert.Equal(t, "c123", query["contract_id"])
		assert.Equal(t, "paid", query["status"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetDetailedPaymentsReport(context.Background(), DetailedPaymentsReportParams{
		StartDate:  "2024-01-01",
		EndDate:    "2024-01-31",
		ContractID: "c123",
		Status:     "paid",
	})

	require.NoError(t, err)
	assert.Equal(t, "rpt456", result.ReportID)
}

func TestGetDetailedPaymentsReport_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/reports/payments/detailed", http.StatusBadRequest, map[string]string{
		"error": "invalid date range",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetDetailedPaymentsReport(context.Background(), DetailedPaymentsReportParams{
		StartDate: "invalid",
		EndDate:   "invalid",
	})

	require.Error(t, err)
}
