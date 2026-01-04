package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListMilestones(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"id": "m1", "title": "Phase 1", "amount": 5000, "status": "pending"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/contracts/c1/milestones", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListMilestones(context.Background(), "c1")

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "m1", result[0].ID)
}

func TestCreateMilestone(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/milestones", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "c1", body["contract_id"])
		assert.Equal(t, "New Milestone", body["title"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":    "m-new",
			"title": "New Milestone",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateMilestone(context.Background(), CreateMilestoneParams{
		ContractID:  "c1",
		Title:       "New Milestone",
		Description: "Do the work",
		Amount:      1000,
	})

	require.NoError(t, err)
	assert.Equal(t, "m-new", result.ID)
}

func TestDeleteMilestone(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/milestones/m1", http.StatusNoContent, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteMilestone(context.Background(), "m1")

	require.NoError(t, err)
}
