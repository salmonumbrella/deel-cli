package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// GPShiftRate represents a shift rate for Global Payroll workers
type GPShiftRate struct {
	ID          string  `json:"id"`
	ExternalID  string  `json:"external_id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Rate        float64 `json:"rate"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"` // "hourly", "daily", "flat"
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// GPCreateShiftRateParams are parameters for creating a shift rate
type GPCreateShiftRateParams struct {
	ExternalID  string  `json:"external_id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Rate        float64 `json:"rate"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"`
}

// GPUpdateShiftRateParams are parameters for updating a shift rate
type GPUpdateShiftRateParams struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Rate        float64 `json:"rate,omitempty"`
}

// CreateGPShiftRate creates a new shift rate for Global Payroll workers
func (c *Client) CreateGPShiftRate(ctx context.Context, params GPCreateShiftRateParams) (*GPShiftRate, error) {
	resp, err := c.Post(ctx, "/rest/v2/gp/shift-rates", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data GPShiftRate `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// ListGPShiftRates retrieves all shift rates for Global Payroll workers
func (c *Client) ListGPShiftRates(ctx context.Context) ([]GPShiftRate, error) {
	resp, err := c.Get(ctx, "/rest/v2/gp/shift-rates")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []GPShiftRate `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// UpdateGPShiftRate updates an existing shift rate
func (c *Client) UpdateGPShiftRate(ctx context.Context, rateID string, params GPUpdateShiftRateParams) (*GPShiftRate, error) {
	path := fmt.Sprintf("/rest/v2/gp/shift-rates/%s", escapePath(rateID))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data GPShiftRate `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteGPShiftRate deletes a shift rate by external ID
func (c *Client) DeleteGPShiftRate(ctx context.Context, externalID string) error {
	path := fmt.Sprintf("/rest/v2/gp/shift-rates/external/%s", escapePath(externalID))
	_, err := c.Delete(ctx, path)
	return err
}
