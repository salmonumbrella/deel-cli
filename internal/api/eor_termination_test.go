package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestEORResignation(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/contracts/eor-123/resignation-request", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Personal reasons", body["reason"])
		assert.Equal(t, "2024-04-01", body["effective_date"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":                 "term-456",
			"contract_id":        "eor-123",
			"type":               "resignation",
			"status":             "pending",
			"reason":             "Personal reasons",
			"effective_date":     "2024-04-01",
			"last_working_day":   "2024-03-31",
			"notice_period_days": 30,
			"created_at":         "2024-03-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.RequestEORResignation(context.Background(), "eor-123", RequestResignationParams{
		Reason:        "Personal reasons",
		EffectiveDate: "2024-04-01",
	})

	require.NoError(t, err)
	assert.Equal(t, "term-456", result.ID)
	assert.Equal(t, "eor-123", result.ContractID)
	assert.Equal(t, "resignation", result.Type)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "Personal reasons", result.Reason)
	assert.Equal(t, "2024-04-01", result.EffectiveDate)
	assert.Equal(t, "2024-03-31", result.LastWorkingDay)
	assert.Equal(t, 30, result.NoticePeriodDays)
}

func TestRequestEORResignation_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/resignation-request", 400, map[string]string{"error": "invalid effective date"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORResignation(context.Background(), "eor-123", RequestResignationParams{
		Reason:        "Personal reasons",
		EffectiveDate: "invalid-date",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestRequestEORResignation_ContractNotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/invalid/resignation-request", 404, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORResignation(context.Background(), "invalid", RequestResignationParams{
		Reason:        "Personal reasons",
		EffectiveDate: "2024-04-01",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestRequestEORResignation_MissingReason(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/resignation-request", 400, map[string]string{"error": "reason is required"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORResignation(context.Background(), "eor-123", RequestResignationParams{
		EffectiveDate: "2024-04-01",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestRequestEORTermination(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/contracts/eor-123/termination-request", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Performance issues", body["reason"])
		assert.Equal(t, "2024-04-01", body["effective_date"])
		assert.Equal(t, true, body["with_cause"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":                 "term-789",
			"contract_id":        "eor-123",
			"type":               "termination",
			"status":             "pending",
			"reason":             "Performance issues",
			"effective_date":     "2024-04-01",
			"last_working_day":   "2024-03-31",
			"notice_period_days": 30,
			"created_at":         "2024-03-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.RequestEORTermination(context.Background(), "eor-123", RequestTerminationParams{
		Reason:        "Performance issues",
		EffectiveDate: "2024-04-01",
		WithCause:     true,
	})

	require.NoError(t, err)
	assert.Equal(t, "term-789", result.ID)
	assert.Equal(t, "eor-123", result.ContractID)
	assert.Equal(t, "termination", result.Type)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "Performance issues", result.Reason)
	assert.Equal(t, "2024-04-01", result.EffectiveDate)
	assert.Equal(t, "2024-03-31", result.LastWorkingDay)
	assert.Equal(t, 30, result.NoticePeriodDays)
}

func TestRequestEORTermination_WithSeverance(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/contracts/eor-123/termination-request", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Restructuring", body["reason"])
		assert.Equal(t, "2024-04-01", body["effective_date"])
		assert.Equal(t, false, body["with_cause"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":                 "term-890",
			"contract_id":        "eor-123",
			"type":               "termination",
			"status":             "pending",
			"reason":             "Restructuring",
			"effective_date":     "2024-04-01",
			"last_working_day":   "2024-03-15",
			"notice_period_days": 15,
			"severance_amount":   25000.0,
			"currency":           "USD",
			"created_at":         "2024-03-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.RequestEORTermination(context.Background(), "eor-123", RequestTerminationParams{
		Reason:        "Restructuring",
		EffectiveDate: "2024-04-01",
		WithCause:     false,
	})

	require.NoError(t, err)
	assert.Equal(t, "term-890", result.ID)
	assert.Equal(t, "termination", result.Type)
	assert.Equal(t, "Restructuring", result.Reason)
	assert.Equal(t, 25000.0, result.SeveranceAmount)
	assert.Equal(t, "USD", result.Currency)
}

func TestRequestEORTermination_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/termination-request", 400, map[string]string{"error": "invalid effective date"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORTermination(context.Background(), "eor-123", RequestTerminationParams{
		Reason:        "Restructuring",
		EffectiveDate: "invalid-date",
		WithCause:     false,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestRequestEORTermination_ContractNotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/invalid/termination-request", 404, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORTermination(context.Background(), "invalid", RequestTerminationParams{
		Reason:        "Restructuring",
		EffectiveDate: "2024-04-01",
		WithCause:     false,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestRequestEORTermination_Unauthorized(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/termination-request", 403, map[string]string{"error": "insufficient permissions"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORTermination(context.Background(), "eor-123", RequestTerminationParams{
		Reason:        "Performance issues",
		EffectiveDate: "2024-04-01",
		WithCause:     true,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 403, apiErr.StatusCode)
}

func TestGetEORTermination(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/eor-123/termination", 200, map[string]any{
		"data": map[string]any{
			"id":                 "term-456",
			"contract_id":        "eor-123",
			"type":               "resignation",
			"status":             "approved",
			"reason":             "Personal reasons",
			"effective_date":     "2024-04-01",
			"last_working_day":   "2024-03-31",
			"notice_period_days": 30,
			"created_at":         "2024-03-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.GetEORTermination(context.Background(), "eor-123")

	require.NoError(t, err)
	assert.Equal(t, "term-456", result.ID)
	assert.Equal(t, "eor-123", result.ContractID)
	assert.Equal(t, "resignation", result.Type)
	assert.Equal(t, "approved", result.Status)
	assert.Equal(t, "Personal reasons", result.Reason)
	assert.Equal(t, "2024-04-01", result.EffectiveDate)
	assert.Equal(t, "2024-03-31", result.LastWorkingDay)
	assert.Equal(t, 30, result.NoticePeriodDays)
}

func TestGetEORTermination_WithSeverance(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/eor-123/termination", 200, map[string]any{
		"data": map[string]any{
			"id":                 "term-890",
			"contract_id":        "eor-123",
			"type":               "termination",
			"status":             "processing",
			"reason":             "Restructuring",
			"effective_date":     "2024-04-01",
			"last_working_day":   "2024-03-15",
			"notice_period_days": 15,
			"severance_amount":   25000.0,
			"currency":           "USD",
			"created_at":         "2024-03-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.GetEORTermination(context.Background(), "eor-123")

	require.NoError(t, err)
	assert.Equal(t, "term-890", result.ID)
	assert.Equal(t, "termination", result.Type)
	assert.Equal(t, "processing", result.Status)
	assert.Equal(t, 25000.0, result.SeveranceAmount)
	assert.Equal(t, "USD", result.Currency)
}

func TestGetEORTermination_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/invalid/termination", 404, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetEORTermination(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestGetEORTermination_NoTermination(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/eor-123/termination", 404, map[string]string{"error": "no termination found for this contract"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetEORTermination(context.Background(), "eor-123")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}
