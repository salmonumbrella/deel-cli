package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddGPBankAccount(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/bank-accounts", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "w-123", body["worker_id"])
		assert.Equal(t, "Jane Smith", body["account_holder"])
		assert.Equal(t, "HSBC", body["bank_name"])
		assert.Equal(t, "12345678", body["account_number"])
		assert.Equal(t, "GB29NWBK60161331926819", body["iban"])
		assert.Equal(t, "NWBKGB2L", body["swift"])
		assert.Equal(t, "GBP", body["currency"])
		assert.Equal(t, true, body["is_primary"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":             "ba-456",
			"worker_id":      "w-123",
			"account_holder": "Jane Smith",
			"bank_name":      "HSBC",
			"account_number": "12345678",
			"iban":           "GB29NWBK60161331926819",
			"swift":          "NWBKGB2L",
			"currency":       "GBP",
			"is_primary":     true,
			"status":         "active",
			"created_at":     "2024-02-20T09:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.AddGPBankAccount(context.Background(), AddGPBankAccountParams{
		WorkerID:      "w-123",
		AccountHolder: "Jane Smith",
		BankName:      "HSBC",
		AccountNumber: "12345678",
		IBAN:          "GB29NWBK60161331926819",
		Swift:         "NWBKGB2L",
		Currency:      "GBP",
		IsPrimary:     true,
	})

	require.NoError(t, err)
	assert.Equal(t, "ba-456", result.ID)
	assert.Equal(t, "w-123", result.WorkerID)
	assert.Equal(t, "Jane Smith", result.AccountHolder)
	assert.Equal(t, "HSBC", result.BankName)
	assert.Equal(t, "12345678", result.AccountNumber)
	assert.Equal(t, "GB29NWBK60161331926819", result.IBAN)
	assert.Equal(t, "NWBKGB2L", result.Swift)
	assert.Equal(t, "GBP", result.Currency)
	assert.Equal(t, true, result.IsPrimary)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "2024-02-20T09:00:00Z", result.CreatedAt)
}

func TestAddGPBankAccount_USAccount(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/bank-accounts", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "w-456", body["worker_id"])
		assert.Equal(t, "John Doe", body["account_holder"])
		assert.Equal(t, "Chase", body["bank_name"])
		assert.Equal(t, "987654321", body["account_number"])
		assert.Equal(t, "021000021", body["routing_number"])
		assert.Equal(t, "USD", body["currency"])
		assert.Equal(t, false, body["is_primary"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":             "ba-789",
			"worker_id":      "w-456",
			"account_holder": "John Doe",
			"bank_name":      "Chase",
			"account_number": "987654321",
			"routing_number": "021000021",
			"currency":       "USD",
			"is_primary":     false,
			"status":         "active",
			"created_at":     "2024-02-21T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.AddGPBankAccount(context.Background(), AddGPBankAccountParams{
		WorkerID:      "w-456",
		AccountHolder: "John Doe",
		BankName:      "Chase",
		AccountNumber: "987654321",
		RoutingNumber: "021000021",
		Currency:      "USD",
		IsPrimary:     false,
	})

	require.NoError(t, err)
	assert.Equal(t, "ba-789", result.ID)
	assert.Equal(t, "w-456", result.WorkerID)
	assert.Equal(t, "John Doe", result.AccountHolder)
	assert.Equal(t, "Chase", result.BankName)
	assert.Equal(t, "987654321", result.AccountNumber)
	assert.Equal(t, "021000021", result.RoutingNumber)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, false, result.IsPrimary)
}

