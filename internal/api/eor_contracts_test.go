package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEORContract(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/contracts", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Senior Software Engineer - Full Stack", body["title"])
		assert.Equal(t, "john.doe@example.com", body["worker_email"])
		assert.Equal(t, "John Doe", body["worker_name"])
		assert.Equal(t, "US", body["country"])
		assert.Equal(t, "2024-02-01", body["start_date"])
		assert.Equal(t, 120000.0, body["salary"])
		assert.Equal(t, "USD", body["currency"])
		assert.Equal(t, "monthly", body["pay_frequency"])
		assert.Equal(t, "Software Engineer", body["job_title"])
		assert.Equal(t, "senior", body["seniority_level"])
		assert.Equal(t, "Full-stack development", body["scope"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":              "eor-123",
			"title":           "Senior Software Engineer - Full Stack",
			"status":          "draft",
			"worker_id":       "w-456",
			"worker_email":    "john.doe@example.com",
			"worker_name":     "John Doe",
			"country":         "US",
			"start_date":      "2024-02-01",
			"salary":          120000.0,
			"currency":        "USD",
			"pay_frequency":   "monthly",
			"job_title":       "Software Engineer",
			"seniority_level": "senior",
			"scope":           "Full-stack development",
			"created_at":      "2024-01-15T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateEORContract(context.Background(), CreateEORContractParams{
		Title:          "Senior Software Engineer - Full Stack",
		WorkerEmail:    "john.doe@example.com",
		WorkerName:     "John Doe",
		Country:        "US",
		StartDate:      "2024-02-01",
		Salary:         120000.0,
		Currency:       "USD",
		PayFrequency:   "monthly",
		JobTitle:       "Software Engineer",
		SeniorityLevel: "senior",
		Scope:          "Full-stack development",
	})

	require.NoError(t, err)
	assert.Equal(t, "eor-123", result.ID)
	assert.Equal(t, "draft", result.Status)
	assert.Equal(t, "john.doe@example.com", result.WorkerEmail)
	assert.Equal(t, 120000.0, result.Salary)
	assert.Equal(t, "USD", result.Currency)
}

func TestCreateEORContract_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts", 400, map[string]string{"error": "invalid country code"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateEORContract(context.Background(), CreateEORContractParams{
		Title:        "Engineer",
		WorkerEmail:  "test@example.com",
		WorkerName:   "Test User",
		Country:      "INVALID",
		StartDate:    "2024-02-01",
		Salary:       100000.0,
		Currency:     "USD",
		PayFrequency: "monthly",
		JobTitle:     "Engineer",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestGetEORContract(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":              "eor-123",
			"title":           "Senior Software Engineer",
			"status":          "active",
			"worker_id":       "w-456",
			"worker_email":    "john.doe@example.com",
			"worker_name":     "John Doe",
			"country":         "US",
			"start_date":      "2024-02-01",
			"end_date":        "2025-02-01",
			"salary":          120000.0,
			"currency":        "USD",
			"pay_frequency":   "monthly",
			"job_title":       "Software Engineer",
			"seniority_level": "senior",
			"scope":           "Full-stack development",
			"benefits": []map[string]any{
				{
					"id":          "b-1",
					"name":        "Health Insurance",
					"description": "Comprehensive health coverage",
					"amount":      500.0,
				},
			},
			"created_at": "2024-01-15T10:00:00Z",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/eor-123", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetEORContract(context.Background(), "eor-123")

	require.NoError(t, err)
	assert.Equal(t, "eor-123", result.ID)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "john.doe@example.com", result.WorkerEmail)
	assert.Equal(t, 120000.0, result.Salary)
	assert.Len(t, result.Benefits, 1)
	assert.Equal(t, "Health Insurance", result.Benefits[0].Name)
	assert.Equal(t, 500.0, result.Benefits[0].Amount)
}

