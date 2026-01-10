package api

import (
	"context"
	"fmt"
)

// EORContract represents an Employer of Record contract
type EORContract struct {
	ID             string       `json:"id"`
	Title          string       `json:"title"`
	Status         string       `json:"status"`
	WorkerID       string       `json:"worker_id"`
	WorkerEmail    string       `json:"worker_email"`
	WorkerName     string       `json:"worker_name"`
	Country        string       `json:"country"`
	StartDate      string       `json:"start_date"`
	EndDate        string       `json:"end_date,omitempty"`
	Salary         float64      `json:"salary"`
	Currency       string       `json:"currency"`
	PayFrequency   string       `json:"pay_frequency"`
	JobTitle       string       `json:"job_title"`
	SeniorityLevel string       `json:"seniority_level,omitempty"`
	Scope          string       `json:"scope,omitempty"`
	Benefits       []EORBenefit `json:"benefits,omitempty"`
	CreatedAt      string       `json:"created_at"`
}

// EORBenefit represents a benefit in an EOR contract
type EORBenefit struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
}

// CreateEORContractParams are parameters for creating an EOR contract
type CreateEORContractParams struct {
	Title          string  `json:"title"`
	WorkerEmail    string  `json:"worker_email"`
	WorkerName     string  `json:"worker_name"`
	Country        string  `json:"country"`
	StartDate      string  `json:"start_date"`
	Salary         float64 `json:"salary"`
	Currency       string  `json:"currency"`
	PayFrequency   string  `json:"pay_frequency"`
	JobTitle       string  `json:"job_title"`
	SeniorityLevel string  `json:"seniority_level,omitempty"`
	Scope          string  `json:"scope,omitempty"`
}

// UpdateEORContractParams are parameters for updating an EOR contract
type UpdateEORContractParams struct {
	Title          string  `json:"title,omitempty"`
	EndDate        string  `json:"end_date,omitempty"`
	Salary         float64 `json:"salary,omitempty"`
	JobTitle       string  `json:"job_title,omitempty"`
	SeniorityLevel string  `json:"seniority_level,omitempty"`
	Scope          string  `json:"scope,omitempty"`
}

// CancelEORContractParams are parameters for cancelling an EOR contract
type CancelEORContractParams struct {
	Reason string `json:"reason"`
}

// CreateEORContract creates a new EOR contract
func (c *Client) CreateEORContract(ctx context.Context, params CreateEORContractParams) (*EORContract, error) {
	resp, err := c.Post(ctx, "/rest/v2/eor/contracts", params)
	if err != nil {
		return nil, err
	}

	return decodeData[EORContract](resp)
}

// GetEORContract returns a single EOR contract by ID
func (c *Client) GetEORContract(ctx context.Context, id string) (*EORContract, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[EORContract](resp)
}

// UpdateEORContract updates an existing EOR contract
func (c *Client) UpdateEORContract(ctx context.Context, id string, params UpdateEORContractParams) (*EORContract, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[EORContract](resp)
}

// CancelEORContract cancels an EOR contract
func (c *Client) CancelEORContract(ctx context.Context, id string, params CancelEORContractParams) (*EORContract, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s/cancel", escapePath(id))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[EORContract](resp)
}

// SignEORContract signs an EOR contract
func (c *Client) SignEORContract(ctx context.Context, id string) (*EORContract, error) {
	path := fmt.Sprintf("/rest/v2/eor/contracts/%s/sign", escapePath(id))
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	return decodeData[EORContract](resp)
}
