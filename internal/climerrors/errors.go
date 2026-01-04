// Package climerrors provides error handling, categorization, and user-friendly
// error messages for the Deel CLI.
package climerrors

import (
	"errors"
	"fmt"
	"net"
)

// Category represents the type of error for suggestion lookup
type Category int

const (
	// CategoryUnknown is used when no specific category matches.
	CategoryUnknown Category = iota
	// CategoryAuth represents authentication failures (401).
	CategoryAuth
	// CategoryForbidden represents permission failures (403).
	CategoryForbidden
	// CategoryNotFound represents missing resources (404).
	CategoryNotFound
	// CategoryValidation represents validation failures (400, 422).
	CategoryValidation
	// CategoryRateLimit represents rate-limit responses (429).
	CategoryRateLimit
	// CategoryServer represents server errors (500+).
	CategoryServer
	// CategoryNetwork represents connection failures.
	CategoryNetwork
	// CategoryConfig represents missing configuration.
	CategoryConfig
)

// CLIError wraps any error with context and suggestions
type CLIError struct {
	Operation   string
	Err         error
	Suggestions []string
	Category    Category
}

func (e *CLIError) Error() string {
	return fmt.Sprintf("failed %s: %s", e.Operation, e.Err)
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

// StatusCoder interface for errors that have HTTP status codes
type StatusCoder interface {
	APIStatusCode() int
}

// Categorize determines the error category from an error
func Categorize(err error) Category {
	// Check for API errors with status codes
	if sc, ok := err.(StatusCoder); ok {
		return categoryFromStatus(sc.APIStatusCode())
	}

	// Check for network errors
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return CategoryNetwork
	}

	return CategoryUnknown
}

func categoryFromStatus(status int) Category {
	switch {
	case status == 401:
		return CategoryAuth
	case status == 403:
		return CategoryForbidden
	case status == 404:
		return CategoryNotFound
	case status == 400 || status == 422:
		return CategoryValidation
	case status == 429:
		return CategoryRateLimit
	case status >= 500:
		return CategoryServer
	default:
		return CategoryUnknown
	}
}

// Wrap creates a CLIError with context and suggestions
func Wrap(err error, operation string) *CLIError {
	if err == nil {
		return nil
	}

	cat := Categorize(err)
	return &CLIError{
		Operation:   operation,
		Err:         err,
		Category:    cat,
		Suggestions: SuggestionsFor(cat, operation),
	}
}
