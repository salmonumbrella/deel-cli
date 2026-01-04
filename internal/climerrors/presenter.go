package climerrors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Messager interface for errors that have a raw message
type Messager interface {
	APIMessage() string
}

// FriendlyMessage extracts a clean message from any error
func FriendlyMessage(err error) string {
	// Try to get raw message from API error
	if m, ok := err.(Messager); ok {
		msg := m.APIMessage()
		// Try to parse JSON error message
		if parsed := parseAPIMessage(msg); parsed != "" {
			return parsed
		}
		// If message is not JSON, return as-is
		if msg != "" && msg[0] != '{' {
			return msg
		}
	}

	// Fall back to status text for API errors
	if sc, ok := err.(StatusCoder); ok {
		return http.StatusText(sc.APIStatusCode())
	}

	return err.Error()
}

// parseAPIMessage tries to extract message from Deel API JSON error format
// Format: {"errors":[{"message":"..."}]}
func parseAPIMessage(raw string) string {
	if raw == "" || raw[0] != '{' {
		return ""
	}

	var parsed struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return ""
	}
	if len(parsed.Errors) > 0 && parsed.Errors[0].Message != "" {
		return parsed.Errors[0].Message
	}
	return ""
}

// FormatError writes a formatted error to the writer
func FormatError(w io.Writer, err *CLIError) {
	msg := FriendlyMessage(err.Err)
	if _, writeErr := fmt.Fprintf(w, "Failed %s: %s\n", err.Operation, msg); writeErr != nil {
		return
	}

	if len(err.Suggestions) > 0 {
		if _, writeErr := fmt.Fprintln(w); writeErr != nil {
			return
		}
		for _, s := range err.Suggestions {
			if _, writeErr := fmt.Fprintf(w, "  -> %s\n", s); writeErr != nil {
				return
			}
		}
	}
}
