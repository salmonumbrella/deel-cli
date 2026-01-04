package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestEORResignation(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/eor-123/terminations/", func(t *testing.T, body map[string]any) {
		// Verify data wrapper
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")
		assert.Equal(t, "EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY", data["reason"])
		assert.Equal(t, true, data["is_employee_staying_with_deel"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":                 "term-456",
			"contract_id":        "eor-123",
			"type":               "resignation",
			"status":             "pending",
			"reason":             "EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY",
			"effective_date":     "2024-04-01",
			"last_working_day":   "2024-03-31",
			"notice_period_days": 30,
			"created_at":         "2024-03-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.RequestEORResignation(context.Background(), "eor-123", EORResignationParams{
		Reason:                "EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY",
		IsEmployeeStayingDeel: true,
	})

	require.NoError(t, err)
	assert.Equal(t, "term-456", result.ID)
	assert.Equal(t, "eor-123", result.ContractID)
	assert.Equal(t, "resignation", result.Type)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY", result.Reason)
	assert.Equal(t, "2024-04-01", result.EffectiveDate)
	assert.Equal(t, "2024-03-31", result.LastWorkingDay)
	assert.Equal(t, 30, result.NoticePeriodDays)
}

func TestRequestEORResignation_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/eor-123/terminations/", http.StatusBadRequest, map[string]string{"error": "invalid reason"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORResignation(context.Background(), "eor-123", EORResignationParams{
		Reason:                "INVALID_REASON",
		IsEmployeeStayingDeel: false,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestRequestEORResignation_ContractNotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/invalid/terminations/", http.StatusNotFound, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORResignation(context.Background(), "invalid", EORResignationParams{
		Reason:                "EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY",
		IsEmployeeStayingDeel: false,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestRequestEORTermination(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/eor-123/terminations/", func(t *testing.T, body map[string]any) {
		// Verify data wrapper
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")
		assert.Equal(t, "PERFORMANCE", data["reason"])
		assert.Equal(t, "Employee has not met performance expectations despite multiple improvement plans.", data["reason_detail"])
		assert.Equal(t, true, data["is_employee_notified"])

		// Check used_time_off
		usedTimeOff, ok := data["used_time_off"].(map[string]any)
		require.True(t, ok, "should have used_time_off")
		assert.Equal(t, float64(5), usedTimeOff["paid_time_off"])
		assert.Equal(t, float64(0), usedTimeOff["unpaid_time_off"])
		assert.Equal(t, float64(2), usedTimeOff["sick_leave"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":                 "term-789",
			"contract_id":        "eor-123",
			"type":               "termination",
			"status":             "pending",
			"reason":             "PERFORMANCE",
			"effective_date":     "2024-04-01",
			"last_working_day":   "2024-03-31",
			"notice_period_days": 30,
			"created_at":         "2024-03-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.RequestEORTermination(context.Background(), "eor-123", EORTerminationParams{
		Reason:             "PERFORMANCE",
		ReasonDetail:       "Employee has not met performance expectations despite multiple improvement plans.",
		IsEmployeeNotified: true,
		UsedTimeOff: EORUsedTimeOff{
			PaidTimeOff:   5,
			UnpaidTimeOff: 0,
			SickLeave:     2,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "term-789", result.ID)
	assert.Equal(t, "eor-123", result.ContractID)
	assert.Equal(t, "termination", result.Type)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "PERFORMANCE", result.Reason)
	assert.Equal(t, "2024-04-01", result.EffectiveDate)
	assert.Equal(t, "2024-03-31", result.LastWorkingDay)
	assert.Equal(t, 30, result.NoticePeriodDays)
}

func TestRequestEORTermination_WithSeverance(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/eor-123/terminations/", func(t *testing.T, body map[string]any) {
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")
		assert.Equal(t, "POSITION_ELIMINATION", data["reason"])
		assert.Equal(t, "WEEKS", data["severance_type"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":                 "term-890",
			"contract_id":        "eor-123",
			"type":               "termination",
			"status":             "pending",
			"reason":             "POSITION_ELIMINATION",
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
	result, err := client.RequestEORTermination(context.Background(), "eor-123", EORTerminationParams{
		Reason:             "POSITION_ELIMINATION",
		ReasonDetail:       "The position has been eliminated due to company restructuring and budget reduction initiatives.",
		IsEmployeeNotified: true,
		SeveranceType:      "WEEKS",
		UsedTimeOff: EORUsedTimeOff{
			PaidTimeOff:   0,
			UnpaidTimeOff: 0,
			SickLeave:     0,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "term-890", result.ID)
	assert.Equal(t, "termination", result.Type)
	assert.Equal(t, "POSITION_ELIMINATION", result.Reason)
	assert.Equal(t, 25000.0, result.SeveranceAmount)
	assert.Equal(t, "USD", result.Currency)
}

func TestRequestEORTermination_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/eor-123/terminations/", http.StatusBadRequest, map[string]string{"error": "reason_detail must be at least 100 characters"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORTermination(context.Background(), "eor-123", EORTerminationParams{
		Reason:             "TERMINATION",
		ReasonDetail:       "Too short",
		IsEmployeeNotified: false,
		UsedTimeOff:        EORUsedTimeOff{},
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestRequestEORTermination_ContractNotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/invalid/terminations/", http.StatusNotFound, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORTermination(context.Background(), "invalid", EORTerminationParams{
		Reason:             "TERMINATION",
		ReasonDetail:       "This is a detailed reason that meets the minimum character requirement of 100 characters for termination reasons.",
		IsEmployeeNotified: false,
		UsedTimeOff:        EORUsedTimeOff{},
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestRequestEORTermination_Unauthorized(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/eor-123/terminations/", http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
	defer server.Close()

	client := testClient(server)
	_, err := client.RequestEORTermination(context.Background(), "eor-123", EORTerminationParams{
		Reason:             "FOR_CAUSE",
		ReasonDetail:       "This is a detailed reason that meets the minimum character requirement of 100 characters for termination reasons.",
		IsEmployeeNotified: true,
		UsedTimeOff:        EORUsedTimeOff{},
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}

func TestGetEORTermination(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/eor-123/terminations/", http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":                 "term-456",
			"contract_id":        "eor-123",
			"type":               "resignation",
			"status":             "approved",
			"reason":             "EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY",
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
	assert.Equal(t, "EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY", result.Reason)
	assert.Equal(t, "2024-04-01", result.EffectiveDate)
	assert.Equal(t, "2024-03-31", result.LastWorkingDay)
	assert.Equal(t, 30, result.NoticePeriodDays)
}

func TestGetEORTermination_WithSeverance(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/eor-123/terminations/", http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":                 "term-890",
			"contract_id":        "eor-123",
			"type":               "termination",
			"status":             "processing",
			"reason":             "POSITION_ELIMINATION",
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
	server := mockServer(t, "GET", "/rest/v2/eor/invalid/terminations/", http.StatusNotFound, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetEORTermination(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestGetEORTermination_NoTermination(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/eor-123/terminations/", http.StatusNotFound, map[string]string{"error": "no termination found for this contract"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetEORTermination(context.Background(), "eor-123")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
