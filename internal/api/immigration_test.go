package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetImmigrationCaseDetails(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":          "case-123",
			"contract_id": "contract-456",
			"worker_name": "John Doe",
			"type":        "work_permit",
			"status":      "in_progress",
			"country":     "US",
			"start_date":  "2024-01-15",
			"expiry_date": "2025-01-15",
			"case_number": "WP-2024-001",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/immigration/cases/case-123", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetImmigrationCaseDetails(context.Background(), "case-123")

	require.NoError(t, err)
	assert.Equal(t, "case-123", result.ID)
	assert.Equal(t, "contract-456", result.ContractID)
	assert.Equal(t, "John Doe", result.WorkerName)
	assert.Equal(t, "work_permit", result.Type)
	assert.Equal(t, "in_progress", result.Status)
	assert.Equal(t, "US", result.Country)
}

func TestListImmigrationDocs(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":         "doc-1",
				"case_id":    "case-123",
				"name":       "passport.pdf",
				"type":       "passport",
				"status":     "approved",
				"expires_at": "2030-06-15",
			},
			{
				"id":         "doc-2",
				"case_id":    "case-123",
				"name":       "employment_letter.pdf",
				"type":       "employment_letter",
				"status":     "pending",
				"expires_at": "",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/immigration/cases/case-123/documents", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListImmigrationDocs(context.Background(), "case-123")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "doc-1", result[0].ID)
	assert.Equal(t, "passport", result[0].Type)
	assert.Equal(t, "doc-2", result[1].ID)
	assert.Equal(t, "employment_letter", result[1].Type)
}

func TestListVisaTypes(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":                   "visa-h1b",
				"name":                 "H-1B",
				"country":              "US",
				"category":             "work",
				"max_duration":         "3 years",
				"requirements_summary": "Bachelor's degree required",
			},
			{
				"id":                   "visa-l1",
				"name":                 "L-1",
				"country":              "US",
				"category":             "intracompany_transfer",
				"max_duration":         "7 years",
				"requirements_summary": "Manager or specialized knowledge",
			},
		},
	}
	server := mockServerWithQuery(t, "GET", "/rest/v2/immigration/visa-types", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "US", query["country"])
	}, 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListVisaTypes(context.Background(), "US")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "visa-h1b", result[0].ID)
	assert.Equal(t, "H-1B", result[0].Name)
	assert.Equal(t, "work", result[0].Category)
	assert.Equal(t, "visa-l1", result[1].ID)
}

func TestCheckVisaRequirement(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"visa_required":  true,
			"suggested_type": "work_permit",
			"max_stay":       "90 days",
			"notes":          "Work permit required for employment",
		},
	}
	server := mockServerWithQuery(t, "GET", "/rest/v2/immigration/check", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "CA", query["from"])
		assert.Equal(t, "US", query["to"])
	}, 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.CheckVisaRequirement(context.Background(), "CA", "US")

	require.NoError(t, err)
	assert.True(t, result.Required)
	assert.Equal(t, "work_permit", result.Type)
	assert.Equal(t, "90 days", result.Duration)
	assert.Equal(t, "Work permit required for employment", result.Notes)
}

func TestCreateImmigrationCase(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/immigration/cases", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "contract-456", body["contract_id"])
		assert.Equal(t, "work_permit", body["type"])
		assert.Equal(t, "US", body["country"])
		assert.Equal(t, "2024-02-01", body["start_date"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":          "case-new",
			"contract_id": "contract-456",
			"worker_name": "Jane Smith",
			"type":        "work_permit",
			"status":      "pending",
			"country":     "US",
			"start_date":  "2024-02-01",
			"expiry_date": "",
			"case_number": "WP-2024-002",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateImmigrationCase(context.Background(), CreateImmigrationCaseParams{
		ContractID: "contract-456",
		Type:       "work_permit",
		Country:    "US",
		StartDate:  "2024-02-01",
	})

	require.NoError(t, err)
	assert.Equal(t, "case-new", result.ID)
	assert.Equal(t, "contract-456", result.ContractID)
	assert.Equal(t, "Jane Smith", result.WorkerName)
	assert.Equal(t, "work_permit", result.Type)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "WP-2024-002", result.CaseNumber)
}

func TestUploadImmigrationDocument(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/immigration/cases/case-123/documents", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "passport_copy.pdf", body["name"])
		assert.Equal(t, "passport", body["type"])
		assert.Equal(t, "https://storage.deel.com/docs/abc123", body["document_url"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":         "doc-new",
			"case_id":    "case-123",
			"name":       "passport_copy.pdf",
			"type":       "passport",
			"status":     "pending_review",
			"expires_at": "2030-06-15",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UploadImmigrationDocument(context.Background(), "case-123", UploadImmigrationDocumentParams{
		Name:        "passport_copy.pdf",
		Type:        "passport",
		DocumentURL: "https://storage.deel.com/docs/abc123",
	})

	require.NoError(t, err)
	assert.Equal(t, "doc-new", result.ID)
	assert.Equal(t, "case-123", result.CaseID)
	assert.Equal(t, "passport_copy.pdf", result.Name)
	assert.Equal(t, "passport", result.Type)
	assert.Equal(t, "pending_review", result.Status)
	assert.Equal(t, "2030-06-15", result.ExpiresAt)
}
