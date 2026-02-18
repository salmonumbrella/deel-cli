// Package dryrun provides dry-run mode for previewing mutations.
package dryrun

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
)

type ctxKey struct{}

// WithDryRun returns a context with dry-run mode enabled or disabled.
func WithDryRun(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, ctxKey{}, enabled)
}

// IsEnabled returns true if dry-run mode is enabled in the context.
func IsEnabled(ctx context.Context) bool {
	v, ok := ctx.Value(ctxKey{}).(bool)
	return ok && v
}

// Preview represents a preview of an operation that would be performed.
type Preview struct {
	Operation   string            // CREATE, UPDATE, DELETE, etc.
	Resource    string            // Type of resource being affected
	Description string            // Human-readable description
	Details     map[string]string // Key-value pairs of details
	Warnings    []string          // Any warnings about the operation
}

// Write outputs the preview to the given writer.
func (p *Preview) Write(w io.Writer) error {
	var sb strings.Builder

	// Header
	fmt.Fprintf(&sb, "[DRY-RUN] %s %s\n", p.Operation, p.Resource)

	// Description
	if p.Description != "" {
		fmt.Fprintf(&sb, "  %s\n", p.Description)
	}

	// Details
	if len(p.Details) > 0 {
		sb.WriteString("\n  Details:\n")
		// Find max key length for alignment
		maxLen := 0
		for k := range p.Details {
			if len(k) > maxLen {
				maxLen = len(k)
			}
		}
		for k, v := range p.Details {
			fmt.Fprintf(&sb, "    %-*s: %s\n", maxLen, k, v)
		}
	}

	// Warnings
	if len(p.Warnings) > 0 {
		sb.WriteString("\n  Warnings:\n")
		for _, w := range p.Warnings {
			fmt.Fprintf(&sb, "    - Warning: %s\n", w)
		}
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

// FormatAmount formats a monetary amount with the appropriate currency symbol.
func FormatAmount(amount float64, currency string) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	// Currencies without decimal places
	noDecimals := map[string]bool{
		"JPY": true,
		"KRW": true,
		"VND": true,
	}

	// Format with appropriate decimal places
	var formatted string
	if noDecimals[currency] {
		formatted = formatWithCommas(int64(math.Round(amount)))
	} else {
		formatted = formatWithCommasDecimal(amount)
	}

	// Add currency symbol prefix
	prefix := ""
	switch currency {
	case "USD":
		prefix = "$"
	}

	// Build final string
	result := prefix + formatted + " " + currency
	if negative {
		if prefix != "" {
			result = "-" + prefix + formatted + " " + currency
		} else {
			result = "-" + formatted + " " + currency
		}
	}

	return result
}

func formatWithCommas(n int64) string {
	str := fmt.Sprintf("%d", n)
	return addCommas(str)
}

func formatWithCommasDecimal(n float64) string {
	intPart := int64(n)
	decPart := int64(math.Round((n - float64(intPart)) * 100))

	intStr := addCommas(fmt.Sprintf("%d", intPart))
	return fmt.Sprintf("%s.%02d", intStr, decPart)
}

func addCommas(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}

	var result strings.Builder
	offset := n % 3
	if offset == 0 {
		offset = 3
	}

	result.WriteString(s[:offset])
	for i := offset; i < n; i += 3 {
		result.WriteString(",")
		result.WriteString(s[i : i+3])
	}

	return result.String()
}
