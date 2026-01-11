// Package cmd provides command implementations for the Deel CLI.
//
// # Client Initialization Patterns
//
// Commands have two patterns for initializing the API client and formatter:
//
// Pattern 1: initClient (combined)
//
//	client, f, err := initClient("operation")
//	if err != nil {
//	    return err
//	}
//
// Use when: The command has no flag/argument validation before the API call.
// The formatter and client are created together with consistent error handling.
//
// Pattern 2: getFormatter + getClient (separate)
//
//	f := getFormatter()
//	if err := validateFlags(); err != nil {
//	    return HandleError(f, err, "operation")
//	}
//	client, err := getClient()
//
// Use when: The command needs to validate flags or arguments BEFORE creating
// the API client. This avoids unnecessary API initialization (token loading,
// config parsing) when validation will fail anyway.
//
// Choose the separate pattern if you need to return validation errors early.
// Choose initClient for simpler commands with no pre-validation.
package cmd

import (
	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

// initClient returns a formatter and initialized API client, handling errors consistently.
func initClient(operation string) (*api.Client, *outfmt.Formatter, error) {
	f := getFormatter()
	client, err := getClient()
	if err != nil {
		return nil, f, HandleError(f, err, operation)
	}
	return client, f, nil
}
