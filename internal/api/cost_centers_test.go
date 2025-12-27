package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCostCenters(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "cc1",
				"name":        "Engineering",
				"code":        "ENG",
				"description": "Engineering department",
				"status":      "active",
				"created_at":  "2024-01-01T00:00:00Z",
			},
			{
				"id":         "cc2",
				"name":       "Sales",
				"code":       "SALES",
				"status":     "active",
				"created_at": "2024-01-02T00:00:00Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/cost-centers", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListCostCenters(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "cc1", result[0].ID)
	assert.Equal(t, "Engineering", result[0].Name)
	assert.Equal(t, "ENG", result[0].Code)
	assert.Equal(t, "Engineering department", result[0].Description)
	assert.Equal(t, "active", result[0].Status)
	assert.Equal(t, "cc2", result[1].ID)
	assert.Equal(t, "Sales", result[1].Name)
}

func TestSyncCostCenters(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/cost-centers/sync", func(t *testing.T, body map[string]any) {
		costCenters := body["cost_centers"].([]any)
		assert.Len(t, costCenters, 2)

		cc1 := costCenters[0].(map[string]any)
		assert.Equal(t, "Engineering", cc1["name"])
		assert.Equal(t, "ENG", cc1["code"])
		assert.Equal(t, "Engineering department", cc1["description"])

		cc2 := costCenters[1].(map[string]any)
		assert.Equal(t, "Marketing", cc2["name"])
		assert.Equal(t, "MKT", cc2["code"])
	}, 200, map[string]any{
		"data": []map[string]any{
			{
				"id":          "cc-new1",
				"name":        "Engineering",
				"code":        "ENG",
				"description": "Engineering department",
				"status":      "active",
				"created_at":  "2024-01-15T00:00:00Z",
			},
			{
				"id":         "cc-new2",
				"name":       "Marketing",
				"code":       "MKT",
				"status":     "active",
				"created_at": "2024-01-15T00:00:00Z",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.SyncCostCenters(context.Background(), SyncCostCentersParams{
		CostCenters: []CostCenterInput{
			{
				Name:        "Engineering",
				Code:        "ENG",
				Description: "Engineering department",
			},
			{
				Name: "Marketing",
				Code: "MKT",
			},
		},
	})

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "cc-new1", result[0].ID)
	assert.Equal(t, "Engineering", result[0].Name)
	assert.Equal(t, "ENG", result[0].Code)
	assert.Equal(t, "Engineering department", result[0].Description)
	assert.Equal(t, "cc-new2", result[1].ID)
	assert.Equal(t, "Marketing", result[1].Name)
}
