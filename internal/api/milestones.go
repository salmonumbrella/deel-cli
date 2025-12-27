package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Milestone represents a contract milestone
type Milestone struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	DueDate     string  `json:"due_date"`
}

// ListMilestones returns milestones for a contract
func (c *Client) ListMilestones(ctx context.Context, contractID string) ([]Milestone, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/milestones", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Milestone `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// GetMilestone returns a single milestone
func (c *Client) GetMilestone(ctx context.Context, contractID, milestoneID string) (*Milestone, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/milestones/%s", escapePath(contractID), escapePath(milestoneID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Milestone `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// CreateMilestoneParams are params for creating a milestone
type CreateMilestoneParams struct {
	ContractID  string  `json:"contract_id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount"`
	DueDate     string  `json:"due_date,omitempty"`
}

// CreateMilestone creates a new milestone
func (c *Client) CreateMilestone(ctx context.Context, params CreateMilestoneParams) (*Milestone, error) {
	resp, err := c.Post(ctx, "/rest/v2/milestones", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Milestone `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteMilestone deletes a milestone
func (c *Client) DeleteMilestone(ctx context.Context, milestoneID string) error {
	path := fmt.Sprintf("/rest/v2/milestones/%s", escapePath(milestoneID))
	_, err := c.Delete(ctx, path)
	return err
}
