// internal/climerrors/presenter_test.go
package climerrors

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFriendlyMessage_APIError(t *testing.T) {
	// Raw JSON message should be parsed
	err := &MockAPIError{
		StatusCode: 404,
		Message:    `{"errors":[{"message":"Resource not found"}]}`,
	}
	msg := FriendlyMessage(err)
	assert.Equal(t, "Resource not found", msg)
}

func TestFriendlyMessage_PlainMessage(t *testing.T) {
	err := &MockAPIError{
		StatusCode: 404,
		Message:    "Not Found",
	}
	msg := FriendlyMessage(err)
	assert.Equal(t, "Not Found", msg)
}

func TestFriendlyMessage_EmptyMessage(t *testing.T) {
	err := &MockAPIError{
		StatusCode: 404,
		Message:    "",
	}
	msg := FriendlyMessage(err)
	assert.Equal(t, "Not Found", msg) // Falls back to HTTP status text
}

func TestFriendlyMessage_StandardError(t *testing.T) {
	err := errors.New("connection refused")
	msg := FriendlyMessage(err)
	assert.Equal(t, "connection refused", msg)
}

func TestFormatError(t *testing.T) {
	cliErr := &CLIError{
		Operation:   "listing people",
		Err:         errors.New("not found"),
		Suggestions: []string{"Check the ID", "Try again"},
	}

	var buf bytes.Buffer
	FormatError(&buf, cliErr)
	output := buf.String()

	assert.Contains(t, output, "Failed listing people")
	assert.Contains(t, output, "not found")
	assert.Contains(t, output, "Check the ID")
	assert.Contains(t, output, "Try again")
}

func TestFormatError_NoSuggestions(t *testing.T) {
	cliErr := &CLIError{
		Operation:   "doing something",
		Err:         errors.New("failed"),
		Suggestions: nil,
	}

	var buf bytes.Buffer
	FormatError(&buf, cliErr)
	output := buf.String()

	assert.Contains(t, output, "Failed doing something")
	assert.Contains(t, output, "failed")
	// Should not have suggestion arrows
	assert.NotContains(t, output, "->")
}
