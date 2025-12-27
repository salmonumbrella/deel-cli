package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTimesheets(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":           "ts1",
				"contract_id":  "c1",
				"status":       "pending",
				"period_start": "2024-01-01",
				"period_end":   "2024-01-31",
				"total_hours":  160.0,
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/timesheets", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListTimesheets(context.Background(), TimesheetsListParams{})

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "ts1", result[0].ID)
	assert.Equal(t, "c1", result[0].ContractID)
	assert.Equal(t, 160.0, result[0].TotalHours)
}

func TestListTimesheets_WithQueryParams(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":           "ts1",
				"contract_id":  "c1",
				"status":       "approved",
				"period_start": "2024-01-01",
				"period_end":   "2024-01-31",
				"total_hours":  160.0,
			},
		},
	}

	server := mockServerWithQuery(t, "GET", "/rest/v2/timesheets", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "c1", query["contract_id"])
		assert.Equal(t, "approved", query["status"])
		assert.Equal(t, "10", query["limit"])
		assert.Equal(t, "abc123", query["cursor"])
	}, 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListTimesheets(context.Background(), TimesheetsListParams{
		ContractID: "c1",
		Status:     "approved",
		Limit:      10,
		Cursor:     "abc123",
	})

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "ts1", result[0].ID)
	assert.Equal(t, "approved", result[0].Status)
}

func TestListTimesheets_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/timesheets", 400, map[string]string{"error": "bad request"})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListTimesheets(context.Background(), TimesheetsListParams{})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestGetTimesheet(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":           "ts1",
			"contract_id":  "c1",
			"status":       "approved",
			"period_start": "2024-01-01",
			"period_end":   "2024-01-31",
			"total_hours":  160.0,
			"entries": []map[string]any{
				{
					"id":           "e1",
					"timesheet_id": "ts1",
					"date":         "2024-01-02",
					"hours":        8.0,
					"description":  "Development work",
				},
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/timesheets/ts1", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetTimesheet(context.Background(), "ts1")

	require.NoError(t, err)
	assert.Equal(t, "ts1", result.ID)
	assert.Equal(t, "approved", result.Status)
	assert.Len(t, result.Entries, 1)
	assert.Equal(t, "e1", result.Entries[0].ID)
}

func TestGetTimesheet_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/timesheets/invalid", 404, map[string]string{"error": "timesheet not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetTimesheet(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestCreateTimesheetEntry(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/timesheet-entries", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "ts1", body["timesheet_id"])
		assert.Equal(t, "2024-01-02", body["date"])
		assert.Equal(t, 8.0, body["hours"])
		assert.Equal(t, "Development work", body["description"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":           "e-new",
			"timesheet_id": "ts1",
			"date":         "2024-01-02",
			"hours":        8.0,
			"description":  "Development work",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateTimesheetEntry(context.Background(), CreateTimesheetEntryParams{
		TimesheetID: "ts1",
		Date:        "2024-01-02",
		Hours:       8.0,
		Description: "Development work",
	})

	require.NoError(t, err)
	assert.Equal(t, "e-new", result.ID)
	assert.Equal(t, 8.0, result.Hours)
}

func TestCreateTimesheetEntry_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/timesheet-entries", 400, map[string]string{"error": "invalid hours value"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateTimesheetEntry(context.Background(), CreateTimesheetEntryParams{
		TimesheetID: "ts1",
		Date:        "2024-01-02",
		Hours:       -1.0, // Invalid negative hours
		Description: "Test",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestUpdateTimesheetEntry(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/timesheet-entries/e1", func(t *testing.T, body map[string]any) {
		assert.Equal(t, 9.5, body["hours"])
		assert.Equal(t, "Updated description", body["description"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":           "e1",
			"timesheet_id": "ts1",
			"date":         "2024-01-02",
			"hours":        9.5,
			"description":  "Updated description",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateTimesheetEntry(context.Background(), "e1", UpdateTimesheetEntryParams{
		Hours:       9.5,
		Description: "Updated description",
	})

	require.NoError(t, err)
	assert.Equal(t, "e1", result.ID)
	assert.Equal(t, 9.5, result.Hours)
	assert.Equal(t, "Updated description", result.Description)
}

func TestUpdateTimesheetEntry_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/timesheet-entries/invalid", 404, map[string]string{"error": "entry not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateTimesheetEntry(context.Background(), "invalid", UpdateTimesheetEntryParams{
		Hours: 8.0,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestDeleteTimesheetEntry(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/timesheet-entries/e1", 204, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteTimesheetEntry(context.Background(), "e1")

	require.NoError(t, err)
}

func TestDeleteTimesheetEntry_Forbidden(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/timesheet-entries/e1", 403, map[string]string{"error": "forbidden"})
	defer server.Close()

	client := testClient(server)
	err := client.DeleteTimesheetEntry(context.Background(), "e1")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 403, apiErr.StatusCode)
}

func TestReviewTimesheet(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/timesheets/ts1/review", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "approved", body["status"])
		assert.Equal(t, "Looks good!", body["comment"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":           "ts1",
			"contract_id":  "c1",
			"status":       "approved",
			"period_start": "2024-01-01",
			"period_end":   "2024-01-31",
			"total_hours":  160.0,
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.ReviewTimesheet(context.Background(), "ts1", ReviewTimesheetParams{
		Status:  "approved",
		Comment: "Looks good!",
	})

	require.NoError(t, err)
	assert.Equal(t, "ts1", result.ID)
	assert.Equal(t, "approved", result.Status)
}

func TestReviewTimesheet_Unauthorized(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/timesheets/ts1/review", 401, map[string]string{"error": "unauthorized"})
	defer server.Close()

	client := testClient(server)
	_, err := client.ReviewTimesheet(context.Background(), "ts1", ReviewTimesheetParams{
		Status:  "approved",
		Comment: "Looks good!",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 401, apiErr.StatusCode)
}
