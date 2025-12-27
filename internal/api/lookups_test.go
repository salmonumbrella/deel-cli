package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCurrencies(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"code": "USD", "name": "US Dollar", "symbol": "$"},
			{"code": "EUR", "name": "Euro", "symbol": "€"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/lookups/currencies", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListCurrencies(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "USD", result[0].Code)
	assert.Equal(t, "US Dollar", result[0].Name)
	assert.Equal(t, "$", result[0].Symbol)
	assert.Equal(t, "EUR", result[1].Code)
}

func TestListCountries(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"code": "US", "name": "United States"},
			{"code": "GB", "name": "United Kingdom"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/lookups/countries", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListCountries(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "US", result[0].Code)
	assert.Equal(t, "United States", result[0].Name)
	assert.Equal(t, "GB", result[1].Code)
}

func TestListJobTitles(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"id": "jt1", "name": "Software Engineer"},
			{"id": "jt2", "name": "Product Manager"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/lookups/job-titles", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListJobTitles(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "jt1", result[0].ID)
	assert.Equal(t, "Software Engineer", result[0].Name)
	assert.Equal(t, "jt2", result[1].ID)
}

func TestListSeniorityLevels(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"id": "sl1", "name": "Junior"},
			{"id": "sl2", "name": "Mid-Level"},
			{"id": "sl3", "name": "Senior"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/lookups/seniority-levels", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListSeniorityLevels(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "sl1", result[0].ID)
	assert.Equal(t, "Junior", result[0].Name)
	assert.Equal(t, "sl2", result[1].ID)
	assert.Equal(t, "Mid-Level", result[1].Name)
}

func TestListTimeOffTypes(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"id": "tot1", "name": "Vacation"},
			{"id": "tot2", "name": "Sick Leave"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/lookups/time-off-types", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListTimeOffTypes(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "tot1", result[0].ID)
	assert.Equal(t, "Vacation", result[0].Name)
	assert.Equal(t, "tot2", result[1].ID)
}

func TestListCurrencies_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/lookups/currencies", 404, map[string]any{
		"error": "not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListCurrencies(context.Background())

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestListCountries_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/lookups/countries", 404, map[string]any{
		"error": "not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListCountries(context.Background())

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestListJobTitles_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/lookups/job-titles", 404, map[string]any{
		"error": "not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListJobTitles(context.Background())

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestListSeniorityLevels_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/lookups/seniority-levels", 404, map[string]any{
		"error": "not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListSeniorityLevels(context.Background())

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestListTimeOffTypes_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/lookups/time-off-types", 404, map[string]any{
		"error": "not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListTimeOffTypes(context.Background())

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}
