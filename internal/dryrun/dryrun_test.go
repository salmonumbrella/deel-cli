package dryrun

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDryRun(t *testing.T) {
	ctx := context.Background()

	// Not enabled by default
	assert.False(t, IsEnabled(ctx))

	// Enable dry-run
	ctx = WithDryRun(ctx, true)
	assert.True(t, IsEnabled(ctx))

	// Disable dry-run
	ctx = WithDryRun(ctx, false)
	assert.False(t, IsEnabled(ctx))
}

func TestPreview_Write(t *testing.T) {
	var buf bytes.Buffer

	preview := Preview{
		Operation:   "CREATE",
		Resource:    "adjustment",
		Description: "Create bonus adjustment for contractor",
		Details: map[string]string{
			"Contractor": "John Doe",
			"Amount":     "$500.00 USD",
			"Type":       "bonus",
		},
		Warnings: []string{"This will affect payroll calculations"},
	}

	err := preview.Write(&buf)
	require.NoError(t, err)

	output := buf.String()

	// Check operation is present
	assert.Contains(t, output, "[DRY-RUN]")
	assert.Contains(t, output, "CREATE")

	// Check resource
	assert.Contains(t, output, "adjustment")

	// Check description
	assert.Contains(t, output, "Create bonus adjustment for contractor")

	// Check details
	assert.Contains(t, output, "Contractor")
	assert.Contains(t, output, "John Doe")
	assert.Contains(t, output, "Amount")
	assert.Contains(t, output, "$500.00 USD")

	// Check warnings
	assert.Contains(t, output, "Warning")
	assert.Contains(t, output, "This will affect payroll calculations")
}

func TestPreview_Write_NoWarnings(t *testing.T) {
	var buf bytes.Buffer

	preview := Preview{
		Operation:   "UPDATE",
		Resource:    "contract",
		Description: "Update contract end date",
	}

	err := preview.Write(&buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "UPDATE")
	assert.Contains(t, output, "contract")
	assert.NotContains(t, output, "Warning")
}

func TestPreview_Write_MinimalDetails(t *testing.T) {
	var buf bytes.Buffer

	preview := Preview{
		Operation: "DELETE",
		Resource:  "webhook",
	}

	err := preview.Write(&buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "[DRY-RUN]")
	assert.Contains(t, output, "DELETE")
	assert.Contains(t, output, "webhook")
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		want     string
	}{
		{
			name:     "USD",
			amount:   1234.56,
			currency: "USD",
			want:     "$1,234.56 USD",
		},
		{
			name:     "EUR",
			amount:   1000.00,
			currency: "EUR",
			want:     "1,000.00 EUR",
		},
		{
			name:     "GBP",
			amount:   99.99,
			currency: "GBP",
			want:     "99.99 GBP",
		},
		{
			name:     "large amount",
			amount:   1234567.89,
			currency: "USD",
			want:     "$1,234,567.89 USD",
		},
		{
			name:     "zero",
			amount:   0,
			currency: "USD",
			want:     "$0.00 USD",
		},
		{
			name:     "negative",
			amount:   -100.50,
			currency: "USD",
			want:     "-$100.50 USD",
		},
		{
			name:     "JPY no decimals",
			amount:   5000,
			currency: "JPY",
			want:     "5,000 JPY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAmount(tt.amount, tt.currency)
			assert.Equal(t, tt.want, got)
		})
	}
}