func TestGetEORContract_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/contracts/invalid", 404, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetEORContract(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestUpdateEORContract(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/eor/contracts/eor-123", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Principal Software Engineer", body["title"])
		assert.Equal(t, 140000.0, body["salary"])
		assert.Equal(t, "Staff Engineer", body["job_title"])
		assert.Equal(t, "staff", body["seniority_level"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":              "eor-123",
			"title":           "Principal Software Engineer",
			"status":          "active",
			"worker_id":       "w-456",
			"worker_email":    "john.doe@example.com",
			"worker_name":     "John Doe",
			"country":         "US",
			"start_date":      "2024-02-01",
			"salary":          140000.0,
			"currency":        "USD",
			"pay_frequency":   "monthly",
			"job_title":       "Staff Engineer",
			"seniority_level": "staff",
			"scope":           "Full-stack development",
			"created_at":      "2024-01-15T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateEORContract(context.Background(), "eor-123", UpdateEORContractParams{
		Title:          "Principal Software Engineer",
		Salary:         140000.0,
		JobTitle:       "Staff Engineer",
		SeniorityLevel: "staff",
	})

	require.NoError(t, err)
	assert.Equal(t, "eor-123", result.ID)
	assert.Equal(t, "Principal Software Engineer", result.Title)
	assert.Equal(t, 140000.0, result.Salary)
	assert.Equal(t, "Staff Engineer", result.JobTitle)
	assert.Equal(t, "staff", result.SeniorityLevel)
}

func TestUpdateEORContract_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/eor/contracts/invalid", 404, map[string]string{"error": "contract not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateEORContract(context.Background(), "invalid", UpdateEORContractParams{
		Salary: 150000.0,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestCancelEORContract(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/contracts/eor-123/cancel", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Employee resigned", body["reason"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":              "eor-123",
			"title":           "Senior Software Engineer",
			"status":          "cancelled",
			"worker_id":       "w-456",
			"worker_email":    "john.doe@example.com",
			"worker_name":     "John Doe",
			"country":         "US",
			"start_date":      "2024-02-01",
			"end_date":        "2024-06-30",
			"salary":          120000.0,
			"currency":        "USD",
			"pay_frequency":   "monthly",
			"job_title":       "Software Engineer",
			"seniority_level": "senior",
			"created_at":      "2024-01-15T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CancelEORContract(context.Background(), "eor-123", CancelEORContractParams{
		Reason: "Employee resigned",
	})

	require.NoError(t, err)
	assert.Equal(t, "eor-123", result.ID)
	assert.Equal(t, "cancelled", result.Status)
	assert.Equal(t, "2024-06-30", result.EndDate)
}

func TestCancelEORContract_Forbidden(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/cancel", 403, map[string]string{"error": "cannot cancel active contract"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CancelEORContract(context.Background(), "eor-123", CancelEORContractParams{
		Reason: "Test",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 403, apiErr.StatusCode)
}

func TestSignEORContract(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/sign", 200, map[string]any{
		"data": map[string]any{
			"id":              "eor-123",
			"title":           "Senior Software Engineer",
			"status":          "signed",
			"worker_id":       "w-456",
			"worker_email":    "john.doe@example.com",
			"worker_name":     "John Doe",
			"country":         "US",
			"start_date":      "2024-02-01",
			"salary":          120000.0,
			"currency":        "USD",
			"pay_frequency":   "monthly",
			"job_title":       "Software Engineer",
			"seniority_level": "senior",
			"created_at":      "2024-01-15T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.SignEORContract(context.Background(), "eor-123")

	require.NoError(t, err)
	assert.Equal(t, "eor-123", result.ID)
	assert.Equal(t, "signed", result.Status)
}

func TestSignEORContract_AlreadySigned(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/contracts/eor-123/sign", 400, map[string]string{"error": "contract already signed"})
	defer server.Close()

	client := testClient(server)
	_, err := client.SignEORContract(context.Background(), "eor-123")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}
