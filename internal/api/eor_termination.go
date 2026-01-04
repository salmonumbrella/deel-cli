package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// EORTermination represents a contract termination (resignation or termination)
type EORTermination struct {
	ID               string  `json:"id"`
	ContractID       string  `json:"contract_id"`
	Type             string  `json:"type"` // "resignation" or "termination"
	Status           string  `json:"status"`
	Reason           string  `json:"reason"`
	EffectiveDate    string  `json:"effective_date"`
	LastWorkingDay   string  `json:"last_working_day"`
	NoticePeriodDays int     `json:"notice_period_days"`
	SeveranceAmount  float64 `json:"severance_amount,omitempty"`
	Currency         string  `json:"currency,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

// EORUsedTimeOff represents time off information for EOR termination
type EORUsedTimeOff struct {
	PaidTimeOff   int `json:"paid_time_off"`   // Days of paid time off used
	UnpaidTimeOff int `json:"unpaid_time_off"` // Days of unpaid time off used
	SickLeave     int `json:"sick_leave"`      // Days of sick leave used
}

// EORResignationParams are parameters for requesting an EOR resignation (employee-initiated)
type EORResignationParams struct {
	Reason                string `json:"reason"` // Enum: EMPLOYEE_IS_MOVING_TO_ANOTHER_COUNTRY, etc.
	IsEmployeeStayingDeel bool   `json:"is_employee_staying_with_deel"`
}

// EORTerminationParams are parameters for requesting an EOR termination (employer-initiated)
type EORTerminationParams struct {
	Reason             string         `json:"reason"`               // Enum: TERMINATION, FOR_CAUSE, PERFORMANCE, etc.
	ReasonDetail       string         `json:"reason_detail"`        // Required, 100-5000 chars
	IsEmployeeNotified bool           `json:"is_employee_notified"` // Has the employee been notified
	UsedTimeOff        EORUsedTimeOff `json:"used_time_off"`        // Required time off information
	IsSensitive        bool           `json:"is_sensitive,omitempty"`
	SeveranceType      string         `json:"severance_type,omitempty"` // DAYS, WEEKS, MONTHS, CASH
}

// eorResignationRequest wraps params in data object as required by API
type eorResignationRequest struct {
	Data EORResignationParams `json:"data"`
}

// eorTerminationRequest wraps params in data object as required by API
type eorTerminationRequest struct {
	Data EORTerminationParams `json:"data"`
}

// RequestEORResignation creates a resignation request for an EOR contract (employee-initiated)
func (c *Client) RequestEORResignation(ctx context.Context, contractOID string, params EORResignationParams) (*EORTermination, error) {
	path := fmt.Sprintf("/rest/v2/eor/%s/terminations/", escapePath(contractOID))
	req := eorResignationRequest{Data: params}
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORTermination `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// RequestEORTermination creates a termination request for an EOR contract (employer-initiated)
func (c *Client) RequestEORTermination(ctx context.Context, contractOID string, params EORTerminationParams) (*EORTermination, error) {
	path := fmt.Sprintf("/rest/v2/eor/%s/terminations/", escapePath(contractOID))
	req := eorTerminationRequest{Data: params}
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORTermination `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// GetEORTermination retrieves the termination details for an EOR contract
func (c *Client) GetEORTermination(ctx context.Context, contractOID string) (*EORTermination, error) {
	path := fmt.Sprintf("/rest/v2/eor/%s/terminations/", escapePath(contractOID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data EORTermination `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
