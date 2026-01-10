package api

import (
	"context"
	"fmt"
)

// GPBankAccount represents a bank account for a Global Payroll worker
type GPBankAccount struct {
	ID            string `json:"id"`
	WorkerID      string `json:"worker_id"`
	AccountHolder string `json:"account_holder"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number,omitempty"`
	IBAN          string `json:"iban,omitempty"`
	Swift         string `json:"swift,omitempty"`
	Currency      string `json:"currency"`
	IsPrimary     bool   `json:"is_primary"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// GPBankGuide provides country-specific banking requirements
type GPBankGuide struct {
	Country         string            `json:"country"`
	RequiredFields  []string          `json:"required_fields"`
	OptionalFields  []string          `json:"optional_fields"`
	SupportedBanks  []string          `json:"supported_banks,omitempty"`
	ValidationRules map[string]string `json:"validation_rules,omitempty"`
}

// AddGPBankAccountParams are parameters for adding a bank account
type AddGPBankAccountParams struct {
	WorkerID      string `json:"worker_id"`
	AccountHolder string `json:"account_holder"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number,omitempty"`
	IBAN          string `json:"iban,omitempty"`
	Swift         string `json:"swift,omitempty"`
	Currency      string `json:"currency"`
	IsPrimary     bool   `json:"is_primary"`
}

// UpdateGPBankAccountParams are parameters for updating a bank account
type UpdateGPBankAccountParams struct {
	AccountHolder string `json:"account_holder,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	AccountNumber string `json:"account_number,omitempty"`
	IsPrimary     bool   `json:"is_primary,omitempty"`
}

// AddGPBankAccount adds a new bank account for a Global Payroll worker
func (c *Client) AddGPBankAccount(ctx context.Context, params AddGPBankAccountParams) (*GPBankAccount, error) {
	resp, err := c.Post(ctx, "/rest/v2/gp/bank-accounts", params)
	if err != nil {
		return nil, err
	}

	return decodeData[GPBankAccount](resp)
}

// ListGPBankAccounts retrieves all bank accounts for a specific worker
func (c *Client) ListGPBankAccounts(ctx context.Context, workerID string) ([]GPBankAccount, error) {
	path := fmt.Sprintf("/rest/v2/gp/bank-accounts?worker_id=%s", escapePath(workerID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	accounts, err := decodeData[[]GPBankAccount](resp)
	if err != nil {
		return nil, err
	}
	return *accounts, nil
}

// UpdateGPBankAccount updates a Global Payroll worker's bank account
func (c *Client) UpdateGPBankAccount(ctx context.Context, accountID string, params UpdateGPBankAccountParams) (*GPBankAccount, error) {
	path := fmt.Sprintf("/rest/v2/gp/bank-accounts/%s", escapePath(accountID))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[GPBankAccount](resp)
}

// GetGPBankGuide retrieves banking requirements and validation rules for a specific country
func (c *Client) GetGPBankGuide(ctx context.Context, country string) (*GPBankGuide, error) {
	path := fmt.Sprintf("/rest/v2/gp/bank-guide?country=%s", escapePath(country))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[GPBankGuide](resp)
}
