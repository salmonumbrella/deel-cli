package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// dateFormat is the expected date format (ISO 8601 date)
const dateFormat = "2006-01-02"

// validateDate validates that a string is a valid date in YYYY-MM-DD format.
func validateDate(date string) error {
	if date == "" {
		return fmt.Errorf("date cannot be empty")
	}
	_, err := time.Parse(dateFormat, date)
	if err != nil {
		return fmt.Errorf("invalid date format %q (expected YYYY-MM-DD)", date)
	}
	return nil
}

// validateCurrency validates that a string is a valid 3-letter ISO 4217 currency code.
func validateCurrency(currency string) error {
	if currency == "" {
		return fmt.Errorf("currency cannot be empty")
	}
	// ISO 4217 currency codes are exactly 3 uppercase letters
	upper := strings.ToUpper(currency)
	if len(upper) != 3 {
		return fmt.Errorf("invalid currency code %q (must be 3 letters)", currency)
	}
	for _, r := range upper {
		if r < 'A' || r > 'Z' {
			return fmt.Errorf("invalid currency code %q (must contain only letters)", currency)
		}
	}
	return nil
}

// validateAmount validates that a string is a valid positive monetary amount.
func validateAmount(amount string) error {
	if amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q (must be a number)", amount)
	}
	if val < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	return nil
}

// validateDateRange validates that start date is not after end date.
func validateDateRange(startDate, endDate string) error {
	if err := validateDate(startDate); err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}
	if err := validateDate(endDate); err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}

	start, _ := time.Parse(dateFormat, startDate)
	end, _ := time.Parse(dateFormat, endDate)

	if start.After(end) {
		return fmt.Errorf("start date %s cannot be after end date %s", startDate, endDate)
	}
	return nil
}

// convertDateToRFC3339 converts a YYYY-MM-DD date to RFC3339 format.
func convertDateToRFC3339(date string) (string, error) {
	if err := validateDate(date); err != nil {
		return "", err
	}
	t, _ := time.Parse(dateFormat, date)
	return t.Format(time.RFC3339), nil
}
