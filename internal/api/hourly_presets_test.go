package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListHourlyPresets(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":             "hp1",
				"name":           "Standard 40h",
				"description":    "Standard full-time hours",
				"hours_per_day":  8.0,
				"hours_per_week": 40.0,
				"rate":           50.0,
				"currency":       "USD",
				"created_at":     "2024-01-01T00:00:00Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/hourly-presets", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListHourlyPresets(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "hp1", result[0].ID)
	assert.Equal(t, "Standard 40h", result[0].Name)
	assert.Equal(t, 40.0, result[0].HoursPerWeek)
}

func TestListHourlyPresets_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/hourly-presets", http.StatusNotFound, map[string]string{"error": "not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListHourlyPresets(context.Background())

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestCreateHourlyPreset(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/hourly-presets", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Part-time 20h", body["name"])
		assert.Equal(t, "Part-time preset", body["description"])
		assert.Equal(t, 4.0, body["hours_per_day"])
		assert.Equal(t, 20.0, body["hours_per_week"])
		assert.Equal(t, 45.0, body["rate"])
		assert.Equal(t, "EUR", body["currency"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":             "hp-new",
			"name":           "Part-time 20h",
			"description":    "Part-time preset",
			"hours_per_day":  4.0,
			"hours_per_week": 20.0,
			"rate":           45.0,
			"currency":       "EUR",
			"created_at":     "2024-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateHourlyPreset(context.Background(), CreateHourlyPresetParams{
		Name:         "Part-time 20h",
		Description:  "Part-time preset",
		HoursPerDay:  4.0,
		HoursPerWeek: 20.0,
		Rate:         45.0,
		Currency:     "EUR",
	})

	require.NoError(t, err)
	assert.Equal(t, "hp-new", result.ID)
	assert.Equal(t, "Part-time 20h", result.Name)
	assert.Equal(t, 20.0, result.HoursPerWeek)
}

func TestCreateHourlyPreset_Error(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/hourly-presets", http.StatusBadRequest, map[string]string{"error": "invalid parameters"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateHourlyPreset(context.Background(), CreateHourlyPresetParams{
		Name:         "",
		HoursPerDay:  8.0,
		HoursPerWeek: 40.0,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestUpdateHourlyPreset(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/hourly-presets/hp1", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Updated name", body["name"])
		assert.Equal(t, 50.0, body["rate"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":             "hp1",
			"name":           "Updated name",
			"hours_per_day":  8.0,
			"hours_per_week": 40.0,
			"rate":           50.0,
			"currency":       "USD",
			"created_at":     "2024-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateHourlyPreset(context.Background(), "hp1", UpdateHourlyPresetParams{
		Name: "Updated name",
		Rate: 50.0,
	})

	require.NoError(t, err)
	assert.Equal(t, "hp1", result.ID)
	assert.Equal(t, "Updated name", result.Name)
	assert.Equal(t, 50.0, result.Rate)
}

func TestUpdateHourlyPreset_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/hourly-presets/nonexistent", http.StatusNotFound, map[string]string{"error": "not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateHourlyPreset(context.Background(), "nonexistent", UpdateHourlyPresetParams{
		Name: "New name",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestDeleteHourlyPreset(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/hourly-presets/hp1", http.StatusNoContent, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteHourlyPreset(context.Background(), "hp1")

	require.NoError(t, err)
}

func TestDeleteHourlyPreset_NotFound(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/hourly-presets/nonexistent", http.StatusNotFound, map[string]string{"error": "not found"})
	defer server.Close()

	client := testClient(server)
	err := client.DeleteHourlyPreset(context.Background(), "nonexistent")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}