func TestAddGPBankAccount_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/gp/bank-accounts", 400, map[string]string{"error": "invalid IBAN format"})
	defer server.Close()

	client := testClient(server)
	_, err := client.AddGPBankAccount(context.Background(), AddGPBankAccountParams{
		WorkerID:      "w-123",
		AccountHolder: "Jane Smith",
		BankName:      "HSBC",
		AccountNumber: "12345678",
		IBAN:          "INVALID_IBAN",
		Currency:      "GBP",
		IsPrimary:     true,
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestListGPBankAccounts(t *testing.T) {
	server := mockServerWithQuery(t, "GET", "/rest/v2/gp/bank-accounts", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "w-123", query["worker_id"])
	}, 200, map[string]any{
		"data": []map[string]any{
			{
				"id":             "ba-456",
				"worker_id":      "w-123",
				"account_holder": "Jane Smith",
				"bank_name":      "HSBC",
				"account_number": "12345678",
				"iban":           "GB29NWBK60161331926819",
				"swift":          "NWBKGB2L",
				"currency":       "GBP",
				"is_primary":     true,
				"status":         "active",
				"created_at":     "2024-02-20T09:00:00Z",
			},
			{
				"id":             "ba-789",
				"worker_id":      "w-123",
				"account_holder": "Jane Smith",
				"bank_name":      "Barclays",
				"account_number": "87654321",
				"iban":           "GB29BARC20201530093459",
				"swift":          "BARCGB22",
				"currency":       "GBP",
				"is_primary":     false,
				"status":         "active",
				"created_at":     "2024-02-21T10:00:00Z",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.ListGPBankAccounts(context.Background(), "w-123")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "ba-456", result[0].ID)
	assert.Equal(t, "w-123", result[0].WorkerID)
	assert.Equal(t, "HSBC", result[0].BankName)
	assert.Equal(t, true, result[0].IsPrimary)
	assert.Equal(t, "ba-789", result[1].ID)
	assert.Equal(t, "Barclays", result[1].BankName)
	assert.Equal(t, false, result[1].IsPrimary)
}

func TestListGPBankAccounts_Empty(t *testing.T) {
	server := mockServerWithQuery(t, "GET", "/rest/v2/gp/bank-accounts", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "w-999", query["worker_id"])
	}, 200, map[string]any{
		"data": []map[string]any{},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.ListGPBankAccounts(context.Background(), "w-999")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListGPBankAccounts_WorkerNotFound(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/gp/bank-accounts", 404, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.ListGPBankAccounts(context.Background(), "invalid")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestUpdateGPBankAccount(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/gp/bank-accounts/ba-456", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Jane Smith-Jones", body["account_holder"])
		assert.Equal(t, "HSBC UK", body["bank_name"])
		assert.Equal(t, "99887766", body["account_number"])
		assert.Equal(t, true, body["is_primary"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":             "ba-456",
			"worker_id":      "w-123",
			"account_holder": "Jane Smith-Jones",
			"bank_name":      "HSBC UK",
			"account_number": "99887766",
			"iban":           "GB29NWBK60161331926819",
			"swift":          "NWBKGB2L",
			"currency":       "GBP",
			"is_primary":     true,
			"status":         "active",
			"created_at":     "2024-02-20T09:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGPBankAccount(context.Background(), "ba-456", UpdateGPBankAccountParams{
		AccountHolder: "Jane Smith-Jones",
		BankName:      "HSBC UK",
		AccountNumber: "99887766",
		IsPrimary:     true,
	})

	require.NoError(t, err)
	assert.Equal(t, "ba-456", result.ID)
	assert.Equal(t, "Jane Smith-Jones", result.AccountHolder)
	assert.Equal(t, "HSBC UK", result.BankName)
	assert.Equal(t, "99887766", result.AccountNumber)
	assert.Equal(t, true, result.IsPrimary)
}

func TestUpdateGPBankAccount_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/gp/bank-accounts/invalid", 404, map[string]string{"error": "bank account not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateGPBankAccount(context.Background(), "invalid", UpdateGPBankAccountParams{
		BankName: "New Bank",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestGetGPBankGuide(t *testing.T) {
	server := mockServerWithQuery(t, "GET", "/rest/v2/gp/bank-guide", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "GB", query["country"])
	}, 200, map[string]any{
		"data": map[string]any{
			"country": "GB",
			"required_fields": []string{
				"account_holder",
				"bank_name",
				"account_number",
				"iban",
				"swift",
			},
			"optional_fields": []string{
				"routing_number",
			},
			"supported_banks": []string{
				"HSBC",
				"Barclays",
				"Lloyds",
				"NatWest",
			},
			"validation_rules": map[string]string{
				"iban":           "^GB[0-9]{2}[A-Z]{4}[0-9]{14}$",
				"swift":          "^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$",
				"account_number": "^[0-9]{8}$",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.GetGPBankGuide(context.Background(), "GB")

	require.NoError(t, err)
	assert.Equal(t, "GB", result.Country)
	assert.Equal(t, []string{
		"account_holder",
		"bank_name",
		"account_number",
		"iban",
		"swift",
	}, result.RequiredFields)
	assert.Equal(t, []string{"routing_number"}, result.OptionalFields)
	assert.Equal(t, []string{
		"HSBC",
		"Barclays",
		"Lloyds",
		"NatWest",
	}, result.SupportedBanks)
	require.NotNil(t, result.ValidationRules)
	assert.Equal(t, "^GB[0-9]{2}[A-Z]{4}[0-9]{14}$", result.ValidationRules["iban"])
	assert.Equal(t, "^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$", result.ValidationRules["swift"])
	assert.Equal(t, "^[0-9]{8}$", result.ValidationRules["account_number"])
}

func TestGetGPBankGuide_USGuide(t *testing.T) {
	server := mockServerWithQuery(t, "GET", "/rest/v2/gp/bank-guide", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "US", query["country"])
	}, 200, map[string]any{
		"data": map[string]any{
			"country": "US",
			"required_fields": []string{
				"account_holder",
				"bank_name",
				"account_number",
				"routing_number",
			},
			"optional_fields": []string{
				"iban",
				"swift",
			},
			"validation_rules": map[string]string{
				"routing_number": "^[0-9]{9}$",
				"account_number": "^[0-9]{4,17}$",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.GetGPBankGuide(context.Background(), "US")

	require.NoError(t, err)
	assert.Equal(t, "US", result.Country)
	assert.Equal(t, []string{
		"account_holder",
		"bank_name",
		"account_number",
		"routing_number",
	}, result.RequiredFields)
	assert.Equal(t, []string{"iban", "swift"}, result.OptionalFields)
	assert.Empty(t, result.SupportedBanks)
	require.NotNil(t, result.ValidationRules)
	assert.Equal(t, "^[0-9]{9}$", result.ValidationRules["routing_number"])
	assert.Equal(t, "^[0-9]{4,17}$", result.ValidationRules["account_number"])
}

func TestGetGPBankGuide_InvalidCountry(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/gp/bank-guide", 400, map[string]string{"error": "invalid country code"})
	defer server.Close()

	client := testClient(server)
	_, err := client.GetGPBankGuide(context.Background(), "INVALID")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}
