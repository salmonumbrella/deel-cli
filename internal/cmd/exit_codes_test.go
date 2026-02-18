package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/pflag"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/climerrors"
)

func TestExitCodeMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code int
	}{
		{"nil", nil, exitOK},
		{"help", pflag.ErrHelp, exitOK},
		{"api 401", &api.APIError{StatusCode: 401, Message: "unauthorized"}, exitAuth},
		{"api 403", &api.APIError{StatusCode: 403, Message: "forbidden"}, exitForbidden},
		{"api 404", &api.APIError{StatusCode: 404, Message: "not found"}, exitNotFound},
		{"api 429", &api.APIError{StatusCode: 429, Message: "rate limited"}, exitRateLimited},
		{"api 500", &api.APIError{StatusCode: 500, Message: "server error"}, exitServer},
		{"api 400", &api.APIError{StatusCode: 400, Message: "bad request"}, exitUsage},
		{"usage", errors.New("unknown command \"nope\""), exitUsage},
		{"usage shorthand", errors.New("unknown shorthand flag: 'a' in -a"), exitUsage},
		{"network", errors.New("dial tcp: connection refused"), exitNetwork},
		{"generic", errors.New("boom"), exitGeneric},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ExitCode(tc.err); got != tc.code {
				t.Fatalf("ExitCode(%v) = %d, want %d", tc.err, got, tc.code)
			}
		})
	}
}

func TestExitCodeFromCLIError(t *testing.T) {
	cases := []struct {
		name     string
		category climerrors.Category
		code     int
	}{
		{"auth", climerrors.CategoryAuth, exitAuth},
		{"forbidden", climerrors.CategoryForbidden, exitForbidden},
		{"not found", climerrors.CategoryNotFound, exitNotFound},
		{"rate limit", climerrors.CategoryRateLimit, exitRateLimited},
		{"server", climerrors.CategoryServer, exitServer},
		{"network", climerrors.CategoryNetwork, exitNetwork},
		{"validation", climerrors.CategoryValidation, exitUsage},
		{"config", climerrors.CategoryConfig, exitUsage},
		{"unknown", climerrors.CategoryUnknown, exitGeneric},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := &climerrors.CLIError{
				Operation: "testing",
				Err:       errors.New("test"),
				Category:  tc.category,
			}
			got := ExitCode(err)
			if got != tc.code {
				t.Fatalf("ExitCode(CLIError{%s}) = %d, want %d", tc.name, got, tc.code)
			}
		})
	}
}
