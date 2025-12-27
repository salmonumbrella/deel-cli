package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// GPShift represents a Global Payroll shift
type GPShift struct {
	ID           string  `json:"id"`
	ExternalID   string  `json:"external_id,omitempty"`
	WorkerID     string  `json:"worker_id"`
	Date         string  `json:"date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	BreakMinutes int     `json:"break_minutes"`
	TotalHours   float64 `json:"total_hours"`
	RateID       string  `json:"rate_id,omitempty"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

// CreateGPShiftParams are parameters for creating a GP shift
type CreateGPShiftParams struct {
	ExternalID   string `json:"external_id,omitempty"`
	WorkerID     string `json:"worker_id"`
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	BreakMinutes int    `json:"break_minutes"`
	RateID       string `json:"rate_id,omitempty"`
}

// UpdateGPShiftParams are parameters for updating a GP shift
type UpdateGPShiftParams struct {
	Date         string `json:"date,omitempty"`
	StartTime    string `json:"start_time,omitempty"`
	EndTime      string `json:"end_time,omitempty"`
	BreakMinutes int    `json:"break_minutes,omitempty"`
	RateID       string `json:"rate_id,omitempty"`
}

// CreateGPShift creates a new Global Payroll shift
func (c *Client) CreateGPShift(ctx context.Context, params CreateGPShiftParams) (*GPShift, error) {
	resp, err := c.Post(ctx, "/rest/v2/gp/shifts", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data GPShift `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateGPShift updates an existing Global Payroll shift
func (c *Client) UpdateGPShift(ctx context.Context, shiftID string, params UpdateGPShiftParams) (*GPShift, error) {
	path := fmt.Sprintf("/rest/v2/gp/shifts/%s", escapePath(shiftID))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data GPShift `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteGPShift deletes a Global Payroll shift by external ID
func (c *Client) DeleteGPShift(ctx context.Context, externalID string) error {
	path := fmt.Sprintf("/rest/v2/gp/shifts/external/%s", escapePath(externalID))
	_, err := c.Delete(ctx, path)
	return err
}
