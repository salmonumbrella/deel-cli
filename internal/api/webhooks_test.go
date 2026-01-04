package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWebhooks(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"id": "wh1", "url": "https://example.com/hook", "events": []string{"contract.created"}},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/webhooks", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListWebhooks(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "wh1", result[0].ID)
}

func TestGetWebhook(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":     "wh1",
			"url":    "https://example.com/hook",
			"events": []string{"contract.created"},
			"status": "active",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/webhooks/wh1", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetWebhook(context.Background(), "wh1")

	require.NoError(t, err)
	assert.Equal(t, "wh1", result.ID)
	assert.Equal(t, "active", result.Status)
}

func TestCreateWebhook(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/webhooks", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "https://example.com/hook", body["url"])
		events := body["events"].([]any)
		assert.Contains(t, events, "contract.created")
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":     "wh-new",
			"url":    "https://example.com/hook",
			"events": []string{"contract.created"},
			"secret": "whsec_xxx",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateWebhook(context.Background(), CreateWebhookParams{
		URL:    "https://example.com/hook",
		Events: []string{"contract.created"},
	})

	require.NoError(t, err)
	assert.Equal(t, "wh-new", result.ID)
	assert.Equal(t, "whsec_xxx", result.Secret)
}

func TestUpdateWebhook(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/webhooks/wh1", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "https://new-url.com/hook", body["url"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":  "wh1",
			"url": "https://new-url.com/hook",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateWebhook(context.Background(), "wh1", UpdateWebhookParams{
		URL: "https://new-url.com/hook",
	})

	require.NoError(t, err)
	assert.Equal(t, "https://new-url.com/hook", result.URL)
}

func TestDeleteWebhook(t *testing.T) {
	server := mockServer(t, "DELETE", "/rest/v2/webhooks/wh1", http.StatusNoContent, nil)
	defer server.Close()

	client := testClient(server)
	err := client.DeleteWebhook(context.Background(), "wh1")

	require.NoError(t, err)
}

func TestListWebhookEventTypes(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{"name": "contract.created", "description": "When a contract is created"},
			{"name": "contract.signed", "description": "When a contract is signed"},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/webhooks/event-types", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListWebhookEventTypes(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "contract.created", result[0].Name)
}
