package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVeriffSession(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/screenings/veriff/session", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "worker123", body["worker_id"])
		assert.Equal(t, "https://example.com/callback", body["callback"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":         "session123",
			"url":        "https://veriff.com/verify/session123",
			"status":     "created",
			"expires_at": "2024-12-31T23:59:59Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateVeriffSession(context.Background(), CreateVeriffSessionParams{
		WorkerID: "worker123",
		Callback: "https://example.com/callback",
	})

	require.NoError(t, err)
	assert.Equal(t, "session123", result.ID)
	assert.Equal(t, "https://veriff.com/verify/session123", result.URL)
	assert.Equal(t, "created", result.Status)
	assert.Equal(t, "2024-12-31T23:59:59Z", result.ExpiresAt)
}

func TestGetKYCDetails(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"worker_id":   "worker123",
			"status":      "verified",
			"verified_at": "2024-01-15T10:30:00Z",
			"provider":    "veriff",
			"details": map[string]any{
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-01",
				"country":       "US",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/screenings/kyc/worker123", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetKYCDetails(context.Background(), "worker123")

	require.NoError(t, err)
	assert.Equal(t, "worker123", result.WorkerID)
	assert.Equal(t, "verified", result.Status)
	assert.Equal(t, "2024-01-15T10:30:00Z", result.VerifiedAt)
	assert.Equal(t, "veriff", result.Provider)
	assert.Equal(t, "John", result.Details.FirstName)
	assert.Equal(t, "Doe", result.Details.LastName)
	assert.Equal(t, "1990-01-01", result.Details.DateOfBirth)
	assert.Equal(t, "US", result.Details.Country)
}

func TestGetAMLData(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"results": []map[string]any{
				{
					"name":          "John Doe",
					"country":       "US",
					"match_type":    "exact",
					"risk_level":    "low",
					"screened_at":   "2024-01-15T10:30:00Z",
					"list_name":     "OFAC",
					"match_details": "No matches found",
				},
			},
			"summary": map[string]any{
				"total_matches": 0,
				"highest_risk":  "low",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/screenings/aml", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetAMLData(context.Background())

	require.NoError(t, err)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, "John Doe", result.Results[0].Name)
	assert.Equal(t, "US", result.Results[0].Country)
	assert.Equal(t, "exact", result.Results[0].MatchType)
	assert.Equal(t, "low", result.Results[0].RiskLevel)
	assert.Equal(t, "2024-01-15T10:30:00Z", result.Results[0].ScreenedAt)
	assert.Equal(t, "OFAC", result.Results[0].ListName)
	assert.Equal(t, 0, result.Summary.TotalMatches)
	assert.Equal(t, "low", result.Summary.HighestRisk)
}

func TestSubmitExternalKYC(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/screenings/kyc/external", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "worker123", body["worker_id"])
		assert.Equal(t, "external_provider", body["provider"])
		assert.Equal(t, "2024-01-15T10:30:00Z", body["verified_at"])
		assert.Equal(t, "passport", body["document_type"])
		assert.Equal(t, "P12345678", body["document_id"])
		assert.Equal(t, "2030-12-31", body["expiration_date"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":           "sub123",
			"worker_id":    "worker123",
			"status":       "pending",
			"submitted_at": "2024-01-15T10:35:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.SubmitExternalKYC(context.Background(), SubmitExternalKYCParams{
		WorkerID:       "worker123",
		Provider:       "external_provider",
		VerifiedAt:     "2024-01-15T10:30:00Z",
		DocumentType:   "passport",
		DocumentID:     "P12345678",
		ExpirationDate: "2030-12-31",
	})

	require.NoError(t, err)
	assert.Equal(t, "sub123", result.ID)
	assert.Equal(t, "worker123", result.WorkerID)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "2024-01-15T10:35:00Z", result.SubmittedAt)
}

func TestCreateManualVerification(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/screenings/verification/manual", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "worker123", body["worker_id"])
		assert.Equal(t, "admin@company.com", body["verified_by"])
		assert.Equal(t, "Manual verification completed", body["notes"])
		docs := body["document_urls"].([]any)
		assert.Len(t, docs, 2)
		assert.Contains(t, docs, "https://example.com/doc1.pdf")
		assert.Contains(t, docs, "https://example.com/doc2.pdf")
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":          "ver123",
			"worker_id":   "worker123",
			"verified_by": "admin@company.com",
			"status":      "completed",
			"notes":       "Manual verification completed",
			"created_at":  "2024-01-15T10:30:00Z",
			"document_urls": []string{
				"https://example.com/doc1.pdf",
				"https://example.com/doc2.pdf",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateManualVerification(context.Background(), CreateManualVerificationParams{
		WorkerID:   "worker123",
		VerifiedBy: "admin@company.com",
		Notes:      "Manual verification completed",
		DocumentURLs: []string{
			"https://example.com/doc1.pdf",
			"https://example.com/doc2.pdf",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "ver123", result.ID)
	assert.Equal(t, "worker123", result.WorkerID)
	assert.Equal(t, "admin@company.com", result.VerifiedBy)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, "Manual verification completed", result.Notes)
	assert.Equal(t, "2024-01-15T10:30:00Z", result.CreatedAt)
	assert.Len(t, result.DocumentURLs, 2)
	assert.Contains(t, result.DocumentURLs, "https://example.com/doc1.pdf")
	assert.Contains(t, result.DocumentURLs, "https://example.com/doc2.pdf")
}
