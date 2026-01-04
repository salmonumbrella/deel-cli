package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCandidate(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/candidates", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "John", body["first_name"])
		assert.Equal(t, "Doe", body["last_name"])
		assert.Equal(t, "john.doe@example.com", body["email"])
		assert.Equal(t, "+1234567890", body["phone"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":         "cand-123",
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john.doe@example.com",
			"phone":      "+1234567890",
			"status":     "new",
			"created_at": "2024-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.AddCandidate(context.Background(), AddCandidateParams{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+1234567890",
	})

	require.NoError(t, err)
	assert.Equal(t, "cand-123", result.ID)
	assert.Equal(t, "John", result.FirstName)
	assert.Equal(t, "Doe", result.LastName)
	assert.Equal(t, "john.doe@example.com", result.Email)
	assert.Equal(t, "+1234567890", result.Phone)
	assert.Equal(t, "new", result.Status)
	assert.Equal(t, "2024-01-01T00:00:00Z", result.CreatedAt)
}

func TestAddCandidate_WithoutPhone(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/candidates", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Jane", body["first_name"])
		assert.Equal(t, "Smith", body["last_name"])
		assert.Equal(t, "jane.smith@example.com", body["email"])
		_, hasPhone := body["phone"]
		assert.False(t, hasPhone)
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":         "cand-456",
			"first_name": "Jane",
			"last_name":  "Smith",
			"email":      "jane.smith@example.com",
			"status":     "new",
			"created_at": "2024-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.AddCandidate(context.Background(), AddCandidateParams{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
	})

	require.NoError(t, err)
	assert.Equal(t, "cand-456", result.ID)
	assert.Equal(t, "Jane", result.FirstName)
	assert.Equal(t, "", result.Phone)
}

func TestUpdateCandidate(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/candidates/cand-123", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "interviewed", body["status"])
		assert.Equal(t, "+9876543210", body["phone"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":         "cand-123",
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john.doe@example.com",
			"phone":      "+9876543210",
			"status":     "interviewed",
			"created_at": "2024-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateCandidate(context.Background(), "cand-123", UpdateCandidateParams{
		Status: "interviewed",
		Phone:  "+9876543210",
	})

	require.NoError(t, err)
	assert.Equal(t, "cand-123", result.ID)
	assert.Equal(t, "interviewed", result.Status)
	assert.Equal(t, "+9876543210", result.Phone)
}

func TestUpdateCandidate_PartialUpdate(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/candidates/cand-456", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "hired", body["status"])
		// Ensure other fields are not sent
		_, hasFirstName := body["first_name"]
		assert.False(t, hasFirstName)
		_, hasLastName := body["last_name"]
		assert.False(t, hasLastName)
		_, hasEmail := body["email"]
		assert.False(t, hasEmail)
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":         "cand-456",
			"first_name": "Jane",
			"last_name":  "Smith",
			"email":      "jane.smith@example.com",
			"status":     "hired",
			"created_at": "2024-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateCandidate(context.Background(), "cand-456", UpdateCandidateParams{
		Status: "hired",
	})

	require.NoError(t, err)
	assert.Equal(t, "cand-456", result.ID)
	assert.Equal(t, "hired", result.Status)
}
