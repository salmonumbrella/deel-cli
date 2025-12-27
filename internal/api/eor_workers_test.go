package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEORWorker(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/workers", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "jane.smith@example.com", body["email"])
		assert.Equal(t, "Jane", body["first_name"])
		assert.Equal(t, "Smith", body["last_name"])
		assert.Equal(t, "GB", body["country"])
		assert.Equal(t, "1990-05-15", body["date_of_birth"])
		assert.Equal(t, "+44 20 1234 5678", body["phone"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":            "worker-789",
			"email":         "jane.smith@example.com",
			"first_name":    "Jane",
			"last_name":     "Smith",
			"country":       "GB",
			"date_of_birth": "1990-05-15",
			"phone":         "+44 20 1234 5678",
			"status":        "pending",
			"created_at":    "2024-01-20T12:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateEORWorker(context.Background(), CreateEORWorkerParams{
		Email:       "jane.smith@example.com",
		FirstName:   "Jane",
		LastName:    "Smith",
		Country:     "GB",
		DateOfBirth: "1990-05-15",
		Phone:       "+44 20 1234 5678",
	})

	require.NoError(t, err)
	assert.Equal(t, "worker-789", result.ID)
	assert.Equal(t, "jane.smith@example.com", result.Email)
	assert.Equal(t, "Jane", result.FirstName)
	assert.Equal(t, "Smith", result.LastName)
	assert.Equal(t, "GB", result.Country)
	assert.Equal(t, "pending", result.Status)
}

func TestCreateEORWorker_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/workers", 400, map[string]string{"error": "invalid email"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateEORWorker(context.Background(), CreateEORWorkerParams{
		Email:     "invalid-email",
		FirstName: "Test",
		LastName:  "User",
		Country:   "US",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestUpdateEORWorker(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/eor/workers/worker-789", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Janet", body["first_name"])
		assert.Equal(t, "+44 20 9876 5432", body["phone"])
		assert.Equal(t, "123 High Street, London", body["address"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":            "worker-789",
			"email":         "jane.smith@example.com",
			"first_name":    "Janet",
			"last_name":     "Smith",
			"country":       "GB",
			"date_of_birth": "1990-05-15",
			"phone":         "+44 20 9876 5432",
			"address":       "123 High Street, London",
			"status":        "active",
			"created_at":    "2024-01-20T12:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateEORWorker(context.Background(), "worker-789", UpdateEORWorkerParams{
		FirstName: "Janet",
		Phone:     "+44 20 9876 5432",
		Address:   "123 High Street, London",
	})

	require.NoError(t, err)
	assert.Equal(t, "worker-789", result.ID)
	assert.Equal(t, "Janet", result.FirstName)
	assert.Equal(t, "+44 20 9876 5432", result.Phone)
	assert.Equal(t, "123 High Street, London", result.Address)
}

func TestUpdateEORWorker_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/eor/workers/invalid", 404, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateEORWorker(context.Background(), "invalid", UpdateEORWorkerParams{
		Phone: "+1 555 1234",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestGetEORWorkerBenefits(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "benefit-1",
				"name":        "Health Insurance",
				"type":        "health",
				"description": "Comprehensive health coverage",
				"amount":      350.50,
				"currency":    "GBP",
				"status":      "active",
			},
			{
				"id":          "benefit-2",
				"name":        "Pension Contribution",
				"type":        "pension",
				"description": "Employer pension contribution",
				"amount":      500.00,
				"currency":    "GBP",
				"status":      "active",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/eor/workers/worker-789/benefits", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetEORWorkerBenefits(context.Background(), "worker-789")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "benefit-1", result[0].ID)
	assert.Equal(t, "Health Insurance", result[0].Name)
	assert.Equal(t, "health", result[0].Type)
	assert.Equal(t, 350.50, result[0].Amount)
	assert.Equal(t, "GBP", result[0].Currency)
	assert.Equal(t, "active", result[0].Status)
}

func TestGetEORWorkerBenefits_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/workers/invalid/benefits", 404, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetEORWorkerBenefits(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestGetEORWorkerTaxDocuments(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":           "tax-doc-1",
				"name":         "W-2 Form 2023",
				"type":         "w2",
				"year":         2023,
				"status":       "available",
				"download_url": "https://api.deel.com/documents/tax-doc-1",
				"created_at":   "2024-01-31T10:00:00Z",
			},
			{
				"id":           "tax-doc-2",
				"name":         "1099 Form 2023",
				"type":         "1099",
				"year":         2023,
				"status":       "processing",
				"created_at":   "2024-01-31T10:00:00Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/eor/workers/worker-789/tax-documents", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetEORWorkerTaxDocuments(context.Background(), "worker-789")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "tax-doc-1", result[0].ID)
	assert.Equal(t, "W-2 Form 2023", result[0].Name)
	assert.Equal(t, "w2", result[0].Type)
	assert.Equal(t, 2023, result[0].Year)
	assert.Equal(t, "available", result[0].Status)
	assert.Equal(t, "https://api.deel.com/documents/tax-doc-1", result[0].DownloadURL)
}

func TestGetEORWorkerTaxDocuments_Unauthorized(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/eor/workers/worker-789/tax-documents", 403, map[string]string{"error": "access denied"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetEORWorkerTaxDocuments(context.Background(), "worker-789")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 403, apiErr.StatusCode)
}

func TestAddEORWorkerBankAccount(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/eor/workers/worker-789/bank-account", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Jane Smith", body["account_holder"])
		assert.Equal(t, "HSBC UK", body["bank_name"])
		assert.Equal(t, "12345678", body["account_number"])
		assert.Equal(t, "GB29NWBK60161331926819", body["iban"])
		assert.Equal(t, "HBUKGB4B", body["swift"])
		assert.Equal(t, "GBP", body["currency"])
		assert.Equal(t, true, body["is_primary"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":             "bank-001",
			"account_holder": "Jane Smith",
			"bank_name":      "HSBC UK",
			"account_number": "12345678",
			"iban":           "GB29NWBK60161331926819",
			"swift":          "HBUKGB4B",
			"currency":       "GBP",
			"is_primary":     true,
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.AddEORWorkerBankAccount(context.Background(), "worker-789", AddBankAccountParams{
		AccountHolder: "Jane Smith",
		BankName:      "HSBC UK",
		AccountNumber: "12345678",
		IBAN:          "GB29NWBK60161331926819",
		Swift:         "HBUKGB4B",
		Currency:      "GBP",
		IsPrimary:     true,
	})

	require.NoError(t, err)
	assert.Equal(t, "bank-001", result.ID)
	assert.Equal(t, "Jane Smith", result.AccountHolder)
	assert.Equal(t, "HSBC UK", result.BankName)
	assert.Equal(t, "GBP", result.Currency)
	assert.True(t, result.IsPrimary)
}

func TestAddEORWorkerBankAccount_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/workers/worker-789/bank-account", 400, map[string]string{"error": "invalid IBAN"})
	defer server.Close()

	client := testClient(server)
	_, err := client.AddEORWorkerBankAccount(context.Background(), "worker-789", AddBankAccountParams{
		AccountHolder: "Test User",
		BankName:      "Test Bank",
		AccountNumber: "123",
		IBAN:          "INVALID",
		Currency:      "USD",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestAddEORWorkerBankAccount_WorkerNotFound(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/eor/workers/invalid/bank-account", 404, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.AddEORWorkerBankAccount(context.Background(), "invalid", AddBankAccountParams{
		AccountHolder: "Test User",
		BankName:      "Test Bank",
		AccountNumber: "123456",
		Currency:      "USD",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}
