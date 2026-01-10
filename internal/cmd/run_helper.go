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
