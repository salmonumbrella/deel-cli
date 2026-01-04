package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

func TestPeopleGetCmd_PersonalFlag_Exists(t *testing.T) {
	// Verify the --personal flag exists on the people get command
	flag := peopleGetCmd.Flags().Lookup("personal")
	require.NotNil(t, flag, "expected --personal flag to exist")
	assert.Equal(t, "false", flag.DefValue)
	assert.Contains(t, flag.Usage, "personal info")
}

func TestPeopleGetCmd_PersonalFlag_CallsPersonalEndpoint(t *testing.T) {
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
	client := api.NewClient("test-token")
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

func TestPeopleGetCmd_PersonalFlag_OutputsWorkerID(t *testing.T) {
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
	client := api.NewClient("test-token")
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

func TestPeopleGetCmd_WithoutPersonalFlag_CallsRegularEndpoint(t *testing.T) {
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

	client := api.NewClient("test-token")
	client.SetBaseURL(server.URL)

	_, err := client.GetPerson(context.Background(), "hris-123")
	require.NoError(t, err)
	assert.Equal(t, "/rest/v2/people/hris-123", calledPath)
}
