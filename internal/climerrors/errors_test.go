// internal/climerrors/errors_test.go
package climerrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIError_Error(t *testing.T) {
	err := &CLIError{
		Operation:   "listing people",
		Err:         errors.New("not found"),
		Suggestions: []string{"check the ID"},
	}
	assert.Equal(t, "failed listing people: not found", err.Error())
}

func TestCategorize_APIError(t *testing.T) {
	tests := []struct {
		status   int
		expected Category
	}{
		{401, CategoryAuth},
		{403, CategoryForbidden},
		{404, CategoryNotFound},
		{400, CategoryValidation},
		{422, CategoryValidation},
		{429, CategoryRateLimit},
		{500, CategoryServer},
		{503, CategoryServer},
	}

	for _, tt := range tests {
		apiErr := &MockAPIError{StatusCode: tt.status}
		got := Categorize(apiErr)
		assert.Equal(t, tt.expected, got, "status %d", tt.status)
	}
}

// MockAPIError for testing without importing api package
type MockAPIError struct {
	StatusCode int
	Message    string
}

func (e *MockAPIError) Error() string {
	return e.Message
}

func (e *MockAPIError) APIStatusCode() int {
	return e.StatusCode
}

func (e *MockAPIError) APIMessage() string {
	return e.Message
}

func TestSuggestionsFor(t *testing.T) {
	tests := []struct {
		category  Category
		operation string
		minCount  int
	}{
		{CategoryAuth, "listing people", 2},
		{CategoryForbidden, "getting contract", 2},
		{CategoryNotFound, "getting contract abc123", 1},
		{CategoryNotFound, "listing people", 1},
		{CategoryRateLimit, "any operation", 2},
		{CategoryServer, "any operation", 2},
		{CategoryNetwork, "any operation", 2},
		{CategoryConfig, "any operation", 2},
	}

	for _, tt := range tests {
		suggestions := SuggestionsFor(tt.category, tt.operation)
		assert.GreaterOrEqual(t, len(suggestions), tt.minCount,
			"category %v, operation %q", tt.category, tt.operation)
	}
}

func TestSuggestionsForNotFound_ContextAware(t *testing.T) {
	// Getting a specific resource should suggest checking the ID
	getSuggestions := SuggestionsFor(CategoryNotFound, "getting contract abc123")
	assert.Contains(t, getSuggestions[0], "ID")

	// Listing should suggest endpoint availability
	listSuggestions := SuggestionsFor(CategoryNotFound, "listing people")
	assert.Contains(t, listSuggestions[0], "available")
}

func TestWrap(t *testing.T) {
	apiErr := &MockAPIError{StatusCode: 404, Message: "not found"}

	wrapped := Wrap(apiErr, "getting contract xyz")

	assert.Equal(t, "getting contract xyz", wrapped.Operation)
	assert.Equal(t, CategoryNotFound, wrapped.Category)
	assert.NotEmpty(t, wrapped.Suggestions)
	assert.Contains(t, wrapped.Error(), "getting contract xyz")
}

func TestWrap_PreservesOriginalError(t *testing.T) {
	original := &MockAPIError{StatusCode: 401, Message: "unauthorized"}
	wrapped := Wrap(original, "listing people")

	// Should be able to unwrap to get original
	assert.Equal(t, original, wrapped.Unwrap())
}

func TestWrap_NilError(t *testing.T) {
	wrapped := Wrap(nil, "any operation")
	assert.Nil(t, wrapped)
}
