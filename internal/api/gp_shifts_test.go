package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGPShift(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/shifts", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "ext-shift-123", body["external_id"])
		assert.Equal(t, "w-789", body["worker_id"])
		assert.Equal(t, "2024-03-15", body["date"])
		assert.Equal(t, "09:00", body["start_time"])
		assert.Equal(t, "17:00", body["end_time"])
		assert.Equal(t, 60.0, body["break_minutes"])
		assert.Equal(t, "rate-001", body["rate_id"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":            "shift-456",
			"external_id":   "ext-shift-123",
			"worker_id":     "w-789",
			"date":          "2024-03-15",
			"start_time":    "09:00",
			"end_time":      "17:00",
			"break_minutes": 60,
			"total_hours":   7.0,
			"rate_id":       "rate-001",
			"status":        "pending",
			"created_at":    "2024-03-10T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateGPShift(context.Background(), CreateGPShiftParams{
		ExternalID:   "ext-shift-123",
		WorkerID:     "w-789",
		Date:         "2024-03-15",
		StartTime:    "09:00",
		EndTime:      "17:00",
		BreakMinutes: 60,
		RateID:       "rate-001",
	})

	require.NoError(t, err)
	assert.Equal(t, "shift-456", result.ID)
	assert.Equal(t, "ext-shift-123", result.ExternalID)
	assert.Equal(t, "w-789", result.WorkerID)
	assert.Equal(t, "2024-03-15", result.Date)
	assert.Equal(t, "09:00", result.StartTime)
	assert.Equal(t, "17:00", result.EndTime)
	assert.Equal(t, 60, result.BreakMinutes)
	assert.Equal(t, 7.0, result.TotalHours)
	assert.Equal(t, "rate-001", result.RateID)
	assert.Equal(t, "pending", result.Status)
}

func TestCreateGPShift_WithoutOptionalFields(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/shifts", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "w-789", body["worker_id"])
		assert.Equal(t, "2024-03-15", body["date"])
		assert.Equal(t, "09:00", body["start_time"])
		assert.Equal(t, "17:00", body["end_time"])
		assert.Equal(t, 30.0, body["break_minutes"])
		_, hasExternalID := body["external_id"]
		assert.False(t, hasExternalID)
		_, hasRateID := body["rate_id"]
		assert.False(t, hasRateID)
	}, 201, map[string]any{
		"data": map[string]any{
			"id":            "shift-457",
			"worker_id":     "w-789",
			"date":          "2024-03-15",
			"start_time":    "09:00",
			"end_time":      "17:00",
			"break_minutes": 30,
			"total_hours":   7.5,
			"status":        "pending",
			"created_at":    "2024-03-10T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateGPShift(context.Background(), CreateGPShiftParams{
		WorkerID:     "w-789",
		Date:         "2024-03-15",
		StartTime:    "09:00",
		EndTime:      "17:00",
		BreakMinutes: 30,
	})

	require.NoError(t, err)
	assert.Equal(t, "shift-457", result.ID)
	assert.Equal(t, "", result.ExternalID)
	assert.Equal(t, "", result.RateID)
	assert.Equal(t, 7.5, result.TotalHours)
}

func TestCreateGPShift_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/gp/shifts", 400, map[string]string{"error": "invalid date format"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateGPShift(context.Background(), CreateGPShiftParams{
		WorkerID:     "w-789",
		Date:         "invalid-date",
		StartTime:    "09:00",
		EndTime:      "17:00",
		BreakMinutes: 30,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestUpdateGPShift(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/gp/shifts/shift-456", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "2024-03-16", body["date"])
		assert.Equal(t, "08:00", body["start_time"])
		assert.Equal(t, "16:00", body["end_time"])
		assert.Equal(t, 45.0, body["break_minutes"])
		assert.Equal(t, "rate-002", body["rate_id"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":            "shift-456",
			"external_id":   "ext-shift-123",
			"worker_id":     "w-789",
			"date":          "2024-03-16",
			"start_time":    "08:00",
			"end_time":      "16:00",
			"break_minutes": 45,
			"total_hours":   7.25,
			"rate_id":       "rate-002",
			"status":        "pending",
			"created_at":    "2024-03-10T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGPShift(context.Background(), "shift-456", UpdateGPShiftParams{
		Date:         "2024-03-16",
		StartTime:    "08:00",
		EndTime:      "16:00",
		BreakMinutes: 45,
		RateID:       "rate-002",
	})

	require.NoError(t, err)
	assert.Equal(t, "shift-456", result.ID)
	assert.Equal(t, "2024-03-16", result.Date)
	assert.Equal(t, "08:00", result.StartTime)
	assert.Equal(t, "16:00", result.EndTime)
	assert.Equal(t, 45, result.BreakMinutes)
	assert.Equal(t, 7.25, result.TotalHours)
	assert.Equal(t, "rate-002", result.RateID)
}

func TestUpdateGPShift_PartialUpdate(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/gp/shifts/shift-456", func(t *testing.T, body map[string]any) {
		assert.Equal(t, 30.0, body["break_minutes"])
		_, hasDate := body["date"]
		assert.False(t, hasDate)
		_, hasStartTime := body["start_time"]
		assert.False(t, hasStartTime)
		_, hasEndTime := body["end_time"]
		assert.False(t, hasEndTime)
		_, hasRateID := body["rate_id"]
		assert.False(t, hasRateID)
	}, 200, map[string]any{
		"data": map[string]any{
			"id":            "shift-456",
			"external_id":   "ext-shift-123",
			"worker_id":     "w-789",
			"date":          "2024-03-15",
			"start_time":    "09:00",
			"end_time":      "17:00",
			"break_minutes": 30,
			"total_hours":   7.5,
			"rate_id":       "rate-001",
			"status":        "pending",
			"created_at":    "2024-03-10T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGPShift(context.Background(), "shift-456", UpdateGPShiftParams{
		BreakMinutes: 30,
	})

	require.NoError(t, err)
	assert.Equal(t, "shift-456", result.ID)
	assert.Equal(t, 30, result.BreakMinutes)
	assert.Equal(t, 7.5, result.TotalHours)
}

func TestUpdateGPShift_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/gp/shifts/invalid", 404, map[string]string{"error": "shift not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateGPShift(context.Background(), "invalid", UpdateGPShiftParams{
		BreakMinutes: 30,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestDeleteGPShift(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/gp/shifts/external/ext-shift-123", 204, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteGPShift(context.Background(), "ext-shift-123")

	require.NoError(t, err)
}

func TestDeleteGPShift_NotFound(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/gp/shifts/external/invalid", 404, map[string]string{"error": "shift not found"})
	defer server.Close()

	client := testClient(server)
	err := client.DeleteGPShift(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestDeleteGPShift_Conflict(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/gp/shifts/external/ext-shift-123", 409, map[string]string{"error": "shift already processed"})
	defer server.Close()

	client := testClient(server)
	err := client.DeleteGPShift(context.Background(), "ext-shift-123")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 409, apiErr.StatusCode)
}
