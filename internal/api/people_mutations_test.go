package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePerson(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/people", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "john.doe@example.com", body["email"])
		assert.Equal(t, "John", body["first_name"])
		assert.Equal(t, "Doe", body["last_name"])
		assert.Equal(t, "contractor", body["type"])
		assert.Equal(t, "US", body["country"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":         "p-123",
			"email":      "john.doe@example.com",
			"first_name": "John",
			"last_name":  "Doe",
			"type":       "contractor",
			"country":    "US",
			"status":     "active",
			"created_at": "2025-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreatePerson(context.Background(), CreatePersonParams{
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Type:      "contractor",
		Country:   "US",
	})

	require.NoError(t, err)
	assert.Equal(t, "p-123", result.ID)
	assert.Equal(t, "john.doe@example.com", result.Email)
	assert.Equal(t, "John", result.FirstName)
	assert.Equal(t, "Doe", result.LastName)
	assert.Equal(t, "contractor", result.Type)
	assert.Equal(t, "US", result.Country)
	assert.Equal(t, "active", result.Status)
}

func TestUpdatePersonalInfo(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/people/p-123/personal-info", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Jane", body["first_name"])
		assert.Equal(t, "Smith", body["last_name"])
		assert.Equal(t, "1990-05-15", body["date_of_birth"])
		assert.Equal(t, "+1234567890", body["phone"])
		assert.Equal(t, "US", body["nationality"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":            "p-123",
			"first_name":    "Jane",
			"last_name":     "Smith",
			"date_of_birth": "1990-05-15",
			"phone":         "+1234567890",
			"nationality":   "US",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdatePersonalInfo(context.Background(), "p-123", PersonalInfo{
		FirstName:   "Jane",
		LastName:    "Smith",
		DateOfBirth: "1990-05-15",
		Phone:       "+1234567890",
		Nationality: "US",
	})

	require.NoError(t, err)
	assert.Equal(t, "p-123", result.ID)
	assert.Equal(t, "Jane", result.FirstName)
	assert.Equal(t, "Smith", result.LastName)
	assert.Equal(t, "1990-05-15", result.DateOfBirth)
	assert.Equal(t, "+1234567890", result.Phone)
	assert.Equal(t, "US", result.Nationality)
}

func TestUpdateWorkingLocation(t *testing.T) {
	server := mockServerWithBody(t, "PUT", "/rest/v2/people/p-123/working-location", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "US", body["country"])
		assert.Equal(t, "CA", body["state"])
		assert.Equal(t, "San Francisco", body["city"])
		assert.Equal(t, "123 Market St", body["address"])
		assert.Equal(t, "94103", body["postal_code"])
		assert.Equal(t, "America/Los_Angeles", body["timezone"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":          "p-123",
			"country":     "US",
			"state":       "CA",
			"city":        "San Francisco",
			"address":     "123 Market St",
			"postal_code": "94103",
			"timezone":    "America/Los_Angeles",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateWorkingLocation(context.Background(), "p-123", WorkingLocation{
		Country:    "US",
		State:      "CA",
		City:       "San Francisco",
		Address:    "123 Market St",
		PostalCode: "94103",
		Timezone:   "America/Los_Angeles",
	})

	require.NoError(t, err)
	assert.Equal(t, "p-123", result.ID)
	assert.Equal(t, "US", result.Country)
	assert.Equal(t, "CA", result.State)
	assert.Equal(t, "San Francisco", result.City)
	assert.Equal(t, "123 Market St", result.Address)
	assert.Equal(t, "94103", result.PostalCode)
	assert.Equal(t, "America/Los_Angeles", result.Timezone)
}

func TestCreateDirectEmployee(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/people/direct-employee", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "employee@example.com", body["email"])
		assert.Equal(t, "Alice", body["first_name"])
		assert.Equal(t, "Johnson", body["last_name"])
		assert.Equal(t, "US", body["country"])
		assert.Equal(t, "2025-02-01", body["start_date"])
		assert.Equal(t, "Software Engineer", body["job_title"])
		assert.Equal(t, "Engineering", body["department"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":         "p-456",
			"email":      "employee@example.com",
			"first_name": "Alice",
			"last_name":  "Johnson",
			"country":    "US",
			"start_date": "2025-02-01",
			"job_title":  "Software Engineer",
			"department": "Engineering",
			"type":       "employee",
			"status":     "active",
			"created_at": "2025-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateDirectEmployee(context.Background(), CreateDirectEmployeeParams{
		Email:      "employee@example.com",
		FirstName:  "Alice",
		LastName:   "Johnson",
		Country:    "US",
		StartDate:  "2025-02-01",
		JobTitle:   "Software Engineer",
		Department: "Engineering",
	})

	require.NoError(t, err)
	assert.Equal(t, "p-456", result.ID)
	assert.Equal(t, "employee@example.com", result.Email)
	assert.Equal(t, "Alice", result.FirstName)
	assert.Equal(t, "Johnson", result.LastName)
	assert.Equal(t, "US", result.Country)
	assert.Equal(t, "2025-02-01", result.StartDate)
	assert.Equal(t, "Software Engineer", result.JobTitle)
	assert.Equal(t, "Engineering", result.Department)
	assert.Equal(t, "employee", result.Type)
	assert.Equal(t, "active", result.Status)
}
