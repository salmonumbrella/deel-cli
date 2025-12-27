package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Currency represents a currency code and details
type Currency struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// Country represents a country code and name
type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// JobTitle represents a job title
type JobTitle struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SeniorityLevel represents a seniority level
type SeniorityLevel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TimeOffType represents a type of time off
type TimeOffType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListCurrencies returns all available currencies
func (c *Client) ListCurrencies(ctx context.Context) ([]Currency, error) {
	resp, err := c.Get(ctx, "/rest/v2/lookups/currencies")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Currency `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// ListCountries returns all available countries
func (c *Client) ListCountries(ctx context.Context) ([]Country, error) {
	resp, err := c.Get(ctx, "/rest/v2/lookups/countries")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Country `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// ListJobTitles returns all available job titles
func (c *Client) ListJobTitles(ctx context.Context) ([]JobTitle, error) {
	resp, err := c.Get(ctx, "/rest/v2/lookups/job-titles")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []JobTitle `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// ListSeniorityLevels returns all available seniority levels
func (c *Client) ListSeniorityLevels(ctx context.Context) ([]SeniorityLevel, error) {
	resp, err := c.Get(ctx, "/rest/v2/lookups/seniority-levels")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []SeniorityLevel `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// ListTimeOffTypes returns all available time off types
func (c *Client) ListTimeOffTypes(ctx context.Context) ([]TimeOffType, error) {
	resp, err := c.Get(ctx, "/rest/v2/lookups/time-off-types")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []TimeOffType `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}
