package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// EORAmendment represents a modification to an existing EOR contract
type EORAmendment struct {
	ID            string                 `json:"id"`
	ContractID    string                 `json:"contract_id"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`
	Changes       map[string]interface{} `json:"changes"`
	EffectiveDate string                 `json:"effective_date"`
	Reason        string                 `json:"reason,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	AcceptedAt    string                 `json:"accepted_at,omitempty"`
	SignedAt      string                 `json:"signed_at,omitempty"`
}

// CreateEORAmendmentParams are parameters for creating an EOR amendment
type CreateEORAmendmentParams struct {
	Type          string                 `json:"type"`
	Changes       map[string]interface{} `json:"changes"`
	EffectiveDate string                 `json:"effective_date"`
	Reason        string                 `json:"reason,omitempty"`
}

// CreateEORAmendment creates a new amendment for an EOR contract
func (c *Client) CreateEORAmendment(ctx context.Context, contractID string, params CreateEORAmendmentParams) (*EORAmendment, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s/amendments", escapePath(contractID))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORAmendment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// ListEORAmendments returns all amendments for an EOR contract
func (c *Client) ListEORAmendments(ctx context.Context, contractID string) ([]EORAmendment, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s/amendments", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []EORAmendment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// AcceptEORAmendment accepts an EOR amendment
func (c *Client) AcceptEORAmendment(ctx context.Context, amendmentID string) (*EORAmendment, error) {
	path := fmt.Sprintf("/rest/v2/eor/amendments/%s/accept", escapePath(amendmentID))
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORAmendment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// SignEORAmendment signs an EOR amendment
func (c *Client) SignEORAmendment(ctx context.Context, amendmentID string) (*EORAmendment, error) {
	path := fmt.Sprintf("/rest/v2/eor/amendments/%s/sign", escapePath(amendmentID))
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORAmendment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
