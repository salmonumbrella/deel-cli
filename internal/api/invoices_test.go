package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInvoicePDF(t *testing.T) {
	// Create a test server that returns PDF bytes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/rest/v2/invoices/inv123/pdf", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/pdf", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		// Write mock PDF data
		_, err := w.Write([]byte("%PDF-1.4 mock pdf content"))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := testClient(server)
	pdfBytes, err := client.GetInvoicePDF(context.Background(), "inv123")

	require.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.Contains(t, string(pdfBytes), "%PDF")
}

func TestGetInvoicePDF_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(`{"error": "invoice not found"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := testClient(server)
	_, err := client.GetInvoicePDF(context.Background(), "inv999")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestListDeelInvoices(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "dinv1",
				"number":      "DEEL-001",
				"status":      "paid",
				"amount":      150.00,
				"currency":    "USD",
				"issue_date":  "2024-01-01",
				"due_date":    "2024-01-31",
				"paid_date":   "2024-01-15",
				"description": "Platform fees",
			},
			{
				"id":          "dinv2",
				"number":      "DEEL-002",
				"status":      "pending",
				"amount":      200.00,
				"currency":    "USD",
				"issue_date":  "2024-02-01",
				"due_date":    "2024-02-28",
				"description": "Service fees",
			},
		},
		"page": map[string]any{
			"next":  "cursor-123",
			"total": 2,
		},
	}
	server := mockServer(t, "GET", "/rest/v2/deel-invoices", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListDeelInvoices(context.Background(), DeelInvoicesListParams{})

	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "dinv1", result.Data[0].ID)
	assert.Equal(t, "DEEL-001", result.Data[0].Number)
	assert.Equal(t, "paid", result.Data[0].Status)
	assert.Equal(t, 150.00, result.Data[0].Amount)
	assert.Equal(t, "cursor-123", result.Page.Next)
	assert.Equal(t, 2, result.Page.Total)
}

func TestListDeelInvoices_WithParams(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":     "dinv1",
				"status": "paid",
			},
		},
		"page": map[string]any{},
	}
	server := mockServerWithQuery(t, "/rest/v2/deel-invoices", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "10", query["limit"])
		assert.Equal(t, "paid", query["status"])
		assert.Equal(t, "cursor-abc", query["cursor"])
	}, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListDeelInvoices(context.Background(), DeelInvoicesListParams{
		Limit:  10,
		Status: "paid",
		Cursor: "cursor-abc",
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
}

func TestListInvoiceAdjustments(t *testing.T) {
	// Test with amount as string (actual API behavior)
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "adj-1",
				"type":        "expense",
				"amount":      "75.50", // API returns string
				"currency":    "USD",
				"description": "Office supplies",
				"status":      "pending",
				"created_at":  "2025-12-26T10:00:00.000Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/invoices/inv123/adjustments", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListInvoiceAdjustments(context.Background(), "inv123")

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "adj-1", result[0].ID)
	assert.Equal(t, 75.50, float64(result[0].Amount))
	assert.Equal(t, "USD", result[0].Currency)
}

func TestListAllInvoiceAdjustments(t *testing.T) {
	// Test with amount as string (actual API behavior)
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":                      "db70024b-f535-481a-9bd9-a7ed8089f746",
				"type":                    "expense",
				"amount":                  "80.00", // API returns string
				"currency":                "USD",
				"description":             "I bought a keyboard",
				"status":                  "processing",
				"date_submitted":          "2025-12-26T16:50:04.578Z",
				"contract_id":             "c123",
				"created_at":              "2025-12-26T16:50:04.578Z",
				"title":                   "Keyboard",
				"adjustment_category_id":  "cat123",
				"date_of_adjustment":      "2024-01-24T00:00:00.000Z",
				"file":                    nil,
				"actual_start_cycle_date": "2025-08-01T00:00:00.000Z",
				"actual_end_cycle_date":   "2025-08-31T00:00:00.000Z",
			},
			{
				"id":             "adj-2",
				"type":           "bonus",
				"amount":         "150.50", // API returns string
				"currency":       "EUR",
				"description":    "Performance bonus",
				"status":         "approved",
				"date_submitted": "2025-12-20T10:00:00.000Z",
				"contract_id":    "c456",
				"created_at":     "2025-12-20T10:00:00.000Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/invoice-adjustments", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListAllInvoiceAdjustments(context.Background(), ListAllInvoiceAdjustmentsParams{})

	require.NoError(t, err)
	assert.Len(t, result, 2)

	// First adjustment
	assert.Equal(t, "db70024b-f535-481a-9bd9-a7ed8089f746", result[0].ID)
	assert.Equal(t, "expense", result[0].Type)
	assert.Equal(t, 80.00, float64(result[0].Amount))
	assert.Equal(t, "USD", result[0].Currency)
	assert.Equal(t, "I bought a keyboard", result[0].Description)
	assert.Equal(t, "processing", result[0].Status)
	assert.Equal(t, "2025-12-26T16:50:04.578Z", result[0].DateSubmitted)
	assert.Equal(t, "c123", result[0].ContractID)
	assert.Equal(t, "Keyboard", result[0].Title)

	// Second adjustment
	assert.Equal(t, "adj-2", result[1].ID)
	assert.Equal(t, "bonus", result[1].Type)
	assert.Equal(t, 150.50, float64(result[1].Amount))
	assert.Equal(t, "EUR", result[1].Currency)
	assert.Equal(t, "approved", result[1].Status)
}

func TestListAllInvoiceAdjustments_WithNumericAmount(t *testing.T) {
	// Test with amount as number (backward compatibility)
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":       "adj-1",
				"type":     "expense",
				"amount":   100.25, // Some APIs might return as number
				"currency": "USD",
				"status":   "pending",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/invoice-adjustments", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListAllInvoiceAdjustments(context.Background(), ListAllInvoiceAdjustmentsParams{})

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 100.25, float64(result[0].Amount))
}

func TestListAllInvoiceAdjustments_WithFilters(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "adj-1",
				"type":        "expense",
				"amount":      "50.00",
				"status":      "pending",
				"contract_id": "c123",
			},
		},
	}
	server := mockServerWithQuery(t, "/rest/v2/invoice-adjustments", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "expense", query["types[]"])
		assert.Equal(t, "c123", query["contract_id"])
		assert.Equal(t, "pending", query["status"])
	}, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListAllInvoiceAdjustments(context.Background(), ListAllInvoiceAdjustmentsParams{
		Types:      []string{"expense"},
		ContractID: "c123",
		Status:     "pending",
	})

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "adj-1", result[0].ID)
}

func TestGetInvoiceAdjustment(t *testing.T) {
	// Single-item endpoint often returns more detailed data
	response := map[string]any{
		"data": map[string]any{
			"id":                      "db70024b-f535-481a-9bd9-a7ed8089f746",
			"type":                    "expense",
			"amount":                  "80.00",
			"currency":                "USD",
			"description":             "I bought a keyboard",
			"status":                  "processing",
			"date_submitted":          "2025-12-26T16:50:04.578Z",
			"contract_id":             "c123",
			"created_at":              "2025-12-26T16:50:04.578Z",
			"title":                   "Keyboard",
			"adjustment_category_id":  "cat123",
			"date_of_adjustment":      "2024-01-24T00:00:00.000Z",
			"file":                    "https://deel.com/files/receipt.pdf",
			"actual_start_cycle_date": "2025-08-01T00:00:00.000Z",
			"actual_end_cycle_date":   "2025-08-31T00:00:00.000Z",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/invoice-adjustments/db70024b-f535-481a-9bd9-a7ed8089f746", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetInvoiceAdjustment(context.Background(), "db70024b-f535-481a-9bd9-a7ed8089f746")

	require.NoError(t, err)
	assert.Equal(t, "db70024b-f535-481a-9bd9-a7ed8089f746", result.ID)
	assert.Equal(t, "expense", result.Type)
	assert.Equal(t, 80.00, float64(result.Amount))
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "I bought a keyboard", result.Description)
	assert.Equal(t, "processing", result.Status)
	assert.Equal(t, "2025-12-26T16:50:04.578Z", result.DateSubmitted)
	assert.Equal(t, "c123", result.ContractID)
	assert.Equal(t, "Keyboard", result.Title)
	assert.Equal(t, "cat123", result.AdjustmentCategoryID)
	assert.Equal(t, "2024-01-24T00:00:00.000Z", result.DateOfAdjustment)
	assert.NotNil(t, result.File)
	assert.Equal(t, "https://deel.com/files/receipt.pdf", *result.File)
	assert.Equal(t, "2025-08-01T00:00:00.000Z", result.ActualStartCycleDate)
	assert.Equal(t, "2025-08-31T00:00:00.000Z", result.ActualEndCycleDate)
}

func TestGetInvoiceAdjustment_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/invoice-adjustments/invalid-id", http.StatusNotFound, map[string]string{"error": "not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetInvoiceAdjustment(context.Background(), "invalid-id")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestGetInvoiceAdjustment_WithNumericAmount(t *testing.T) {
	// Test backward compatibility with numeric amount
	response := map[string]any{
		"data": map[string]any{
			"id":       "adj-1",
			"type":     "bonus",
			"amount":   150.75, // API might return as number
			"currency": "EUR",
			"status":   "approved",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/invoice-adjustments/adj-1", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetInvoiceAdjustment(context.Background(), "adj-1")

	require.NoError(t, err)
	assert.Equal(t, 150.75, float64(result.Amount))
}

func TestGetInvoiceAdjustment_NullFile(t *testing.T) {
	// Test when file is null
	response := map[string]any{
		"data": map[string]any{
			"id":     "adj-1",
			"type":   "expense",
			"amount": "100.00",
			"status": "pending",
			"file":   nil,
		},
	}
	server := mockServer(t, "GET", "/rest/v2/invoice-adjustments/adj-1", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetInvoiceAdjustment(context.Background(), "adj-1")

	require.NoError(t, err)
	assert.Nil(t, result.File)
}
