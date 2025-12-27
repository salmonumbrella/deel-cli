package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// OnboardingEmployee represents an employee in onboarding
type OnboardingEmployee struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Country    string `json:"country"`
	Status     string `json:"status"`
	Stage      string `json:"stage"`
	StartDate  string `json:"start_date"`
	ContractID string `json:"contract_id"`
	Progress   int    `json:"progress_percent"`
}

// OnboardingListResponse is the response from list onboarding
type OnboardingListResponse struct {
	Data []OnboardingEmployee `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"page"`
}

// OnboardingListParams are params for listing onboarding
type OnboardingListParams struct {
	Status string
	Limit  int
	Cursor string
}

// ListOnboardingEmployees returns employees in onboarding
func (c *Client) ListOnboardingEmployees(ctx context.Context, params OnboardingListParams) (*OnboardingListResponse, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/onboarding"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result OnboardingListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// OnboardingDetails represents detailed onboarding info
type OnboardingDetails struct {
	ID             string   `json:"id"`
	EmployeeName   string   `json:"employee_name"`
	Status         string   `json:"status"`
	Stage          string   `json:"stage"`
	Progress       int      `json:"progress_percent"`
	PendingTasks   []string `json:"pending_tasks"`
	CompletedTasks []string `json:"completed_tasks"`
	StartDate      string   `json:"start_date"`
	EstimatedEnd   string   `json:"estimated_end_date"`
}

// GetOnboardingDetails returns onboarding details for an employee
func (c *Client) GetOnboardingDetails(ctx context.Context, employeeID string) (*OnboardingDetails, error) {
	path := fmt.Sprintf("/rest/v2/onboarding/%s", escapePath(employeeID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data OnboardingDetails `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
