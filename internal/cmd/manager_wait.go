package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

const managerWaitInterval = 5 * time.Second

func waitForPersonByEmail(ctx context.Context, client *api.Client, email string, timeout time.Duration) (*api.Person, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}

	deadline := time.Now().Add(timeout)
	var lastErr error

	for {
		person, err := client.SearchPeopleByEmail(ctx, email)
		if err == nil && person != nil {
			if person.HRISProfileID != "" {
				return person, nil
			}
			lastErr = fmt.Errorf("worker profile not ready")
		} else if err != nil {
			if !isNotFoundError(err) {
				return nil, err
			}
			lastErr = fmt.Errorf("worker not found yet")
		}

		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, fmt.Errorf("timed out waiting for worker: %w", lastErr)
			}
			return nil, fmt.Errorf("timed out waiting for worker to appear")
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(managerWaitInterval):
		}
	}
}

func isNotFoundError(err error) bool {
	var apiErr *api.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}
