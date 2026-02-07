package api

import (
	"context"
	"fmt"
)

// GPContract represents a Global Payroll contract
type GPContract struct {
	ID           string  `json:"id"`
	WorkerID     string  `json:"worker_id"`
	WorkerName   string  `json:"worker_name"`
	WorkerEmail  string  `json:"worker_email"`
	Country      string  `json:"country"`
	StartDate    string  `json:"start_date"`
	Status       string  `json:"status"`
	JobTitle     string  `json:"job_title"`
	Department   string  `json:"department,omitempty"`
	Salary       float64 `json:"salary"`
	Currency     string  `json:"currency"`
	PayFrequency string  `json:"pay_frequency"`
	CreatedAt    string  `json:"created_at"`
}

// GPWorker represents a worker in the Global Payroll system
type GPWorker struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Country     string `json:"country"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	TaxID       string `json:"tax_id,omitempty"`
	Status      string `json:"status"`
}

// GPCompensation represents compensation details for a GP worker
type GPCompensation struct {
	WorkerID      string        `json:"worker_id"`
	Salary        float64       `json:"salary"`
	Currency      string        `json:"currency"`
	PayFrequency  string        `json:"pay_frequency"`
	EffectiveDate string        `json:"effective_date"`
	Allowances    []GPAllowance `json:"allowances"`
}

// GPAllowance represents an allowance in a GP compensation package
type GPAllowance struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// CreateGPContractParams are parameters for creating a GP contract
type CreateGPContractParams struct {
	WorkerEmail  string  `json:"worker_email"`
	WorkerName   string  `json:"worker_name"`
	Country      string  `json:"country"`
	StartDate    string  `json:"start_date"`
	JobTitle     string  `json:"job_title"`
	Department   string  `json:"department,omitempty"`
	Salary       float64 `json:"salary"`
	Currency     string  `json:"currency"`
	PayFrequency string  `json:"pay_frequency"`
}

// UpdateGPWorkerParams are parameters for updating a GP worker
type UpdateGPWorkerParams struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Address   string `json:"address,omitempty"`
	TaxID     string `json:"tax_id,omitempty"`
}

// UpdateGPCompensationParams are parameters for updating GP worker compensation
type UpdateGPCompensationParams struct {
	Salary        float64       `json:"salary,omitempty"`
	Currency      string        `json:"currency,omitempty"`
	PayFrequency  string        `json:"pay_frequency,omitempty"`
	EffectiveDate string        `json:"effective_date"`
	Allowances    []GPAllowance `json:"allowances,omitempty"`
}

// CreateGPContract creates a new Global Payroll contract
func (c *Client) CreateGPContract(ctx context.Context, params CreateGPContractParams) (*GPContract, error) {
	resp, err := c.Post(ctx, "/rest/v2/gp/contracts", params)
	if err != nil {
		return nil, err
	}

	return decodeData[GPContract](resp)
}

// UpdateGPWorker updates a Global Payroll worker's information
func (c *Client) UpdateGPWorker(ctx context.Context, workerID string, params UpdateGPWorkerParams) (*GPWorker, error) {
	path := fmt.Sprintf("/rest/v2/gp/workers/%s", escapePath(workerID))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[GPWorker](resp)
}

// UpdateGPCompensation updates a Global Payroll worker's compensation
func (c *Client) UpdateGPCompensation(ctx context.Context, workerID string, params UpdateGPCompensationParams) (*GPCompensation, error) {
	path := fmt.Sprintf("/rest/v2/gp/workers/%s/compensation", escapePath(workerID))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[GPCompensation](resp)
}
