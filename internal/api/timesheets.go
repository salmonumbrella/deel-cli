package api

import (
	"context"
	"fmt"
	"net/url"
)

// Timesheet represents a timesheet for a contract period
type Timesheet struct {
	ID          string           `json:"id"`
	ContractID  string           `json:"contract_id"`
	Status      string           `json:"status"`
	PeriodStart string           `json:"period_start"`
	PeriodEnd   string           `json:"period_end"`
	TotalHours  float64          `json:"total_hours"`
	Entries     []TimesheetEntry `json:"entries"`
	CreatedAt   string           `json:"created_at"`
}

// TimesheetEntry represents a single timesheet entry
type TimesheetEntry struct {
	ID          string  `json:"id"`
	TimesheetID string  `json:"timesheet_id"`
	Date        string  `json:"date"`
	Hours       float64 `json:"hours"`
	Description string  `json:"description"`
}

// TimesheetsListParams are parameters for listing timesheets
type TimesheetsListParams struct {
	ContractID string
	Status     string
	Cursor     string
	Limit      int
}

// TimesheetsListResponse is the response from list timesheets
type TimesheetsListResponse = ListResponse[Timesheet]

// CreateTimesheetEntryParams are parameters for creating a timesheet entry
type CreateTimesheetEntryParams struct {
	TimesheetID string  `json:"timesheet_id"`
	Date        string  `json:"date"`
	Hours       float64 `json:"hours"`
	Description string  `json:"description,omitempty"`
}

// UpdateTimesheetEntryParams are parameters for updating a timesheet entry
type UpdateTimesheetEntryParams struct {
	Hours       float64 `json:"hours,omitempty"`
	Description string  `json:"description,omitempty"`
}

// ReviewTimesheetParams are parameters for reviewing a timesheet
type ReviewTimesheetParams struct {
	Status  string `json:"status"` // "approved" or "rejected"
	Comment string `json:"comment,omitempty"`
}

// ListTimesheets returns timesheets for a contract
func (c *Client) ListTimesheets(ctx context.Context, params TimesheetsListParams) (*TimesheetsListResponse, error) {
	q := url.Values{}
	if params.ContractID != "" {
		q.Set("contract_id", params.ContractID)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}

	path := "/rest/v2/timesheets"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[Timesheet](resp)
}

// GetTimesheet returns a single timesheet
func (c *Client) GetTimesheet(ctx context.Context, id string) (*Timesheet, error) {
	path := fmt.Sprintf("/rest/v2/timesheets/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[Timesheet](resp)
}

// CreateTimesheetEntry creates a new timesheet entry
func (c *Client) CreateTimesheetEntry(ctx context.Context, params CreateTimesheetEntryParams) (*TimesheetEntry, error) {
	resp, err := c.Post(ctx, "/rest/v2/timesheet-entries", params)
	if err != nil {
		return nil, err
	}

	return decodeData[TimesheetEntry](resp)
}

// UpdateTimesheetEntry updates an existing timesheet entry
func (c *Client) UpdateTimesheetEntry(ctx context.Context, id string, params UpdateTimesheetEntryParams) (*TimesheetEntry, error) {
	path := fmt.Sprintf("/rest/v2/timesheet-entries/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[TimesheetEntry](resp)
}

// DeleteTimesheetEntry deletes a timesheet entry
func (c *Client) DeleteTimesheetEntry(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/v2/timesheet-entries/%s", escapePath(id))
	_, err := c.Delete(ctx, path)
	return err
}

// ReviewTimesheet submits a review for a timesheet
func (c *Client) ReviewTimesheet(ctx context.Context, id string, params ReviewTimesheetParams) (*Timesheet, error) {
	path := fmt.Sprintf("/rest/v2/timesheets/%s/review", escapePath(id))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[Timesheet](resp)
}
