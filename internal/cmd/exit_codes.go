package cmd

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/pflag"

	"github.com/salmonumbrella/deel-cli/internal/climerrors"
)

const (
	exitOK          = 0
	exitGeneric     = 1
	exitUsage       = 2
	exitAuth        = 3
	exitNotFound    = 4
	exitForbidden   = 5
	exitRateLimited = 6
	exitServer      = 7
	exitNetwork     = 8
)

// ExitCode maps an error to a process exit code.
func ExitCode(err error) int {
	if err == nil {
		return exitOK
	}
	if errors.Is(err, pflag.ErrHelp) {
		return exitOK
	}

	if code := exitCodeFromCLIError(err); code != 0 {
		return code
	}
	if code := exitCodeFromAPIError(err); code != 0 {
		return code
	}
	if isUsageError(err) {
		return exitUsage
	}
	if isNetworkError(err) {
		return exitNetwork
	}
	return exitGeneric
}

func exitCodeFromCLIError(err error) int {
	var cliErr *climerrors.CLIError
	if !errors.As(err, &cliErr) {
		return 0
	}
	return exitCodeFromCategory(cliErr.Category)
}

func exitCodeFromAPIError(err error) int {
	var sc climerrors.StatusCoder
	if !errors.As(err, &sc) {
		return 0
	}
	cat := climerrors.Categorize(err)
	return exitCodeFromCategory(cat)
}

func exitCodeFromCategory(cat climerrors.Category) int {
	switch cat {
	case climerrors.CategoryAuth:
		return exitAuth
	case climerrors.CategoryForbidden:
		return exitForbidden
	case climerrors.CategoryNotFound:
		return exitNotFound
	case climerrors.CategoryRateLimit:
		return exitRateLimited
	case climerrors.CategoryServer:
		return exitServer
	case climerrors.CategoryNetwork:
		return exitNetwork
	case climerrors.CategoryValidation, climerrors.CategoryConfig:
		return exitUsage
	default:
		return 0
	}
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "tls") ||
		strings.Contains(msg, "certificate") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "timeout")
}

func isUsageError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	indicators := []string{
		"unknown command",
		"unknown flag",
		"unknown shorthand flag",
		"flag needs an argument",
		"flag provided but not defined",
		"requires at least",
		"requires exactly",
		"invalid argument",
		"invalid value",
		"must be",
		"is required",
		"missing",
	}
	for _, indicator := range indicators {
		if strings.Contains(msg, indicator) {
			return true
		}
	}
	return false
}
