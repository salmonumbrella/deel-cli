package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// HourlyPreset represents a template for hourly contract configurations
type HourlyPreset struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	HoursPerDay  float64 `json:"hours_per_day"`
	HoursPerWeek float64 `json:"hours_per_week"`
	Rate         float64 `json:"rate,omitempty"`
	Currency     string  `json:"currency,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// ListHourlyPresets returns all hourly presets
func (c *Client) ListHourlyPresets(ctx context.Context) ([]HourlyPreset, error) {
	resp, err := c.Get(ctx, "/rest/v2/hourly-presets")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []HourlyPreset `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// CreateHourlyPresetParams are params for creating an hourly preset
type CreateHourlyPresetParams struct {
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	HoursPerDay  float64 `json:"hours_per_day"`
	HoursPerWeek float64 `json:"hours_per_week"`
	Rate         float64 `json:"rate,omitempty"`
	Currency     string  `json:"currency,omitempty"`
}

// CreateHourlyPreset creates a new hourly preset
func (c *Client) CreateHourlyPreset(ctx context.Context, params CreateHourlyPresetParams) (*HourlyPreset, error) {
	resp, err := c.Post(ctx, "/rest/v2/hourly-presets", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data HourlyPreset `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateHourlyPresetParams are params for updating an hourly preset
type UpdateHourlyPresetParams struct {
	Name         string  `json:"name,omitempty"`
	Description  string  `json:"description,omitempty"`
	HoursPerDay  float64 `json:"hours_per_day,omitempty"`
	HoursPerWeek float64 `json:"hours_per_week,omitempty"`
	Rate         float64 `json:"rate,omitempty"`
	Currency     string  `json:"currency,omitempty"`
}

// UpdateHourlyPreset updates an existing hourly preset
func (c *Client) UpdateHourlyPreset(ctx context.Context, presetID string, params UpdateHourlyPresetParams) (*HourlyPreset, error) {
	path := fmt.Sprintf("/rest/v2/hourly-presets/%s", escapePath(presetID))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data HourlyPreset `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteHourlyPreset deletes an hourly preset
func (c *Client) DeleteHourlyPreset(ctx context.Context, presetID string) error {
	path := fmt.Sprintf("/rest/v2/hourly-presets/%s", escapePath(presetID))
	_, err := c.Delete(ctx, path)
	return err
}
