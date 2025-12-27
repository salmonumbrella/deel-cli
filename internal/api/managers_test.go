package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListManagers(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":         "mgr-1",
				"email":      "manager1@example.com",
				"first_name": "John",
				"last_name":  "Doe",
				"role":       "admin",
				"status":     "active",
				"team_ids":   []any{"team-1", "team-2"},
				"created_at": "2025-01-15T10:00:00Z",
			},
			{
				"id":         "mgr-2",
				"email":      "manager2@example.com",
				"first_name": "Jane",
				"last_name":  "Smith",
				"role":       "manager",
				"status":     "active",
				"created_at": "2025-01-16T10:00:00Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/managers", 200, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListManagers(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "mgr-1", result[0].ID)
	assert.Equal(t, "manager1@example.com", result[0].Email)
	assert.Equal(t, "John", result[0].FirstName)
	assert.Equal(t, "Doe", result[0].LastName)
	assert.Equal(t, "admin", result[0].Role)
	assert.Equal(t, "active", result[0].Status)
	assert.Len(t, result[0].TeamIDs, 2)
	assert.Equal(t, "2025-01-15T10:00:00Z", result[0].CreatedAt)
}

func TestCreateManager(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/managers", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "newmgr@example.com", body["email"])
		assert.Equal(t, "Alice", body["first_name"])
		assert.Equal(t, "Johnson", body["last_name"])
		assert.Equal(t, "manager", body["role"])
		teamIDs := body["team_ids"].([]any)
		assert.Len(t, teamIDs, 2)
		assert.Equal(t, "team-1", teamIDs[0])
		assert.Equal(t, "team-3", teamIDs[1])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":         "mgr-new",
			"email":      "newmgr@example.com",
			"first_name": "Alice",
			"last_name":  "Johnson",
			"role":       "manager",
			"status":     "pending",
			"team_ids":   []any{"team-1", "team-3"},
			"created_at": "2025-01-17T10:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateManager(context.Background(), CreateManagerParams{
		Email:     "newmgr@example.com",
		FirstName: "Alice",
		LastName:  "Johnson",
		Role:      "manager",
		TeamIDs:   []string{"team-1", "team-3"},
	})

	require.NoError(t, err)
	assert.Equal(t, "mgr-new", result.ID)
	assert.Equal(t, "newmgr@example.com", result.Email)
	assert.Equal(t, "Alice", result.FirstName)
	assert.Equal(t, "Johnson", result.LastName)
	assert.Equal(t, "manager", result.Role)
	assert.Equal(t, "pending", result.Status)
	assert.Len(t, result.TeamIDs, 2)
}

func TestCreateManagerWithoutTeams(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/managers", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "minimal@example.com", body["email"])
		assert.Equal(t, "Bob", body["first_name"])
		assert.Equal(t, "Williams", body["last_name"])
		assert.Equal(t, "viewer", body["role"])
		// team_ids should be omitted or empty
		_, hasTeamIDs := body["team_ids"]
		if hasTeamIDs {
			assert.Empty(t, body["team_ids"])
		}
	}, 201, map[string]any{
		"data": map[string]any{
			"id":         "mgr-min",
			"email":      "minimal@example.com",
			"first_name": "Bob",
			"last_name":  "Williams",
			"role":       "viewer",
			"status":     "pending",
			"created_at": "2025-01-17T11:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateManager(context.Background(), CreateManagerParams{
		Email:     "minimal@example.com",
		FirstName: "Bob",
		LastName:  "Williams",
		Role:      "viewer",
	})

	require.NoError(t, err)
	assert.Equal(t, "mgr-min", result.ID)
	assert.Equal(t, "minimal@example.com", result.Email)
}

func TestCreateManagerMagicLink(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/managers/magic-link", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "magiclink@example.com", body["email"])
	}, 201, map[string]any{
		"data": map[string]any{
			"manager_id": "mgr-123",
			"link":       "https://app.deel.com/magic/abc123xyz",
			"expires_at": "2025-01-17T12:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateManagerMagicLink(context.Background(), CreateMagicLinkParams{
		Email: "magiclink@example.com",
	})

	require.NoError(t, err)
	assert.Equal(t, "mgr-123", result.ManagerID)
	assert.Equal(t, "https://app.deel.com/magic/abc123xyz", result.Link)
	assert.Equal(t, "2025-01-17T12:00:00Z", result.ExpiresAt)
	assert.Contains(t, result.Link, "magic")
}
