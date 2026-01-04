package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAdjustment(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/adjustments", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "c1", body["contract_id"])
		assert.Equal(t, "cat1", body["category_id"])
		assert.Equal(t, 1000.0, body["amount"])
		assert.Equal(t, "USD", body["currency"])
		assert.Equal(t, "Performance bonus", body["description"])
		assert.Equal(t, "2024-01-15", body["date"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":          "adj1",
			"contract_id": "c1",
			"category_id": "cat1",
			"amount":      1000.0,
			"currency":    "USD",
			"description": "Performance bonus",
			"date":        "2024-01-15",
			"status":      "pending",
			"created_at":  "2024-01-15T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateAdjustment(context.Background(), CreateAdjustmentParams{
		ContractID:  "c1",
		CategoryID:  "cat1",
		Amount:      1000.0,
		Currency:    "USD",
		Description: "Performance bonus",
		Date:        "2024-01-15",
	})

	require.NoError(t, err)
	assert.Equal(t, "adj1", result.ID)
	assert.Equal(t, "c1", result.ContractID)
	assert.Equal(t, 1000.0, result.Amount)
	assert.Equal(t, "pending", result.Status)
}

func TestGetAdjustment(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":          "adj1",
			"contract_id": "c1",
			"category_id": "cat1",
			"amount":      1000.0,
			"currency":    "USD",
			"description": "Performance bonus",
			"date":        "2024-01-15",
			"status":      "pending",
			"created_at":  "2024-01-15T10:00:00Z",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/adjustments/adj1", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetAdjustment(context.Background(), "adj1")

	require.NoError(t, err)
	assert.Equal(t, "adj1", result.ID)
	assert.Equal(t, "c1", result.ContractID)
	assert.Equal(t, 1000.0, result.Amount)
}

func TestUpdateAdjustment(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/adjustments/adj1", func(t *testing.T, body map[string]any) {
		assert.Equal(t, 1500.0, body["amount"])
		assert.Equal(t, "Updated bonus amount", body["description"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":          "adj1",
			"contract_id": "c1",
			"category_id": "cat1",
			"amount":      1500.0,
			"currency":    "USD",
			"description": "Updated bonus amount",
			"date":        "2024-01-15",
			"status":      "pending",
			"created_at":  "2024-01-15T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateAdjustment(context.Background(), "adj1", UpdateAdjustmentParams{
		Amount:      1500.0,
		Description: "Updated bonus amount",
	})

	require.NoError(t, err)
	assert.Equal(t, "adj1", result.ID)
	assert.Equal(t, 1500.0, result.Amount)
	assert.Equal(t, "Updated bonus amount", result.Description)
}

func TestDeleteAdjustment(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/adjustments/adj1", http.StatusNoContent, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteAdjustment(context.Background(), "adj1")

	require.NoError(t, err)
}

func TestListAdjustments(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "adj1",
				"contract_id": "c1",
				"category_id": "cat1",
				"amount":      1000.0,
				"currency":    "USD",
				"status":      "pending",
			},
			{
				"id":          "adj2",
				"contract_id": "c1",
				"category_id": "cat2",
				"amount":      -500.0,
				"currency":    "USD",
				"status":      "approved",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/adjustments", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListAdjustments(context.Background(), ListAdjustmentsParams{})

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "adj1", result[0].ID)
	assert.Equal(t, "adj2", result[1].ID)
}

func TestListAdjustments_WithFilters(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "adj1",
				"contract_id": "c1",
				"category_id": "cat1",
				"amount":      1000.0,
				"currency":    "USD",
				"status":      "pending",
			},
		},
	}
	server := mockServerWithQuery(t, "GET", "/rest/v2/adjustments", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "c1", query["contract_id"])
		assert.Equal(t, "cat1", query["category_id"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListAdjustments(context.Background(), ListAdjustmentsParams{
		ContractID: "c1",
		CategoryID: "cat1",
	})

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "adj1", result[0].ID)
}

func TestListAdjustmentCategories(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "cat1",
				"name":        "Performance Bonus",
				"description": "Bonuses for exceptional performance",
				"type":        "bonus",
			},
			{
				"id":   "cat2",
				"name": "Equipment Expense",
				"type": "expense",
			},
			{
				"id":          "cat3",
				"name":        "Tax Deduction",
				"description": "Tax withholding adjustments",
				"type":        "deduction",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/adjustments/categories", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListAdjustmentCategories(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "cat1", result[0].ID)
	assert.Equal(t, "Performance Bonus", result[0].Name)
	assert.Equal(t, "bonus", result[0].Type)
	assert.Equal(t, "cat2", result[1].ID)
	assert.Equal(t, "expense", result[1].Type)
	assert.Equal(t, "cat3", result[2].ID)
	assert.Equal(t, "deduction", result[2].Type)
}
