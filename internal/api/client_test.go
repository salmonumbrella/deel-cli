package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockServer creates a test server that returns the given response
func mockServer(t *testing.T, method, path string, statusCode int, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, method, r.Method)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}
	}))
}

// mockServerWithBody creates a test server that validates request body
func mockServerWithBody(t *testing.T, method, path string, validateBody func(t *testing.T, body map[string]any), statusCode int, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, method, r.Method)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		if validateBody != nil {
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			validateBody(t, body)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}
	}))
}

// mockServerWithMultipart creates a test server that validates multipart/form-data bodies
func mockServerWithMultipart(t *testing.T, method, path string, validateForm func(t *testing.T, fields map[string]string), statusCode int, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, method, r.Method)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		contentType := r.Header.Get("Content-Type")
		assert.Contains(t, contentType, "multipart/form-data")

		if validateForm != nil {
			err := r.ParseMultipartForm(10 << 20) // 10MB max
			require.NoError(t, err)

			fields := make(map[string]string)
			for key, values := range r.MultipartForm.Value {
				if len(values) > 0 {
					fields[key] = values[0]
				}
			}
			validateForm(t, fields)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}
	}))
}

// mockServerWithQuery creates a test server that validates query parameters.
func mockServerWithQuery(t *testing.T, path string, validateQuery func(t *testing.T, query map[string]string), response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		if validateQuery != nil {
			query := make(map[string]string)
			for key, values := range r.URL.Query() {
				if len(values) > 0 {
					query[key] = values[0]
				}
			}
			validateQuery(t, query)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}
	}))
}

// testClient creates a client pointing to the test server
func testClient(server *httptest.Server) *Client {
	c := NewClient("test-token")
	c.baseURL = server.URL
	return c
}

func TestClient_Get_Success(t *testing.T) {
	expected := map[string]any{"data": "test"}
	server := mockServer(t, "GET", "/test", http.StatusOK, expected)
	defer server.Close()

	client := testClient(server)
	resp, err := client.Get(context.Background(), "/test")

	require.NoError(t, err)
	assert.Contains(t, string(resp), "test")
}

func TestClient_Get_APIError(t *testing.T) {
	server := mockServer(t, "GET", "/test", http.StatusNotFound, map[string]string{"error": "not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.Get(context.Background(), "/test")

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestClient_Post_Success(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/test", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "value", body["key"])
	}, http.StatusCreated, map[string]any{"id": "123"})
	defer server.Close()

	client := testClient(server)
	resp, err := client.Post(context.Background(), "/test", map[string]string{"key": "value"})

	require.NoError(t, err)
	assert.Contains(t, string(resp), "123")
}
