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
			{"id": "t1", "title": "Task 1", "status": "pending", "amount": "100.00"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/contracts/c1/tasks", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListTasks(context.Background(), TasksListParams{ContractID: "c1"})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, 100.0, result.Data[0].Amount)
}

func TestCreateTask(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/contracts/c1/tasks", func(t *testing.T, body map[string]any) {
		data := body["data"].(map[string]any)
		assert.Equal(t, "New Task", data["title"])
		assert.Equal(t, "50.00", data["amount"])
		assert.NotEmpty(t, data["date_submitted"])
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
	server := mockServerWithBody(t, "POST", "/rest/v2/contracts/c1/tasks/t1/reviews", func(t *testing.T, body map[string]any) {
		data := body["data"].(map[string]any)
		assert.Equal(t, "approved", data["status"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{"id": "t1", "status": "approved"},
	})
	defer server.Close()

	client := testClient(server)
	err := client.ReviewTask(context.Background(), "c1", "t1", "approved")

	require.NoError(t, err)
}
