package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWorkerRelations(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/profiles/profile-123/worker-relations", 200, map[string]any{
		"data": []map[string]any{
			{
				"id":            "relation-001",
				"profile_id":    "profile-123",
				"manager_id":    "manager-456",
				"relation_type": "direct_report",
				"start_date":    "2024-01-01",
				"end_date":      "",
				"status":        "active",
				"created_at":    "2024-01-01T10:00:00Z",
			},
			{
				"id":            "relation-002",
				"profile_id":    "profile-123",
				"manager_id":    "manager-789",
				"relation_type": "dotted_line",
				"start_date":    "2024-02-01",
				"end_date":      "2024-06-30",
				"status":        "inactive",
				"created_at":    "2024-02-01T10:00:00Z",
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.ListWorkerRelations(context.Background(), "profile-123")

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "relation-001", result[0].ID)
	assert.Equal(t, "profile-123", result[0].ProfileID)
	assert.Equal(t, "manager-456", result[0].ManagerID)
	assert.Equal(t, "direct_report", result[0].RelationType)
	assert.Equal(t, "2024-01-01", result[0].StartDate)
	assert.Equal(t, "", result[0].EndDate)
	assert.Equal(t, "active", result[0].Status)
	assert.Equal(t, "2024-01-01T10:00:00Z", result[0].CreatedAt)

	assert.Equal(t, "relation-002", result[1].ID)
	assert.Equal(t, "dotted_line", result[1].RelationType)
	assert.Equal(t, "2024-06-30", result[1].EndDate)
	assert.Equal(t, "inactive", result[1].Status)
}

func TestCreateWorkerRelation(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/worker-relations", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "profile-123", body["profile_id"])
		assert.Equal(t, "manager-456", body["manager_id"])
		assert.Equal(t, "direct_report", body["relation_type"])
		assert.Equal(t, "2024-01-15", body["start_date"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":            "relation-new",
			"profile_id":    "profile-123",
			"manager_id":    "manager-456",
			"relation_type": "direct_report",
			"start_date":    "2024-01-15",
			"end_date":      "",
			"status":        "active",
			"created_at":    "2024-01-15T14:30:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateWorkerRelation(context.Background(), CreateWorkerRelationParams{
		ProfileID:    "profile-123",
		ManagerID:    "manager-456",
		RelationType: "direct_report",
		StartDate:    "2024-01-15",
	})

	require.NoError(t, err)
	assert.Equal(t, "relation-new", result.ID)
	assert.Equal(t, "profile-123", result.ProfileID)
	assert.Equal(t, "manager-456", result.ManagerID)
	assert.Equal(t, "direct_report", result.RelationType)
	assert.Equal(t, "2024-01-15", result.StartDate)
	assert.Equal(t, "", result.EndDate)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "2024-01-15T14:30:00Z", result.CreatedAt)
}

func TestDeleteWorkerRelation(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/worker-relations/relation-123", 204, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteWorkerRelation(context.Background(), "relation-123")

	require.NoError(t, err)
}
