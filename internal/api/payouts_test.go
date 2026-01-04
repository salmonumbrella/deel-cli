package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawFunds(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/payouts/withdraw", func(t *testing.T, body map[string]any) {
		assert.Equal(t, 1000.0, body["amount"])
		assert.Equal(t, "USD", body["currency"])
		assert.Equal(t, "Monthly payout", body["description"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":          "wd-123",
			"amount":      1000.0,
			"currency":    "USD",
			"status":      "pending",
			"description": "Monthly payout",
			"created_at":  "2024-01-15T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.WithdrawFunds(context.Background(), WithdrawFundsParams{
		Amount:      1000.0,
		Currency:    "USD",
		Description: "Monthly payout",
	})

	require.NoError(t, err)
	assert.Equal(t, "wd-123", result.ID)
	assert.Equal(t, 1000.0, result.Amount)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, "Monthly payout", result.Description)
}

func TestWithdrawFunds_MinimalParams(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/payouts/withdraw", func(t *testing.T, body map[string]any) {
		assert.Equal(t, 500.0, body["amount"])
		assert.Equal(t, "EUR", body["currency"])
		assert.Nil(t, body["description"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":         "wd-456",
			"amount":     500.0,
			"currency":   "EUR",
			"status":     "processing",
			"created_at": "2024-01-15T11:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.WithdrawFunds(context.Background(), WithdrawFundsParams{
		Amount:   500.0,
		Currency: "EUR",
	})

	require.NoError(t, err)
	assert.Equal(t, "wd-456", result.ID)
	assert.Equal(t, "processing", result.Status)
}

func TestGetAutoWithdrawal(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"enabled":   true,
			"threshold": 5000.0,
			"currency":  "USD",
			"schedule":  "weekly",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/payouts/auto-withdrawal", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetAutoWithdrawal(context.Background())

	require.NoError(t, err)
	assert.True(t, result.Enabled)
	assert.Equal(t, 5000.0, result.Threshold)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "weekly", result.Schedule)
}

func TestGetAutoWithdrawal_Disabled(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"enabled": false,
		},
	}
	server := mockServer(t, "GET", "/rest/v2/payouts/auto-withdrawal", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetAutoWithdrawal(context.Background())

	require.NoError(t, err)
	assert.False(t, result.Enabled)
}

func TestSetAutoWithdrawal_Enable(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/payouts/auto-withdrawal", func(t *testing.T, body map[string]any) {
		assert.Equal(t, true, body["enabled"])
		assert.Equal(t, 10000.0, body["threshold"])
		assert.Equal(t, "USD", body["currency"])
		assert.Equal(t, "monthly", body["schedule"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"enabled":   true,
			"threshold": 10000.0,
			"currency":  "USD",
			"schedule":  "monthly",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.SetAutoWithdrawal(context.Background(), SetAutoWithdrawalParams{
		Enabled:   true,
		Threshold: 10000.0,
		Currency:  "USD",
		Schedule:  "monthly",
	})

	require.NoError(t, err)
	assert.True(t, result.Enabled)
	assert.Equal(t, 10000.0, result.Threshold)
	assert.Equal(t, "USD", result.Currency)
	assert.Equal(t, "monthly", result.Schedule)
}

func TestSetAutoWithdrawal_Disable(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/payouts/auto-withdrawal", func(t *testing.T, body map[string]any) {
		assert.Equal(t, false, body["enabled"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"enabled": false,
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.SetAutoWithdrawal(context.Background(), SetAutoWithdrawalParams{
		Enabled: false,
	})

	require.NoError(t, err)
	assert.False(t, result.Enabled)
}

func TestListContractorBalances(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"contractor_id":   "ctr-1",
				"contractor_name": "John Doe",
				"balance":         2500.0,
				"currency":        "USD",
				"pending_amount":  500.0,
				"updated_at":      "2024-01-15T10:00:00Z",
			},
			{
				"contractor_id":   "ctr-2",
				"contractor_name": "Jane Smith",
				"balance":         3200.0,
				"currency":        "EUR",
				"pending_amount":  0.0,
				"updated_at":      "2024-01-15T09:30:00Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/contractors/balances", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListContractorBalances(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, "ctr-1", result[0].ContractorID)
	assert.Equal(t, "John Doe", result[0].ContractorName)
	assert.Equal(t, 2500.0, result[0].Balance)
	assert.Equal(t, "USD", result[0].Currency)
	assert.Equal(t, 500.0, result[0].PendingAmount)

	assert.Equal(t, "ctr-2", result[1].ContractorID)
	assert.Equal(t, "Jane Smith", result[1].ContractorName)
	assert.Equal(t, 3200.0, result[1].Balance)
	assert.Equal(t, "EUR", result[1].Currency)
}

func TestListContractorBalances_Empty(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{},
	}
	server := mockServer(t, "GET", "/rest/v2/contractors/balances", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListContractorBalances(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 0)
}
