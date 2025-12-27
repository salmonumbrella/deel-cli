package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// EORWorker represents an employee managed through Deel's EOR service
type EORWorker struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Country     string `json:"country"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	ContractID  string `json:"contract_id,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// EORWorkerBenefit represents a benefit provided to an EOR worker
type EORWorkerBenefit struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Status      string  `json:"status"`
}

// EORTaxDocument represents a tax document for an EOR worker
type EORTaxDocument struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Year        int    `json:"year"`
	Status      string `json:"status"`
	DownloadURL string `json:"download_url,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// EORBankAccount represents a bank account for an EOR worker
type EORBankAccount struct {
	ID            string `json:"id"`
	AccountHolder string `json:"account_holder"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number,omitempty"`
	IBAN          string `json:"iban,omitempty"`
	Swift         string `json:"swift,omitempty"`
	Currency      string `json:"currency"`
	IsPrimary     bool   `json:"is_primary"`
}

// CreateEORWorkerParams are parameters for creating an EOR worker
type CreateEORWorkerParams struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Country     string `json:"country"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// UpdateEORWorkerParams are parameters for updating an EOR worker
type UpdateEORWorkerParams struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
}

// AddBankAccountParams are parameters for adding a bank account to an EOR worker
type AddBankAccountParams struct {
	AccountHolder string `json:"account_holder"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number,omitempty"`
	IBAN          string `json:"iban,omitempty"`
	Swift         string `json:"swift,omitempty"`
	Currency      string `json:"currency"`
	IsPrimary     bool   `json:"is_primary"`
}

// CreateEORWorker creates a new EOR worker
func (c *Client) CreateEORWorker(ctx context.Context, params CreateEORWorkerParams) (*EORWorker, error) {
	resp, err := c.Post(ctx, "/rest/v2/eor/workers", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORWorker `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateEORWorker updates an existing EOR worker
func (c *Client) UpdateEORWorker(ctx context.Context, id string, params UpdateEORWorkerParams) (*EORWorker, error) {
	path := fmt.Sprintf("/rest/v2/eor/workers/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORWorker `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// GetEORWorkerBenefits returns all benefits for an EOR worker
func (c *Client) GetEORWorkerBenefits(ctx context.Context, workerID string) ([]EORWorkerBenefit, error) {
	path := fmt.Sprintf("/rest/v2/eor/workers/%s/benefits", escapePath(workerID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []EORWorkerBenefit `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// GetEORWorkerTaxDocuments returns all tax documents for an EOR worker
func (c *Client) GetEORWorkerTaxDocuments(ctx context.Context, workerID string) ([]EORTaxDocument, error) {
	path := fmt.Sprintf("/rest/v2/eor/workers/%s/tax-documents", escapePath(workerID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []EORTaxDocument `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// AddEORWorkerBankAccount adds a bank account to an EOR worker
func (c *Client) AddEORWorkerBankAccount(ctx context.Context, workerID string, params AddBankAccountParams) (*EORBankAccount, error) {
	path := fmt.Sprintf("/rest/v2/eor/workers/%s/bank-account", escapePath(workerID))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORBankAccount `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
