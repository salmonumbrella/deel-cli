package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestGetPersonPersonal_CallsCorrectEndpoint(t *testing.T) {
	// Track which endpoint was called
	var calledPath string
	var calledMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		calledMethod = r.Method

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		// Return appropriate response based on endpoint
		if strings.HasSuffix(r.URL.Path, "/personal") {
			// Personal endpoint response
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":         "personal-123",
					"worker_id":  12345,
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john@example.com",
				},
			})
		} else {
			// Regular endpoint response
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"hris_profile_id": "hris-123",
					"first_name":      "John",
					"last_name":       "Doe",
					"email":           "john@example.com",
					"job_title":       "Engineer",
					"status":          "active",
					"country":         "US",
					"start_date":      "2024-01-01",
				},
			})
		}
	}))
	defer server.Close()

	// Create client pointing to test server
	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	// Test that GetPersonPersonal calls the /personal endpoint
	_, err := client.GetPersonPersonal(context.Background(), "hris-123")
	require.NoError(t, err)
	assert.Equal(t, "GET", calledMethod)
	assert.Equal(t, "/rest/v2/people/hris-123/personal", calledPath)

	// Reset for next call
	calledPath = ""
	calledMethod = ""

	// Test that GetPerson calls the regular endpoint
	_, err = client.GetPerson(context.Background(), "hris-123")
	require.NoError(t, err)
	assert.Equal(t, "GET", calledMethod)
	assert.Equal(t, "/rest/v2/people/hris-123", calledPath)
}

func TestGetPersonPersonal_ReturnsWorkerID(t *testing.T) {
	// Set up mock server that returns personal data with worker_id
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		if strings.HasSuffix(r.URL.Path, "/personal") {
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":         "personal-123",
					"worker_id":  12345,
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john@example.com",
				},
			})
		}
	}))
	defer server.Close()

	// Create client and call GetPersonPersonal
	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	rawData, err := client.GetPersonPersonal(context.Background(), "hris-123")
	require.NoError(t, err)

	// Parse the response
	var data map[string]any
	err = json.Unmarshal(rawData, &data)
	require.NoError(t, err)

	// Verify worker_id is present and correct
	assert.Equal(t, float64(12345), data["worker_id"])
	assert.Equal(t, "John", data["first_name"])
	assert.Equal(t, "Doe", data["last_name"])
}

func TestGetPerson_CallsCorrectEndpoint(t *testing.T) {
	var calledPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"hris_profile_id": "hris-123",
				"first_name":      "John",
				"last_name":       "Doe",
				"email":           "john@example.com",
				"job_title":       "Engineer",
				"status":          "active",
				"country":         "US",
				"start_date":      "2024-01-01",
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.SetBaseURL(server.URL)

	_, err := client.GetPerson(context.Background(), "hris-123")
	require.NoError(t, err)
	assert.Equal(t, "/rest/v2/people/hris-123", calledPath)
}
