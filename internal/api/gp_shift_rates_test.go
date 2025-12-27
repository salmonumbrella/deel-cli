package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGPShiftRate(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/shift-rates", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "ext-sr-123", body["external_id"])
		assert.Equal(t, "Night Shift Premium", body["name"])
		assert.Equal(t, "Additional pay for night shifts", body["description"])
		assert.Equal(t, 25.50, body["rate"])
		assert.Equal(t, "USD", body["currency"])
		assert.Equal(t, "hourly", body["type"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":          "sr-456",
			"external_id": "ext-sr-123",
			"name":        "Night Shift Premium",
			"description": "Additional pay for night shifts",
			"rate":        25.50,
			"currency":    "USD",
			"type":        "hourly",
			"status":      "active",
			"created_at":  "2024-02-20T09:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateGPShiftRate(context.Background(), GPCreateShiftRateParams{
		ExternalID:  "ext-sr-123",
		Name:        "Night Shift Premium",
		Description: "Additional pay for night shifts",
		Rate:        25.50,
		Currency:    "USD",
		Type:        "hourly",
	})

	require.NoError(t, err)
	assert.Equal(t, "sr-456", result.ID)
	assert.Equal(t, "ext-sr-123", result.ExternalID)
	assert.Equal(t, "Night Shift Premium", result.Name)
	assert.Equal(t, "Additional pay for night shifts", result.Description)
	assert.Equal(t, 25.50, result.Rate)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "hourly", result.Type)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "2024-02-20T09:00:00Z", result.CreatedAt)
}

func TestCreateGPShiftRate_DailyRate(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/shift-rates", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Weekend Shift", body["name"])
		assert.Equal(t, 200.00, body["rate"])
		assert.Equal(t, "EUR", body["currency"])
		assert.Equal(t, "daily", body["type"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":         "sr-789",
			"name":       "Weekend Shift",
			"rate":       200.00,
			"currency":   "EUR",
			"type":       "daily",
			"status":     "active",
			"created_at": "2024-02-21T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateGPShiftRate(context.Background(), GPCreateShiftRateParams{
		Name:     "Weekend Shift",
		Rate:     200.00,
		Currency: "EUR",
		Type:     "daily",
	})

	require.NoError(t, err)
	assert.Equal(t, "sr-789", result.ID)
	assert.Equal(t, "Weekend Shift", result.Name)
	assert.Equal(t, 200.00, result.Rate)
	assert.Equal(t, "EUR", result.Currency)
	assert.Equal(t, "daily", result.Type)
}

func TestCreateGPShiftRate_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/gp/shift-rates", 400, map[string]string{"error": "invalid rate type"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateGPShiftRate(context.Background(), GPCreateShiftRateParams{
		Name:     "Invalid Shift",
		Rate:     50.00,
		Currency: "USD",
		Type:     "invalid",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestListGPShiftRates(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/gp/shift-rates", 200, map[string]any{
		"data": []map[string]any{
			{
				"id":          "sr-456",
				"external_id": "ext-sr-123",
				"name":        "Night Shift Premium",
				"description": "Additional pay for night shifts",
				"rate":        25.50,
				"currency":    "USD",
				"type":        "hourly",
				"status":      "active",
				"created_at":  "2024-02-20T09:00:00Z",
			},
			{
				"id":         "sr-789",
				"name":       "Weekend Shift",
				"rate":       200.00,
				"currency":   "EUR",
				"type":       "daily",
				"status":     "active",
				"created_at": "2024-02-21T10:00:00Z",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.ListGPShiftRates(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "sr-456", result[0].ID)
	assert.Equal(t, "ext-sr-123", result[0].ExternalID)
	assert.Equal(t, "Night Shift Premium", result[0].Name)
	assert.Equal(t, 25.50, result[0].Rate)
	assert.Equal(t, "hourly", result[0].Type)
	assert.Equal(t, "sr-789", result[1].ID)
	assert.Equal(t, "Weekend Shift", result[1].Name)
	assert.Equal(t, 200.00, result[1].Rate)
	assert.Equal(t, "daily", result[1].Type)
}

func TestListGPShiftRates_Empty(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/gp/shift-rates", 200, map[string]any{
		"data": []map[string]any{},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.ListGPShiftRates(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestUpdateGPShiftRate(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/gp/shift-rates/sr-456", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Night Shift Premium Updated", body["name"])
		assert.Equal(t, "Updated description", body["description"])
		assert.Equal(t, 30.00, body["rate"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":          "sr-456",
			"external_id": "ext-sr-123",
			"name":        "Night Shift Premium Updated",
			"description": "Updated description",
			"rate":        30.00,
			"currency":    "USD",
			"type":        "hourly",
			"status":      "active",
			"created_at":  "2024-02-20T09:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGPShiftRate(context.Background(), "sr-456", GPUpdateShiftRateParams{
		Name:        "Night Shift Premium Updated",
		Description: "Updated description",
		Rate:        30.00,
	})

	require.NoError(t, err)
	assert.Equal(t, "sr-456", result.ID)
	assert.Equal(t, "Night Shift Premium Updated", result.Name)
	assert.Equal(t, "Updated description", result.Description)
	assert.Equal(t, 30.00, result.Rate)
}

func TestUpdateGPShiftRate_PartialUpdate(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/gp/shift-rates/sr-789", func(t *testing.T, body map[string]any) {
		assert.Equal(t, 250.00, body["rate"])
		// Ensure other fields are not present
		_, hasName := body["name"]
		_, hasDescription := body["description"]
		assert.False(t, hasName)
		assert.False(t, hasDescription)
	}, 200, map[string]any{
		"data": map[string]any{
			"id":         "sr-789",
			"name":       "Weekend Shift",
			"rate":       250.00,
			"currency":   "EUR",
			"type":       "daily",
			"status":     "active",
			"created_at": "2024-02-21T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGPShiftRate(context.Background(), "sr-789", GPUpdateShiftRateParams{
		Rate: 250.00,
	})

	require.NoError(t, err)
	assert.Equal(t, "sr-789", result.ID)
	assert.Equal(t, 250.00, result.Rate)
}

func TestUpdateGPShiftRate_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/gp/shift-rates/invalid", 404, map[string]string{"error": "shift rate not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateGPShiftRate(context.Background(), "invalid", GPUpdateShiftRateParams{
		Rate: 100.00,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestDeleteGPShiftRate(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/gp/shift-rates/external/ext-sr-123", 204, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteGPShiftRate(context.Background(), "ext-sr-123")

	require.NoError(t, err)
}

func TestDeleteGPShiftRate_NotFound(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/gp/shift-rates/external/invalid", 404, map[string]string{"error": "shift rate not found"})
	defer server.Close()

	client := testClient(server)
	err := client.DeleteGPShiftRate(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}
