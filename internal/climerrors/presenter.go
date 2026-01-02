package climerrors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FriendlyMessage extracts a clean message from any error
func FriendlyMessage(err error) string {
	if sc, ok := err.(StatusCoder); ok {
		msg := err.Error()
		// Try to parse JSON error message
		if parsed := parseAPIMessage(msg); parsed != "" {
			return parsed
		}
		// If message is empty or just the raw JSON, use status text
		if msg == "" || (len(msg) > 0 && msg[0] == '{') {
			return http.StatusText(sc.APIStatusCode())
		}
		return msg
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
	fmt.Fprintf(w, "Failed %s: %s\n", err.Operation, msg)

	if len(err.Suggestions) > 0 {
		fmt.Fprintln(w)
		for _, s := range err.Suggestions {
			fmt.Fprintf(w, "  -> %s\n", s)
		}
	}
}
