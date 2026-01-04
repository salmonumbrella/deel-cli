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
	}, http.StatusOK, response)
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
