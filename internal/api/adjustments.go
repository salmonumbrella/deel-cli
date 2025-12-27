package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Adjustment represents a contract adjustment (bonus, deduction, expense)
type Adjustment struct {
	ID          string  `json:"id"`
	ContractID  string  `json:"contract_id"`
	CategoryID  string  `json:"category_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// AdjustmentCategory represents a category of adjustments
type AdjustmentCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"` // "bonus", "deduction", "expense"
}

// CreateAdjustmentParams are parameters for creating an adjustment
type CreateAdjustmentParams struct {
	ContractID  string  `json:"contract_id"`
	CategoryID  string  `json:"category_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

// UpdateAdjustmentParams are parameters for updating an adjustment
type UpdateAdjustmentParams struct {
	Amount      float64 `json:"amount,omitempty"`
	Description string  `json:"description,omitempty"`
	Date        string  `json:"date,omitempty"`
}

// ListAdjustmentsParams are parameters for listing adjustments
type ListAdjustmentsParams struct {
	ContractID string
	CategoryID string
}

// CreateAdjustment creates a new adjustment
func (c *Client) CreateAdjustment(ctx context.Context, params CreateAdjustmentParams) (*Adjustment, error) {
	resp, err := c.Post(ctx, "/rest/v2/adjustments", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// GetAdjustment returns a single adjustment
func (c *Client) GetAdjustment(ctx context.Context, id string) (*Adjustment, error) {
	path := fmt.Sprintf("/rest/v2/adjustments/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateAdjustment updates an existing adjustment
func (c *Client) UpdateAdjustment(ctx context.Context, id string, params UpdateAdjustmentParams) (*Adjustment, error) {
	path := fmt.Sprintf("/rest/v2/adjustments/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteAdjustment deletes an adjustment
func (c *Client) DeleteAdjustment(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/v2/adjustments/%s", escapePath(id))
	_, err := c.Delete(ctx, path)
	return err
}

// ListAdjustments returns adjustments with optional filters
func (c *Client) ListAdjustments(ctx context.Context, params ListAdjustmentsParams) ([]Adjustment, error) {
	q := url.Values{}
	if params.ContractID != "" {
		q.Set("contract_id", params.ContractID)
	}
	if params.CategoryID != "" {
		q.Set("category_id", params.CategoryID)
	}

	path := "/rest/v2/adjustments"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// ListAdjustmentCategories returns all available adjustment categories
func (c *Client) ListAdjustmentCategories(ctx context.Context) ([]AdjustmentCategory, error) {
	resp, err := c.Get(ctx, "/rest/v2/adjustments/categories")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []AdjustmentCategory `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}
