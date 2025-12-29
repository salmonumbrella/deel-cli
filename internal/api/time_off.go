package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// TimeOffRequest represents a time off request
type TimeOffRequest struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Type       string  `json:"type"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	Days       float64 `json:"days"`
	Reason     string  `json:"reason"`
	WorkerName string  `json:"worker_name"`
	PolicyName string  `json:"policy_name"`
}

// TimeOffListParams are params for listing time off
type TimeOffListParams struct {
	HRISProfileID string
	Status        []string
	Limit         int
	Cursor        string
}

// TimeOffListResponse is the response from list time off
type TimeOffListResponse struct {
	Data []TimeOffRequest `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"page"`
}

// ListTimeOffRequests returns time off requests
func (c *Client) ListTimeOffRequests(ctx context.Context, params TimeOffListParams) (*TimeOffListResponse, error) {
	q := url.Values{}
	if params.HRISProfileID != "" {
		q.Set("hris_profile_id", params.HRISProfileID)
	}
	for _, s := range params.Status {
		q.Add("status", s)
	}
	if params.Limit > 0 {
		q.Set("page_size", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("next", params.Cursor)
	}

	path := "/rest/v2/time-off"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result TimeOffListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// TimeOffPolicy represents a time off policy
type TimeOffPolicy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ListTimeOffPolicies returns time off policies
func (c *Client) ListTimeOffPolicies(ctx context.Context) ([]TimeOffPolicy, error) {
	resp, err := c.Get(ctx, "/rest/v2/time-off/policies")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []TimeOffPolicy `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// CreateTimeOffRequest creates a new time off request
type CreateTimeOffParams struct {
	HRISProfileID string `json:"hris_profile_id"`
	PolicyID      string `json:"policy_id"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Reason        string `json:"reason,omitempty"`
}

// CreateTimeOffRequest creates a time off request
func (c *Client) CreateTimeOffRequest(ctx context.Context, params CreateTimeOffParams) (*TimeOffRequest, error) {
	resp, err := c.Post(ctx, "/rest/v2/time-off", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data TimeOffRequest `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// CancelTimeOffRequest cancels a time off request
func (c *Client) CancelTimeOffRequest(ctx context.Context, requestID string) error {
	path := fmt.Sprintf("/rest/v2/time-off/%s/cancel", escapePath(requestID))
	_, err := c.Post(ctx, path, nil)
	return err
}
