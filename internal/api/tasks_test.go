package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTasks(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"id": "t1", "title": "Task 1", "status": "pending", "amount": 100},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/tasks", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListTasks(context.Background(), TasksListParams{})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
}

func TestCreateTask(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/tasks", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "c1", body["contract_id"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{"id": "t-new"},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateTask(context.Background(), CreateTaskParams{
		ContractID: "c1",
		Title:      "New Task",
		Amount:     50,
	})

	require.NoError(t, err)
	assert.Equal(t, "t-new", result.ID)
}

func TestReviewTask(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/tasks/t1/review", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "approved", body["status"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{"id": "t1", "status": "approved"},
	})
	defer server.Close()

	client := testClient(server)
	err := client.ReviewTask(context.Background(), "t1", "approved")

	require.NoError(t, err)
}
