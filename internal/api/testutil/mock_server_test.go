package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockServer_HandleJSON(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	expected := map[string]any{"id": "123", "name": "test"}
	server.HandleJSON("GET", "/api/test", http.StatusOK, expected)

	resp, err := http.Get(server.URL() + "/api/test")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "test", result["name"])
}

func TestMockServer_HandleError(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.HandleError("POST", "/api/create", http.StatusBadRequest, "invalid input")

	resp, err := http.Post(server.URL()+"/api/create", "application/json", nil)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	assert.Equal(t, "invalid input", result["error"])
}

func TestMockServer_NotFound(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Request an unregistered path
	resp, err := http.Get(server.URL() + "/api/unregistered")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMockServer_Handle(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Custom handler
	server.Handle("PUT", "/api/custom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "header")
		w.WriteHeader(http.StatusAccepted)
		if _, err := w.Write([]byte("custom response")); err != nil {
			return
		}
	})

	req, err := http.NewRequest("PUT", server.URL()+"/api/custom", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "header", resp.Header.Get("X-Custom"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "custom response", string(body))
}

func TestMockServer_MultipleRoutes(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	server.HandleJSON("GET", "/api/one", http.StatusOK, map[string]string{"route": "one"})
	server.HandleJSON("GET", "/api/two", http.StatusOK, map[string]string{"route": "two"})
	server.HandleJSON("POST", "/api/one", http.StatusCreated, map[string]string{"action": "created"})

	// Test route one with GET
	resp, err := http.Get(server.URL() + "/api/one")
	require.NoError(t, err)
	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, "one", result["route"])

	// Test route two with GET
	resp, err = http.Get(server.URL() + "/api/two")
	require.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, "two", result["route"])

	// Test route one with POST
	resp, err = http.Post(server.URL()+"/api/one", "application/json", nil)
	require.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "created", result["action"])
}
