package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListG2NReports(t *testing.T) {
	server := mockServerWithQuery(t, "GET", "/rest/v2/gp/reports/gross-to-net", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "w-123", query["worker_id"])
		assert.Equal(t, "2024-03", query["period"])
	}, http.StatusOK, map[string]any{
		"data": []map[string]any{
			{
				"id":           "g2n-001",
				"worker_id":    "w-123",
				"worker_name":  "John Doe",
				"period":       "2024-03",
				"gross_amount": 8000.0,
				"net_amount":   6500.0,
				"deductions":   500.0,
				"taxes":        1000.0,
				"currency":     "GBP",
				"status":       "completed",
				"created_at":   "2024-03-31T23:59:59Z",
			},
			{
				"id":           "g2n-002",
				"worker_id":    "w-123",
				"worker_name":  "John Doe",
				"period":       "2024-02",
				"gross_amount": 8000.0,
				"net_amount":   6400.0,
				"deductions":   600.0,
				"taxes":        1000.0,
				"currency":     "GBP",
				"status":       "completed",
				"created_at":   "2024-02-29T23:59:59Z",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	results, err := client.ListG2NReports(context.Background(), ListG2NReportsParams{
		WorkerID: "w-123",
		Period:   "2024-03",
	})

	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "g2n-001", results[0].ID)
	assert.Equal(t, "w-123", results[0].WorkerID)
	assert.Equal(t, "John Doe", results[0].WorkerName)
	assert.Equal(t, "2024-03", results[0].Period)
	assert.Equal(t, 8000.0, results[0].GrossAmount)
	assert.Equal(t, 6500.0, results[0].NetAmount)
	assert.Equal(t, 500.0, results[0].Deductions)
	assert.Equal(t, 1000.0, results[0].Taxes)
	assert.Equal(t, "GBP", results[0].Currency)
	assert.Equal(t, "completed", results[0].Status)
}

func TestListG2NReports_EmptyResults(t *testing.T) {
	server := mockServerWithQuery(t, "GET", "/rest/v2/gp/reports/gross-to-net", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "w-999", query["worker_id"])
	}, http.StatusOK, map[string]any{
		"data": []map[string]any{},
	})
	defer server.Close()

	client := testClient(server)
	results, err := client.ListG2NReports(context.Background(), ListG2NReportsParams{
		WorkerID: "w-999",
	})

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestListG2NReports_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/gp/reports/gross-to-net", http.StatusNotFound, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListG2NReports(context.Background(), ListG2NReportsParams{
		WorkerID: "invalid",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestDownloadG2NReport(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/gp/reports/gross-to-net/download", http.StatusOK, map[string]any{
		"data": map[string]any{
			"report_id":    "g2n-001",
			"download_url": "https://example.com/reports/g2n-001.pdf",
			"expires_at":   "2024-04-01T12:00:00Z",
			"format":       "pdf",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.DownloadG2NReport(context.Background(), "g2n-001")

	require.NoError(t, err)
	assert.Equal(t, "g2n-001", result.ReportID)
	assert.Equal(t, "https://example.com/reports/g2n-001.pdf", result.DownloadURL)
	assert.Equal(t, "2024-04-01T12:00:00Z", result.ExpiresAt)
	assert.Equal(t, "pdf", result.Format)
}

func TestDownloadG2NReport_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/gp/reports/gross-to-net/download", http.StatusNotFound, map[string]string{"error": "report not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.DownloadG2NReport(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestRequestGPTermination(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/termination-requests", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "w-123", body["worker_id"])
		assert.Equal(t, "Resignation", body["reason"])
		assert.Equal(t, "2024-05-31", body["effective_date"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":             "term-001",
			"worker_id":      "w-123",
			"reason":         "Resignation",
			"effective_date": "2024-05-31",
			"status":         "pending",
			"created_at":     "2024-04-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.RequestGPTermination(context.Background(), RequestGPTerminationParams{
		WorkerID:      "w-123",
		Reason:        "Resignation",
		EffectiveDate: "2024-05-31",
	})

	require.NoError(t, err)
	assert.Equal(t, "term-001", result.ID)
	assert.Equal(t, "w-123", result.WorkerID)
	assert.Equal(t, "Resignation", result.Reason)
	assert.Equal(t, "2024-05-31", result.EffectiveDate)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "2024-04-01T10:00:00Z", result.CreatedAt)
}

func TestRequestGPTermination_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/gp/termination-requests", http.StatusBadRequest, map[string]string{"error": "effective_date is required"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestGPTermination(context.Background(), RequestGPTerminationParams{
		WorkerID: "w-123",
		Reason:   "Resignation",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestRequestGPTermination_WorkerNotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/gp/termination-requests", http.StatusNotFound, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestGPTermination(context.Background(), RequestGPTerminationParams{
		WorkerID:      "invalid",
		Reason:        "Resignation",
		EffectiveDate: "2024-05-31",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
