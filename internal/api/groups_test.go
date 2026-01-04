package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGroups(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":           "grp1",
				"name":         "Engineering",
				"description":  "Engineering team",
				"member_count": 15,
				"created_at":   "2024-01-01T00:00:00Z",
			},
			{
				"id":           "grp2",
				"name":         "Sales",
				"member_count": 8,
				"created_at":   "2024-01-02T00:00:00Z",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/groups", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListGroups(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "grp1", result[0].ID)
	assert.Equal(t, "Engineering", result[0].Name)
	assert.Equal(t, 15, result[0].MemberCount)
}

func TestGetGroup(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":           "grp1",
			"name":         "Engineering",
			"description":  "Engineering team",
			"member_count": 15,
			"created_at":   "2024-01-01T00:00:00Z",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/groups/grp1", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetGroup(context.Background(), "grp1")

	require.NoError(t, err)
	assert.Equal(t, "grp1", result.ID)
	assert.Equal(t, "Engineering", result.Name)
	assert.Equal(t, "Engineering team", result.Description)
	assert.Equal(t, 15, result.MemberCount)
	assert.Equal(t, "2024-01-01T00:00:00Z", result.CreatedAt)
}

func TestCreateGroup(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/groups", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Marketing", body["name"])
		assert.Equal(t, "Marketing department", body["description"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":           "grp-new",
			"name":         "Marketing",
			"description":  "Marketing department",
			"member_count": 0,
			"created_at":   "2024-01-15T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateGroup(context.Background(), CreateGroupParams{
		Name:        "Marketing",
		Description: "Marketing department",
	})

	require.NoError(t, err)
	assert.Equal(t, "grp-new", result.ID)
	assert.Equal(t, "Marketing", result.Name)
	assert.Equal(t, "Marketing department", result.Description)
	assert.Equal(t, 0, result.MemberCount)
}

func TestUpdateGroup(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/groups/grp1", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Engineering Team", body["name"])
		assert.Equal(t, "Updated description", body["description"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":           "grp1",
			"name":         "Engineering Team",
			"description":  "Updated description",
			"member_count": 15,
			"created_at":   "2024-01-01T00:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGroup(context.Background(), "grp1", UpdateGroupParams{
		Name:        "Engineering Team",
		Description: "Updated description",
	})

	require.NoError(t, err)
	assert.Equal(t, "grp1", result.ID)
	assert.Equal(t, "Engineering Team", result.Name)
	assert.Equal(t, "Updated description", result.Description)
}

func TestDeleteGroup(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/groups/grp1", http.StatusNoContent, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteGroup(context.Background(), "grp1")

	require.NoError(t, err)
}

func TestCloneGroup(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":           "grp-clone",
			"name":         "Engineering (Copy)",
			"description":  "Engineering team",
			"member_count": 0,
			"created_at":   "2024-01-15T00:00:00Z",
		},
	}
	server := mockServer(t, "POST", "/rest/v2/groups/grp1/clone", http.StatusCreated, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.CloneGroup(context.Background(), "grp1")

	require.NoError(t, err)
	assert.Equal(t, "grp-clone", result.ID)
	assert.Equal(t, "Engineering (Copy)", result.Name)
	assert.Equal(t, 0, result.MemberCount)
}
