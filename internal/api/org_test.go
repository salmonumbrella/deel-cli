package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLegalEntity(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/legal-entities", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Acme Corp", body["name"])
		assert.Equal(t, "US", body["country"])
		assert.Equal(t, "llc", body["type"])
		assert.Equal(t, "12-3456789", body["registration_number"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":                  "le-123",
			"name":                "Acme Corp",
			"country":             "US",
			"type":                "llc",
			"status":              "active",
			"registration_number": "12-3456789",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateLegalEntity(context.Background(), CreateLegalEntityParams{
		Name:               "Acme Corp",
		Country:            "US",
		Type:               "llc",
		RegistrationNumber: "12-3456789",
	})

	require.NoError(t, err)
	assert.Equal(t, "le-123", result.ID)
	assert.Equal(t, "Acme Corp", result.Name)
	assert.Equal(t, "US", result.Country)
	assert.Equal(t, "llc", result.Type)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "12-3456789", result.RegistrationNumber)
}

func TestCreateLegalEntity_WithoutRegistrationNumber(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/legal-entities", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Global Inc", body["name"])
		assert.Equal(t, "UK", body["country"])
		assert.Equal(t, "limited", body["type"])
		_, hasRegNum := body["registration_number"]
		assert.False(t, hasRegNum)
	}, 201, map[string]any{
		"data": map[string]any{
			"id":      "le-456",
			"name":    "Global Inc",
			"country": "UK",
			"type":    "limited",
			"status":  "pending",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateLegalEntity(context.Background(), CreateLegalEntityParams{
		Name:    "Global Inc",
		Country: "UK",
		Type:    "limited",
	})

	require.NoError(t, err)
	assert.Equal(t, "le-456", result.ID)
	assert.Equal(t, "Global Inc", result.Name)
	assert.Equal(t, "pending", result.Status)
}

func TestCreateLegalEntity_Error(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/legal-entities", 400, map[string]any{
		"error": "Invalid country code",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateLegalEntity(context.Background(), CreateLegalEntityParams{
		Name:    "Test Corp",
		Country: "INVALID",
		Type:    "llc",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestUpdateLegalEntity(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/legal-entities/le-123", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Updated Corp", body["name"])
		assert.Equal(t, "98-7654321", body["registration_number"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":                  "le-123",
			"name":                "Updated Corp",
			"country":             "US",
			"type":                "llc",
			"status":              "active",
			"registration_number": "98-7654321",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateLegalEntity(context.Background(), "le-123", UpdateLegalEntityParams{
		Name:               "Updated Corp",
		RegistrationNumber: "98-7654321",
	})

	require.NoError(t, err)
	assert.Equal(t, "le-123", result.ID)
	assert.Equal(t, "Updated Corp", result.Name)
	assert.Equal(t, "98-7654321", result.RegistrationNumber)
}

func TestUpdateLegalEntity_PartialUpdate(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/legal-entities/le-789", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "corporation", body["type"])
		// Only type should be in the request
		_, hasName := body["name"]
		assert.False(t, hasName)
	}, 200, map[string]any{
		"data": map[string]any{
			"id":      "le-789",
			"name":    "Unchanged Name",
			"country": "CA",
			"type":    "corporation",
			"status":  "active",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateLegalEntity(context.Background(), "le-789", UpdateLegalEntityParams{
		Type: "corporation",
	})

	require.NoError(t, err)
	assert.Equal(t, "le-789", result.ID)
	assert.Equal(t, "corporation", result.Type)
}

func TestUpdateLegalEntity_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/legal-entities/le-999", 404, map[string]any{
		"error": "Legal entity not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateLegalEntity(context.Background(), "le-999", UpdateLegalEntityParams{
		Name: "Test",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestDeleteLegalEntity(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/legal-entities/le-123", 204, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteLegalEntity(context.Background(), "le-123")

	require.NoError(t, err)
}

func TestDeleteLegalEntity_NotFound(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/legal-entities/le-999", 404, map[string]any{
		"error": "Legal entity not found",
	})
	defer server.Close()

	client := testClient(server)
	err := client.DeleteLegalEntity(context.Background(), "le-999")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestDeleteLegalEntity_Conflict(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/legal-entities/le-123", 409, map[string]any{
		"error": "Cannot delete legal entity with active contracts",
	})
	defer server.Close()

	client := testClient(server)
	err := client.DeleteLegalEntity(context.Background(), "le-123")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 409, apiErr.StatusCode)
}

func TestGetPayrollSettings(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":                 "ps-123",
			"legal_entity_id":    "le-123",
			"payroll_frequency":  "monthly",
			"payment_method":     "ach",
			"currency":           "USD",
			"tax_id":             "12-3456789",
			"bank_account":       "****1234",
			"payroll_provider":   "deel",
			"auto_approval":      true,
			"notification_email": "payroll@example.com",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/legal-entities/le-123/payroll-settings", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetPayrollSettings(context.Background(), "le-123")

	require.NoError(t, err)
	assert.Equal(t, "ps-123", result.ID)
	assert.Equal(t, "le-123", result.LegalEntityID)
	assert.Equal(t, "monthly", result.PayrollFrequency)
	assert.Equal(t, "ach", result.PaymentMethod)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "12-3456789", result.TaxID)
	assert.Equal(t, "****1234", result.BankAccount)
	assert.Equal(t, "deel", result.PayrollProvider)
	assert.True(t, result.AutoApproval)
	assert.Equal(t, "payroll@example.com", result.NotificationEmail)
}

func TestGetPayrollSettings_MinimalData(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":                "ps-456",
			"legal_entity_id":   "le-456",
			"payroll_frequency": "biweekly",
			"payment_method":    "wire",
			"currency":          "EUR",
			"auto_approval":     false,
		},
	}
	server := mockServer(t, "GET", "/rest/v2/legal-entities/le-456/payroll-settings", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetPayrollSettings(context.Background(), "le-456")

	require.NoError(t, err)
	assert.Equal(t, "ps-456", result.ID)
	assert.Equal(t, "le-456", result.LegalEntityID)
	assert.Equal(t, "biweekly", result.PayrollFrequency)
	assert.Equal(t, "wire", result.PaymentMethod)
	assert.Equal(t, "EUR", result.Currency)
	assert.False(t, result.AutoApproval)
	assert.Empty(t, result.TaxID)
	assert.Empty(t, result.BankAccount)
	assert.Empty(t, result.PayrollProvider)
	assert.Empty(t, result.NotificationEmail)
}

func TestGetPayrollSettings_NotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/legal-entities/le-999/payroll-settings", 404, map[string]any{
		"error": "Payroll settings not found",
	})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetPayrollSettings(context.Background(), "le-999")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestListLegalEntities(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"id": "le-1", "name": "Delicious Milk Corporation", "country": "CA"},
			{"id": "le-2", "name": "Wanver Inc", "country": "US"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/legal-entities", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListLegalEntities(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "le-1", result[0].ID)
	assert.Equal(t, "Delicious Milk Corporation", result[0].Name)
}
