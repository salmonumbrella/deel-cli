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
