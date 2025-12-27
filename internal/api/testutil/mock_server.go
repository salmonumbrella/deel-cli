// Package testutil provides testing utilities for API tests.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockServer is a configurable test HTTP server for API testing.
type MockServer struct {
	server   *httptest.Server
	mu       sync.RWMutex
	handlers map[string]http.HandlerFunc
}

// NewMockServer creates a new MockServer.
func NewMockServer() *MockServer {
	m := &MockServer{
		handlers: make(map[string]http.HandlerFunc),
	}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		key := r.Method + " " + r.URL.Path
		handler, ok := m.handlers[key]
		m.mu.RUnlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		handler(w, r)
	}))

	return m
}

// Close shuts down the mock server.
func (m *MockServer) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// URL returns the base URL of the mock server.
func (m *MockServer) URL() string {
	return m.server.URL
}

// Handle registers a custom handler for a method and path.
func (m *MockServer) Handle(method, path string, handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[method+" "+path] = handler
}

// HandleJSON registers a handler that returns JSON response.
func (m *MockServer) HandleJSON(method, path string, statusCode int, response any) {
	m.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			json.NewEncoder(w).Encode(response)
		}
	})
}

// HandleError registers a handler that returns a JSON error response.
func (m *MockServer) HandleError(method, path string, statusCode int, message string) {
	m.HandleJSON(method, path, statusCode, map[string]string{"error": message})
}
