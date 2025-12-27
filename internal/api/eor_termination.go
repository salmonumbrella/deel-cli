package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// EORTermination represents a contract termination (resignation or termination)
type EORTermination struct {
	ID               string  `json:"id"`
	ContractID       string  `json:"contract_id"`
	Type             string  `json:"type"` // "resignation" or "termination"
	Status           string  `json:"status"`
	Reason           string  `json:"reason"`
	EffectiveDate    string  `json:"effective_date"`
	LastWorkingDay   string  `json:"last_working_day"`
	NoticePeriodDays int     `json:"notice_period_days"`
	SeveranceAmount  float64 `json:"severance_amount,omitempty"`
	Currency         string  `json:"currency,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

// RequestResignationParams are parameters for requesting a resignation
type RequestResignationParams struct {
	Reason        string `json:"reason"`
	EffectiveDate string `json:"effective_date"`
}

// RequestTerminationParams are parameters for requesting a termination
type RequestTerminationParams struct {
	Reason        string `json:"reason"`
	EffectiveDate string `json:"effective_date"`
	WithCause     bool   `json:"with_cause"`
}

// RequestEORResignation creates a resignation request for an EOR contract
func (c *Client) RequestEORResignation(ctx context.Context, contractID string, params RequestResignationParams) (*EORTermination, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s/resignation-request", escapePath(contractID))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORTermination `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// RequestEORTermination creates a termination request for an EOR contract
func (c *Client) RequestEORTermination(ctx context.Context, contractID string, params RequestTerminationParams) (*EORTermination, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s/termination-request", escapePath(contractID))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORTermination `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// GetEORTermination retrieves the termination details for an EOR contract
func (c *Client) GetEORTermination(ctx context.Context, contractID string) (*EORTermination, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s/termination", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORTermination `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
