package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid ISO date",
			input:       "2024-01-15",
			expectError: false,
		},
		{
			name:        "valid ISO date end of month",
			input:       "2024-12-31",
			expectError: false,
		},
		{
			name:        "invalid format - slash separator",
			input:       "2024/01/15",
			expectError: true,
		},
		{
			name:        "invalid format - US style",
			input:       "01-15-2024",
			expectError: true,
		},
		{
			name:        "invalid format - no separator",
			input:       "20240115",
			expectError: true,
		},
		{
			name:        "invalid date - month 13",
			input:       "2024-13-01",
			expectError: true,
		},
		{
			name:        "invalid date - day 32",
			input:       "2024-01-32",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "random text",
			input:       "not-a-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDate(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid USD",
			input:       "USD",
			expectError: false,
		},
		{
			name:        "valid EUR",
			input:       "EUR",
			expectError: false,
		},
		{
			name:        "valid GBP",
			input:       "GBP",
			expectError: false,
		},
		{
			name:        "valid JPY",
			input:       "JPY",
			expectError: false,
		},
		{
			name:        "lowercase valid",
			input:       "usd",
			expectError: false,
		},
		{
			name:        "mixed case",
			input:       "Usd",
			expectError: false,
		},
		{
			name:        "invalid - too short",
			input:       "US",
			expectError: true,
		},
		{
			name:        "invalid - too long",
			input:       "USDD",
			expectError: true,
		},
		{
			name:        "invalid - numeric",
			input:       "123",
			expectError: true,
		},
		{
			name:        "invalid - mixed alphanumeric",
			input:       "US1",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid - single letter",
			input:       "U",
			expectError: true,
		},
		{
			name:        "invalid - special characters",
			input:       "U$D",
			expectError: true,
		},
		{
			name:        "invalid - unicode letters",
			input:       "ÜSD",
			expectError: true,
		},
		{
			name:        "invalid - spaces",
			input:       "U S",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCurrency(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid integer",
			input:       "100",
			expectError: false,
		},
		{
			name:        "valid decimal",
			input:       "100.50",
			expectError: false,
		},
		{
			name:        "valid with leading zero decimal",
			input:       "0.99",
			expectError: false,
		},
		{
			name:        "valid large amount",
			input:       "1000000.00",
			expectError: false,
		},
		{
			name:        "valid zero",
			input:       "0",
			expectError: false,
		},
		{
			name:        "negative amount",
			input:       "-100",
			expectError: true,
		},
		{
			name:        "invalid - letters",
			input:       "abc",
			expectError: true,
		},
		{
			name:        "invalid - currency symbol",
			input:       "$100",
			expectError: true,
		},
		{
			name:        "invalid - comma separator",
			input:       "1,000",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAmount(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name        string
		startDate   string
		endDate     string
		expectError bool
	}{
		{
			name:        "valid range - same day",
			startDate:   "2024-01-15",
			endDate:     "2024-01-15",
			expectError: false,
		},
		{
			name:        "valid range - different days",
			startDate:   "2024-01-01",
			endDate:     "2024-01-31",
			expectError: false,
		},
		{
			name:        "valid range - across months",
			startDate:   "2024-01-15",
			endDate:     "2024-03-15",
			expectError: false,
		},
		{
			name:        "invalid range - end before start",
			startDate:   "2024-01-31",
			endDate:     "2024-01-01",
			expectError: true,
		},
		{
			name:        "invalid - bad start date",
			startDate:   "not-a-date",
			endDate:     "2024-01-15",
			expectError: true,
		},
		{
			name:        "invalid - bad end date",
			startDate:   "2024-01-15",
			endDate:     "not-a-date",
			expectError: true,
		},
		{
			name:        "valid range - across years",
			startDate:   "2023-12-31",
			endDate:     "2024-01-01",
			expectError: false,
		},
		{
			name:        "invalid - both bad dates",
			startDate:   "bad",
			endDate:     "also-bad",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.startDate, tt.endDate)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertDateToRFC3339(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "valid date converts to RFC3339",
			input:    "2024-01-15",
			expected: "2024-01-15T00:00:00Z",
			hasError: false,
		},
		{
			name:     "end of year",
			input:    "2024-12-31",
			expected: "2024-12-31T00:00:00Z",
			hasError: false,
		},
		{
			name:     "invalid date",
			input:    "not-a-date",
			expected: "",
			hasError: true,
		},
		{
			name:     "leap day",
			input:    "2024-02-29",
			expected: "2024-02-29T00:00:00Z",
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
			hasError: true,
		},
		{
			name:     "first day of year",
			input:    "2024-01-01",
			expected: "2024-01-01T00:00:00Z",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertDateToRFC3339(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Parse both to compare as time values
				expected, _ := time.Parse(time.RFC3339, tt.expected)
				actual, _ := time.Parse(time.RFC3339, result)
				assert.True(t, expected.Equal(actual), "expected %s, got %s", tt.expected, result)
			}
		})
	}
}
