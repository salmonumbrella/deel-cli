package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// BackgroundCheckOption represents a background check option
type BackgroundCheckOption struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Provider string  `json:"provider"`
	Country  string  `json:"country"`
	Type     string  `json:"type"`
	Cost     float64 `json:"cost"`
	Currency string  `json:"currency"`
	Duration string  `json:"estimated_duration"`
}

// ListBackgroundCheckOptions returns available check options
func (c *Client) ListBackgroundCheckOptions(ctx context.Context, country string) ([]BackgroundCheckOption, error) {
	path := fmt.Sprintf("/rest/v2/background-checks/options?country=%s", escapePath(country))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []BackgroundCheckOption `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// BackgroundCheck represents a background check
type BackgroundCheck struct {
	ID          string `json:"id"`
	ContractID  string `json:"contract_id"`
	WorkerName  string `json:"worker_name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Provider    string `json:"provider"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
	Result      string `json:"result"`
}

// ListBackgroundChecksByContract returns checks for a contract
func (c *Client) ListBackgroundChecksByContract(ctx context.Context, contractID string) ([]BackgroundCheck, error) {
	path := fmt.Sprintf("/rest/v2/background-checks/contracts/%s", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []BackgroundCheck `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}
