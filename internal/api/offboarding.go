package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Offboarding represents an offboarding record
type Offboarding struct {
	ID            string `json:"id"`
	ContractID    string `json:"contract_id"`
	WorkerName    string `json:"worker_name"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	EffectiveDate string `json:"effective_date"`
	CreatedAt     string `json:"created_at"`
}

// TerminationDetails represents detailed termination information
type TerminationDetails struct {
	ID            string `json:"id"`
	ContractID    string `json:"contract_id"`
	Reason        string `json:"reason"`
	Status        string `json:"status"`
	NoticeDate    string `json:"notice_date"`
	EffectiveDate string `json:"effective_date"`
	FinalPayDate  string `json:"final_pay_date,omitempty"`
}

// ListOffboarding returns all offboarding records
func (c *Client) ListOffboarding(ctx context.Context) ([]Offboarding, error) {
	resp, err := c.Get(ctx, "/rest/v2/offboarding")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Offboarding `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// GetTerminationDetails returns termination details by ID
func (c *Client) GetTerminationDetails(ctx context.Context, id string) (*TerminationDetails, error) {
	path := fmt.Sprintf("/rest/v2/terminations/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data TerminationDetails `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
