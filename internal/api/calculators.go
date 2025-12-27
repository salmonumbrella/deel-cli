package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// CostCalculation represents a cost calculation result
type CostCalculation struct {
	Country      string  `json:"country"`
	GrossSalary  float64 `json:"gross_salary"`
	EmployerCost float64 `json:"employer_cost"`
	TotalCost    float64 `json:"total_cost"`
	Currency     string  `json:"currency"`
	TaxesAndFees float64 `json:"taxes_and_fees"`
}

// CalculateCostParams are params for cost calculation
type CalculateCostParams struct {
	Country     string  `json:"country"`
	GrossSalary float64 `json:"gross_salary"`
	Currency    string  `json:"currency"`
}

// CalculateCost calculates employer cost
func (c *Client) CalculateCost(ctx context.Context, params CalculateCostParams) (*CostCalculation, error) {
	resp, err := c.Post(ctx, "/rest/v2/calculators/cost", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data CostCalculation `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// TakeHomeCalculation represents a take-home calculation result
type TakeHomeCalculation struct {
	Country     string  `json:"country"`
	GrossSalary float64 `json:"gross_salary"`
	NetSalary   float64 `json:"net_salary"`
	TaxRate     float64 `json:"tax_rate"`
	Currency    string  `json:"currency"`
}

// CalculateTakeHomeParams are params for take-home calculation
type CalculateTakeHomeParams struct {
	Country     string  `json:"country"`
	GrossSalary float64 `json:"gross_salary"`
	Currency    string  `json:"currency"`
}

// CalculateTakeHome calculates net salary
func (c *Client) CalculateTakeHome(ctx context.Context, params CalculateTakeHomeParams) (*TakeHomeCalculation, error) {
	resp, err := c.Post(ctx, "/rest/v2/calculators/take-home", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data TakeHomeCalculation `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// SalaryHistogram represents salary data for a role
type SalaryHistogram struct {
	Role         string  `json:"role"`
	Country      string  `json:"country"`
	Currency     string  `json:"currency"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Median       float64 `json:"median"`
	Percentile25 float64 `json:"percentile_25"`
	Percentile75 float64 `json:"percentile_75"`
}

// GetSalaryHistogram returns salary data
func (c *Client) GetSalaryHistogram(ctx context.Context, role, country string) (*SalaryHistogram, error) {
	path := fmt.Sprintf("/rest/v2/calculators/salary-histogram?role=%s&country=%s", escapePath(role), escapePath(country))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data SalaryHistogram `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
