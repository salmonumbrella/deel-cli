package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEORAmendment(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/contracts/eor-123/amendments", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "salary_change", body["type"])
		assert.Equal(t, "2024-03-01", body["effective_date"])
		assert.Equal(t, "Annual performance review", body["reason"])
		changes, ok := body["changes"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, 130000.0, changes["salary"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":             "amend-456",
			"contract_id":    "eor-123",
			"type":           "salary_change",
			"status":         "pending",
			"changes":        map[string]any{"salary": 130000.0},
			"effective_date": "2024-03-01",
			"reason":         "Annual performance review",
			"created_at":     "2024-02-01T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateEORAmendment(context.Background(), "eor-123", CreateEORAmendmentParams{
		Type:          "salary_change",
		Changes:       map[string]interface{}{"salary": 130000.0},
		EffectiveDate: "2024-03-01",
		Reason:        "Annual performance review",
	})

	require.NoError(t, err)
	assert.Equal(t, "amend-456", result.ID)
	assert.Equal(t, "eor-123", result.ContractID)
	assert.Equal(t, "salary_change", result.Type)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, 130000.0, result.Changes["salary"])
	assert.Equal(t, "2024-03-01", result.EffectiveDate)
	assert.Equal(t, "Annual performance review", result.Reason)
}

func TestCreateEORAmendment_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/amendments", http.StatusBadRequest, map[string]string{"error": "invalid effective date"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateEORAmendment(context.Background(), "eor-123", CreateEORAmendmentParams{
		Type:          "salary_change",
		Changes:       map[string]interface{}{"salary": 130000.0},
		EffectiveDate: "invalid-date",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestCreateEORAmendment_ContractNotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/invalid/amendments", http.StatusNotFound, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateEORAmendment(context.Background(), "invalid", CreateEORAmendmentParams{
		Type:          "salary_change",
		Changes:       map[string]interface{}{"salary": 130000.0},
		EffectiveDate: "2024-03-01",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestListEORAmendments(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":             "amend-456",
				"contract_id":    "eor-123",
				"type":           "salary_change",
				"status":         "accepted",
				"changes":        map[string]any{"salary": 130000.0},
				"effective_date": "2024-03-01",
				"reason":         "Annual performance review",
				"created_at":     "2024-02-01T10:00:00Z",
				"accepted_at":    "2024-02-02T14:30:00Z",
			},
			{
				"id":             "amend-789",
				"contract_id":    "eor-123",
				"type":           "title_change",
				"status":         "pending",
				"changes":        map[string]any{"job_title": "Staff Engineer"},
				"effective_date": "2024-04-01",
				"reason":         "Promotion",
				"created_at":     "2024-02-15T11:00:00Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/eor-123/amendments", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListEORAmendments(context.Background(), "eor-123")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "amend-456", result[0].ID)
	assert.Equal(t, "salary_change", result[0].Type)
	assert.Equal(t, "accepted", result[0].Status)
	assert.Equal(t, 130000.0, result[0].Changes["salary"])
	assert.Equal(t, "2024-02-02T14:30:00Z", result[0].AcceptedAt)
	assert.Equal(t, "amend-789", result[1].ID)
	assert.Equal(t, "title_change", result[1].Type)
	assert.Equal(t, "pending", result[1].Status)
}

func TestListEORAmendments_Empty(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{},
	}
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/eor-123/amendments", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListEORAmendments(context.Background(), "eor-123")

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestListEORAmendments_ContractNotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/invalid/amendments", http.StatusNotFound, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListEORAmendments(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestAcceptEORAmendment(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/amendments/amend-456/accept", http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":             "amend-456",
			"contract_id":    "eor-123",
			"type":           "salary_change",
			"status":         "accepted",
			"changes":        map[string]any{"salary": 130000.0},
			"effective_date": "2024-03-01",
			"reason":         "Annual performance review",
			"created_at":     "2024-02-01T10:00:00Z",
			"accepted_at":    "2024-02-02T14:30:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.AcceptEORAmendment(context.Background(), "amend-456")

	require.NoError(t, err)
	assert.Equal(t, "amend-456", result.ID)
	assert.Equal(t, "accepted", result.Status)
	assert.Equal(t, "2024-02-02T14:30:00Z", result.AcceptedAt)
}

func TestAcceptEORAmendment_NotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/amendments/invalid/accept", http.StatusNotFound, map[string]string{"error": "amendment not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.AcceptEORAmendment(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestAcceptEORAmendment_AlreadyAccepted(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/amendments/amend-456/accept", http.StatusBadRequest, map[string]string{"error": "amendment already accepted"})
	defer server.Close()

	client := testClient(server)
	_, err := client.AcceptEORAmendment(context.Background(), "amend-456")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestSignEORAmendment(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/amendments/amend-456/sign", http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":             "amend-456",
			"contract_id":    "eor-123",
			"type":           "salary_change",
			"status":         "signed",
			"changes":        map[string]any{"salary": 130000.0},
			"effective_date": "2024-03-01",
			"reason":         "Annual performance review",
			"created_at":     "2024-02-01T10:00:00Z",
			"accepted_at":    "2024-02-02T14:30:00Z",
			"signed_at":      "2024-02-03T09:15:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.SignEORAmendment(context.Background(), "amend-456")

	require.NoError(t, err)
	assert.Equal(t, "amend-456", result.ID)
	assert.Equal(t, "signed", result.Status)
	assert.Equal(t, "2024-02-03T09:15:00Z", result.SignedAt)
}

func TestSignEORAmendment_NotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/amendments/invalid/sign", http.StatusNotFound, map[string]string{"error": "amendment not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.SignEORAmendment(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestSignEORAmendment_NotAccepted(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/amendments/amend-456/sign", http.StatusBadRequest, map[string]string{"error": "amendment must be accepted before signing"})
	defer server.Close()

	client := testClient(server)
	_, err := client.SignEORAmendment(context.Background(), "amend-456")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestSignEORAmendment_AlreadySigned(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/amendments/amend-456/sign", http.StatusBadRequest, map[string]string{"error": "amendment already signed"})
	defer server.Close()

	client := testClient(server)
	_, err := client.SignEORAmendment(context.Background(), "amend-456")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}
