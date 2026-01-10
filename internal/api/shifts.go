package api

import (
	"context"
	"fmt"
	"net/url"
)

// Shift represents a worker shift
type Shift struct {
	ID         string  `json:"id"`
	WorkerID   string  `json:"worker_id"`
	WorkerName string  `json:"worker_name"`
	Date       string  `json:"date"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Hours      float64 `json:"hours"`
	Status     string  `json:"status"`
}

// ShiftsListParams are params for listing shifts
type ShiftsListParams struct {
	WorkerID  string
	StartDate string
	EndDate   string
	Limit     int
}

// ListShifts returns shifts
func (c *Client) ListShifts(ctx context.Context, params ShiftsListParams) ([]Shift, error) {
	q := url.Values{}
	if params.WorkerID != "" {
		q.Set("worker_id", params.WorkerID)
	}
	if params.StartDate != "" {
		q.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		q.Set("end_date", params.EndDate)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}

	path := "/rest/v2/shifts"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	shifts, err := decodeData[[]Shift](resp)
	if err != nil {
		return nil, err
	}
	return *shifts, nil
}

// ShiftRate represents shift rates
type ShiftRate struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Country  string  `json:"country"`
	Rate     float64 `json:"rate"`
	Currency string  `json:"currency"`
	Type     string  `json:"type"`
}

// ListShiftRates returns shift rates
func (c *Client) ListShiftRates(ctx context.Context, country string) ([]ShiftRate, error) {
	path := fmt.Sprintf("/rest/v2/shifts/rates?country=%s", escapePath(country))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	rates, err := decodeData[[]ShiftRate](resp)
	if err != nil {
		return nil, err
	}
	return *rates, nil
}
