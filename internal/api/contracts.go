package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Contract represents a Deel contract
type Contract struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Type               string  `json:"type"`
	Status             string  `json:"status"`
	WorkerName         string  `json:"worker_name"`
	WorkerEmail        string  `json:"worker_email"`
	StartDate          string  `json:"start_date"`
	EndDate            string  `json:"end_date"`
	Currency           string  `json:"currency"`
	CompensationAmount float64 `json:"compensation_amount"`
	Country            string  `json:"country"`
}

// ContractsListResponse is the response from list contracts
type ContractsListResponse struct {
	Data []Contract `json:"data"`
	Page struct {
		Next  string `json:"next"`
		Total int    `json:"total"`
	} `json:"page"`
}

// ContractsListParams are params for listing contracts
type ContractsListParams struct {
	Limit  int
	Cursor string
	Status string
	Type   string
}

// ListContracts returns a list of contracts
func (c *Client) ListContracts(ctx context.Context, params ContractsListParams) (*ContractsListResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Type != "" {
		q.Set("type", params.Type)
	}

	path := "/rest/v2/contracts"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result ContractsListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetContract returns a single contract
func (c *Client) GetContract(ctx context.Context, contractID string) (*Contract, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Contract `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// ContractAmendment represents a contract amendment
type ContractAmendment struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// ListContractAmendments returns amendments for a contract
func (c *Client) ListContractAmendments(ctx context.Context, contractID string) ([]ContractAmendment, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/amendments", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []ContractAmendment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// PaymentDate represents a contract payment date
type PaymentDate struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
}

// GetContractPaymentDates returns payment dates for a contract
func (c *Client) GetContractPaymentDates(ctx context.Context, contractID string) ([]PaymentDate, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/payment-dates", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []PaymentDate `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// CreateContractParams are params for creating a contract
type CreateContractParams struct {
	Title          string  `json:"title"`
	Type           string  `json:"type"` // fixed_rate, pay_as_you_go, milestone, task_based
	WorkerEmail    string  `json:"worker_email"`
	Currency       string  `json:"currency"`
	Rate           float64 `json:"rate,omitempty"`
	Country        string  `json:"country"`
	JobTitle       string  `json:"job_title,omitempty"`
	ScopeOfWork    string  `json:"scope_of_work,omitempty"`
	StartDate      string  `json:"start_date,omitempty"`
	EndDate        string  `json:"end_date,omitempty"`
	PaymentCycle   string  `json:"payment_cycle,omitempty"` // weekly, bi_weekly, monthly
	SeniorityLevel string  `json:"seniority_level,omitempty"`
}

// CreateContract creates a new contractor contract
func (c *Client) CreateContract(ctx context.Context, params CreateContractParams) (*Contract, error) {
	resp, err := c.Post(ctx, "/rest/v2/contracts", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Contract `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// SignContract signs a contract (as the client/employer)
func (c *Client) SignContract(ctx context.Context, contractID string) (*Contract, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/sign", escapePath(contractID))
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Contract `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// TerminateContractParams are params for terminating a contract
type TerminateContractParams struct {
	Reason        string `json:"reason"`
	EffectiveDate string `json:"effective_date,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

// TerminateContract initiates contract termination
func (c *Client) TerminateContract(ctx context.Context, contractID string, params TerminateContractParams) error {
	path := fmt.Sprintf("/rest/v2/contracts/%s/terminations", escapePath(contractID))
	_, err := c.Post(ctx, path, params)
	return err
}

// ListTerminationReasons returns available termination reasons
func (c *Client) ListTerminationReasons(ctx context.Context, contractID string) ([]string, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/terminations/reasons", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []struct {
			Value string `json:"value"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	reasons := make([]string, len(wrapper.Data))
	for i, r := range wrapper.Data {
		reasons[i] = r.Value
	}
	return reasons, nil
}

// GetContractPDF returns the download URL for a contract PDF
func (c *Client) GetContractPDF(ctx context.Context, contractID string) (string, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/pdf", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return "", err
	}

	var wrapper struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data.URL, nil
}

// InviteWorker sends an invitation email to the worker
func (c *Client) InviteWorker(ctx context.Context, contractID string) error {
	path := fmt.Sprintf("/rest/v2/contracts/%s/invite", escapePath(contractID))
	_, err := c.Post(ctx, path, nil)
	return err
}

// GetInviteLink returns the invite link for a contract
func (c *Client) GetInviteLink(ctx context.Context, contractID string) (string, error) {
	path := fmt.Sprintf("/rest/v2/contracts/%s/invite-link", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return "", err
	}

	var wrapper struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data.URL, nil
}

// RemoveInvite removes a pending invitation
func (c *Client) RemoveInvite(ctx context.Context, contractID string) error {
	path := fmt.Sprintf("/rest/v2/contracts/%s/invite", escapePath(contractID))
	_, err := c.Delete(ctx, path)
	return err
}

// ContractTemplate represents a contract template
type ContractTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ListContractTemplates returns available contract templates
func (c *Client) ListContractTemplates(ctx context.Context) ([]ContractTemplate, error) {
	resp, err := c.Get(ctx, "/rest/v2/contract-templates")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []ContractTemplate `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}
