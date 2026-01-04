package api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonUnmarshalJSON_NameComputation(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		expectedName string
	}{
		{
			name:         "both first and last name present",
			json:         `{"first_name": "John", "last_name": "Doe"}`,
			expectedName: "John Doe",
		},
		{
			name:         "only first name present",
			json:         `{"first_name": "John", "last_name": ""}`,
			expectedName: "John",
		},
		{
			name:         "only last name present",
			json:         `{"first_name": "", "last_name": "Doe"}`,
			expectedName: "Doe",
		},
		{
			name:         "both names empty",
			json:         `{"first_name": "", "last_name": ""}`,
			expectedName: "",
		},
		{
			name:         "first name only - last name missing from JSON",
			json:         `{"first_name": "Jane"}`,
			expectedName: "Jane",
		},
		{
			name:         "last name only - first name missing from JSON",
			json:         `{"last_name": "Smith"}`,
			expectedName: "Smith",
		},
		{
			name:         "neither name in JSON",
			json:         `{"id": "123"}`,
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var person Person
			err := json.Unmarshal([]byte(tt.json), &person)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedName, person.Name)
		})
	}
}

func TestGetPersonPersonal(t *testing.T) {
	// Personal endpoint returns varied data - test that raw JSON is preserved
	personalData := map[string]any{
		"worker_id":    12345,
		"first_name":   "John",
		"last_name":    "Doe",
		"email":        "john.doe@example.com",
		"phone_number": "+1-555-123-4567",
	}
	response := map[string]any{
		"data": personalData,
	}
	server := mockServer(t, "GET", "/rest/v2/people/hris-123/personal", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetPersonPersonal(context.Background(), "hris-123")

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify we can unmarshal the raw JSON back to a map
	var parsed map[string]any
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)

	// Check that numeric worker_id is preserved
	assert.Equal(t, float64(12345), parsed["worker_id"])
	assert.Equal(t, "John", parsed["first_name"])
	assert.Equal(t, "Doe", parsed["last_name"])
	assert.Equal(t, "john.doe@example.com", parsed["email"])
	assert.Equal(t, "+1-555-123-4567", parsed["phone_number"])
}

func TestGetPersonPersonal_Error(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/people/nonexistent/personal", 404, map[string]any{
		"error": "person not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetPersonPersonal(context.Background(), "nonexistent")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestGetPersonPersonal_EmptyData(t *testing.T) {
	// API returns 200 but with no data field
	response := map[string]any{}
	server := mockServer(t, "GET", "/rest/v2/people/hris-empty/personal", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetPersonPersonal(context.Background(), "hris-empty")

	require.NoError(t, err)
	// When data field is missing, result should be nil or empty
	assert.Nil(t, result)
}
