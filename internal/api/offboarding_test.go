package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOffboardingTracker(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":             "off1",
			"contract_id":    "ct123",
			"worker_name":    "John Doe",
			"status":         "pending",
			"type":           "voluntary",
			"effective_date": "2024-12-31",
			"created_at":     "2024-12-01T10:00:00Z",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/offboarding/tracker/off1", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetOffboardingTracker(context.Background(), "off1")

	require.NoError(t, err)
	assert.Equal(t, "off1", result.ID)
	assert.Equal(t, "ct123", result.ContractID)
	assert.Equal(t, "John Doe", result.WorkerName)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "voluntary", result.Type)
}

func TestGetTerminationDetails(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":             "term1",
			"contract_id":    "ct123",
			"reason":         "Performance issues",
			"status":         "completed",
			"notice_date":    "2024-11-01",
			"effective_date": "2024-12-01",
			"final_pay_date": "2024-12-15",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/terminations/term1", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetTerminationDetails(context.Background(), "term1")

	require.NoError(t, err)
	assert.Equal(t, "term1", result.ID)
	assert.Equal(t, "ct123", result.ContractID)
	assert.Equal(t, "Performance issues", result.Reason)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, "2024-11-01", result.NoticeDate)
	assert.Equal(t, "2024-12-01", result.EffectiveDate)
	assert.Equal(t, "2024-12-15", result.FinalPayDate)
}

func TestGetTerminationDetails_WithoutFinalPayDate(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":             "term2",
			"contract_id":    "ct456",
			"reason":         "Contract ended",
			"status":         "pending",
			"notice_date":    "2024-12-01",
			"effective_date": "2025-01-01",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/terminations/term2", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetTerminationDetails(context.Background(), "term2")

	require.NoError(t, err)
	assert.Equal(t, "term2", result.ID)
	assert.Equal(t, "pending", result.Status)
	assert.Empty(t, result.FinalPayDate)
}
