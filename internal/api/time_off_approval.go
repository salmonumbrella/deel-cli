package api

import (
	"context"
	"fmt"
)

// TimeOffApproval represents the result of approving/rejecting a time off request
type TimeOffApproval struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"` // "approved" or "rejected"
	Comment   string `json:"comment,omitempty"`
}

// TimeOffValidation represents the validation result for a time off request
type TimeOffValidation struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// WorkSchedule represents a profile's work schedule
type WorkSchedule struct {
	ProfileID   string   `json:"profile_id"`
	WorkDays    []string `json:"work_days"` // ["monday", "tuesday", ...]
	HoursPerDay float64  `json:"hours_per_day"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Timezone    string   `json:"timezone"`
}

// Entitlement represents a time off entitlement
type Entitlement struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"` // "vacation", "sick", etc
	TotalDays   float64 `json:"total_days"`
	UsedDays    float64 `json:"used_days"`
	PendingDays float64 `json:"pending_days"`
	Balance     float64 `json:"balance"`
	Year        int     `json:"year"`
}

// ApproveRejectParams are parameters for approving or rejecting a time off request
type ApproveRejectParams struct {
	RequestID string `json:"request_id"`
	Action    string `json:"action"` // "approve" or "reject"
	Comment   string `json:"comment,omitempty"`
}

// ValidateTimeOffParams are parameters for validating a time off request
type ValidateTimeOffParams struct {
	ProfileID string `json:"profile_id"`
	Type      string `json:"type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// SyncTimeOffParams are parameters for syncing external time off
type SyncTimeOffParams struct {
	ProfileID  string `json:"profile_id"`
	ExternalID string `json:"external_id"`
	Type       string `json:"type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Status     string `json:"status"`
}

// ApproveRejectTimeOff approves or rejects a time off request
func (c *Client) ApproveRejectTimeOff(ctx context.Context, params ApproveRejectParams) (*TimeOffApproval, error) {
	resp, err := c.Post(ctx, "/rest/v2/time-off-requests/approve-reject", params)
	if err != nil {
		return nil, err
	}

	return decodeData[TimeOffApproval](resp)
}

// ValidateTimeOffRequest validates a time off request before creation
func (c *Client) ValidateTimeOffRequest(ctx context.Context, params ValidateTimeOffParams) (*TimeOffValidation, error) {
	resp, err := c.Post(ctx, "/rest/v2/time-off-requests/validate", params)
	if err != nil {
		return nil, err
	}

	return decodeData[TimeOffValidation](resp)
}

// GetWorkSchedule retrieves the work schedule for a profile
func (c *Client) GetWorkSchedule(ctx context.Context, profileID string) (*WorkSchedule, error) {
	path := fmt.Sprintf("/rest/v2/profiles/%s/work-schedule", escapePath(profileID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[WorkSchedule](resp)
}

// GetEntitlements retrieves time off entitlements for a profile
func (c *Client) GetEntitlements(ctx context.Context, profileID string) ([]Entitlement, error) {
	path := fmt.Sprintf("/rest/v2/profiles/%s/entitlements", escapePath(profileID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	entitlements, err := decodeData[[]Entitlement](resp)
	if err != nil {
		return nil, err
	}
	return *entitlements, nil
}

// SyncExternalTimeOff syncs time off from an external system
func (c *Client) SyncExternalTimeOff(ctx context.Context, params SyncTimeOffParams) error {
	_, err := c.Post(ctx, "/rest/v2/time-off/sync", params)
	return err
}
