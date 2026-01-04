package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Contract represents a Deel contract
type Contract struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Type               string  `json:"type"`
	Status             string  `json:"status"`
	WorkerName         string  `json:"worker_name"`
	WorkerEmail        string  `json:"worker_email"`
	Entity             string  `json:"entity"`
	StartDate          string  `json:"start_date"`
	EndDate            string  `json:"end_date"`
	Currency           string  `json:"currency"`
	CompensationAmount float64 `json:"compensation_amount"`
	Country            string  `json:"country"`
}

// rawContract is used for parsing the nested API response
type rawContract struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"termination_date"`
	Client struct {
		LegalEntity struct {
			Name string `json:"name"`
		} `json:"legal_entity"`
	} `json:"client"`
	Worker struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Country  string `json:"country"`
	} `json:"worker"`
	CompensationDetails struct {
		CurrencyCode string `json:"currency_code"`
		Amount       string `json:"amount"`
	} `json:"compensation_details"`
}

// UnmarshalJSON implements custom unmarshaling to handle nested API response
func (c *Contract) UnmarshalJSON(data []byte) error {
	var raw rawContract
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.ID = raw.ID
	c.Title = raw.Title
	c.Type = raw.Type
	c.Status = raw.Status
	c.StartDate = raw.StartDate
	c.EndDate = raw.EndDate
	c.WorkerName = raw.Worker.FullName
	c.WorkerEmail = raw.Worker.Email
	c.Entity = raw.Client.LegalEntity.Name
	c.Country = raw.Worker.Country
	c.Currency = raw.CompensationDetails.CurrencyCode

	if raw.CompensationDetails.Amount != "" {
		if amount, err := strconv.ParseFloat(raw.CompensationDetails.Amount, 64); err == nil {
			c.CompensationAmount = amount
		}
	}

	return nil
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
	Type           string  `json:"type"` // fixed_rate, pay_as_you_go, pay_as_you_go_time_based, milestone, task_based
	WorkerEmail    string  `json:"worker_email"`
	Currency       string  `json:"currency"`
	Rate           float64 `json:"rate,omitempty"`
	Country        string  `json:"country"`
	JobTitle       string  `json:"job_title,omitempty"`
	ScopeOfWork    string  `json:"scope_of_work,omitempty"`
	StartDate      string  `json:"start_date,omitempty"`
	EndDate        string  `json:"end_date,omitempty"`
	PaymentCycle   string  `json:"payment_cycle,omitempty"` // weekly, bi_weekly, monthly (legacy)
	SeniorityLevel string  `json:"seniority_level,omitempty"`

	// Extended fields for pay-as-you-go contracts
	TemplateID    string `json:"-"` // Handled specially in request
	LegalEntityID string `json:"-"` // Handled specially in request
	GroupID       string `json:"-"` // Handled specially in request
	CycleEnd      int    `json:"-"` // Day of month/week for payment cycle
	CycleEndType  string `json:"-"` // DAY_OF_MONTH, DAY_OF_WEEK, DAY_OF_LAST_WEEK
	Frequency     string `json:"-"` // monthly, weekly, biweekly, semimonthly
}

// createContractRequest is the API request body structure
type createContractRequest struct {
	Title               string                `json:"title"`
	Type                string                `json:"type"`
	WorkerEmail         string                `json:"worker_email,omitempty"`
	Currency            string                `json:"currency,omitempty"`
	Country             string                `json:"country,omitempty"`
	JobTitle            string                `json:"job_title,omitempty"`
	ScopeOfWork         string                `json:"scope_of_work,omitempty"`
	StartDate           string                `json:"start_date,omitempty"`
	EndDate             string                `json:"end_date,omitempty"`
	SeniorityLevel      string                `json:"seniority_level,omitempty"`
	PaymentCycle        string                `json:"payment_cycle,omitempty"`
	ContractTemplateID  string                `json:"contract_template_id,omitempty"`
	Client              *createContractClient `json:"client,omitempty"`
	CompensationDetails *compensationDetails  `json:"compensation_details,omitempty"`
}

type createContractClient struct {
	LegalEntity *entityRef `json:"legal_entity,omitempty"`
	Team        *entityRef `json:"team,omitempty"`
}

type entityRef struct {
	ID string `json:"id"`
}

type compensationDetails struct {
	Amount       float64 `json:"amount,omitempty"`
	CurrencyCode string  `json:"currency_code,omitempty"`
	CycleEnd     int     `json:"cycle_end,omitempty"`
	CycleEndType string  `json:"cycle_end_type,omitempty"`
	Frequency    string  `json:"frequency,omitempty"`
}

// CreateContract creates a new contractor contract
func (c *Client) CreateContract(ctx context.Context, params CreateContractParams) (*Contract, error) {
	req := createContractRequest{
		Title:          params.Title,
		Type:           params.Type,
		WorkerEmail:    params.WorkerEmail,
		Country:        params.Country,
		JobTitle:       params.JobTitle,
		ScopeOfWork:    params.ScopeOfWork,
		StartDate:      params.StartDate,
		EndDate:        params.EndDate,
		SeniorityLevel: params.SeniorityLevel,
		PaymentCycle:   params.PaymentCycle,
	}

	// Add template if specified
	if params.TemplateID != "" {
		req.ContractTemplateID = params.TemplateID
	}

	// Add client structure if legal entity or group specified
	if params.LegalEntityID != "" || params.GroupID != "" {
		req.Client = &createContractClient{}
		if params.LegalEntityID != "" {
			req.Client.LegalEntity = &entityRef{ID: params.LegalEntityID}
		}
		if params.GroupID != "" {
			req.Client.Team = &entityRef{ID: params.GroupID}
		}
	}

	// Add compensation details
	if params.Rate > 0 || params.Currency != "" || params.CycleEnd > 0 || params.Frequency != "" {
		cycleEndType := ""
		if params.CycleEnd > 0 {
			cycleEndType = params.CycleEndType
		}
		req.CompensationDetails = &compensationDetails{
			Amount:       params.Rate,
			CurrencyCode: params.Currency,
			CycleEnd:     params.CycleEnd,
			CycleEndType: cycleEndType,
			Frequency:    params.Frequency,
		}
	}

	resp, err := c.Post(ctx, "/rest/v2/contracts", req)
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

// TerminateContractParams are params for terminating a contractor contract
type TerminateContractParams struct {
	TerminateNow                 bool   `json:"terminate_now,omitempty"`
	CompletionDate               string `json:"completion_date,omitempty"`                 // Required if terminate_now is false (YYYY-MM-DD)
	TerminationType              string `json:"termination_type,omitempty"`                // RESIGNATION, TERMINATION, END_OF_CONTRACT
	TerminationReasonID          string `json:"termination_reason_id,omitempty"`           // UUID from termination reasons endpoint
	TerminationReasonDescription string `json:"termination_reason_description,omitempty"`  // Free text description
	EligibleForRehire            string `json:"eligible_for_rehire,omitempty"`             // YES, NO, DONT_KNOW
	Message                      string `json:"message,omitempty"`                         // Optional message (max 1000 chars)
}

// terminateContractRequest wraps params in data object as required by API
type terminateContractRequest struct {
	Data TerminateContractParams `json:"data"`
}

// TerminateContract initiates contract termination for a contractor
func (c *Client) TerminateContract(ctx context.Context, contractID string, params TerminateContractParams) error {
	path := fmt.Sprintf("/rest/v2/contracts/%s/terminations", escapePath(contractID))
	req := terminateContractRequest{Data: params}
	_, err := c.Post(ctx, path, req)
	return err
}

// TerminationReason represents a termination reason from the API
type TerminationReason struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ListTerminationReasons returns available termination reasons (global list, not per-contract)
func (c *Client) ListTerminationReasons(ctx context.Context) ([]TerminationReason, error) {
	path := "/rest/v2/contracts/termination-reasons"
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []TerminationReason `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return wrapper.Data, nil
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
