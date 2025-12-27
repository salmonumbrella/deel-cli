package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// WithdrawFundsParams are params for withdrawing funds
type WithdrawFundsParams struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description,omitempty"`
}

// Withdrawal represents a payout withdrawal
type Withdrawal struct {
	ID          string  `json:"id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Description string  `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// WithdrawFunds creates a payout withdrawal
func (c *Client) WithdrawFunds(ctx context.Context, params WithdrawFundsParams) (*Withdrawal, error) {
	resp, err := c.Post(ctx, "/rest/v2/payouts/withdraw", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Withdrawal `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// AutoWithdrawalSettings represents auto-withdrawal configuration
type AutoWithdrawalSettings struct {
	Enabled   bool    `json:"enabled"`
	Threshold float64 `json:"threshold,omitempty"`
	Currency  string  `json:"currency,omitempty"`
	Schedule  string  `json:"schedule,omitempty"`
}

// GetAutoWithdrawal returns auto-withdrawal settings
func (c *Client) GetAutoWithdrawal(ctx context.Context) (*AutoWithdrawalSettings, error) {
	resp, err := c.Get(ctx, "/rest/v2/payouts/auto-withdrawal")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data AutoWithdrawalSettings `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// SetAutoWithdrawalParams are params for updating auto-withdrawal settings
type SetAutoWithdrawalParams struct {
	Enabled   bool    `json:"enabled"`
	Threshold float64 `json:"threshold,omitempty"`
	Currency  string  `json:"currency,omitempty"`
	Schedule  string  `json:"schedule,omitempty"`
}

// SetAutoWithdrawal updates auto-withdrawal settings
func (c *Client) SetAutoWithdrawal(ctx context.Context, params SetAutoWithdrawalParams) (*AutoWithdrawalSettings, error) {
	resp, err := c.Patch(ctx, "/rest/v2/payouts/auto-withdrawal", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data AutoWithdrawalSettings `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// ContractorBalance represents a contractor's balance
type ContractorBalance struct {
	ContractorID   string  `json:"contractor_id"`
	ContractorName string  `json:"contractor_name,omitempty"`
	Balance        float64 `json:"balance"`
	Currency       string  `json:"currency"`
	PendingAmount  float64 `json:"pending_amount,omitempty"`
	UpdatedAt      string  `json:"updated_at,omitempty"`
}

// ListContractorBalances returns balances for all contractors
func (c *Client) ListContractorBalances(ctx context.Context) ([]ContractorBalance, error) {
	resp, err := c.Get(ctx, "/rest/v2/contractors/balances")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []ContractorBalance `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}
