package api

import (
	"context"
	"fmt"
)

// Benefit represents an employee benefit
type Benefit struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Provider    string  `json:"provider"`
	Description string  `json:"description"`
	Cost        float64 `json:"cost"`
	Currency    string  `json:"currency"`
	Country     string  `json:"country"`
	Status      string  `json:"status"`
}

// ListBenefitsByCountry returns benefits available in a country
func (c *Client) ListBenefitsByCountry(ctx context.Context, countryCode string) ([]Benefit, error) {
	path := fmt.Sprintf("/rest/v2/benefits/countries/%s", escapePath(countryCode))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	benefits, err := decodeData[[]Benefit](resp)
	if err != nil {
		return nil, err
	}
	return *benefits, nil
}

// EmployeeBenefit represents a benefit assigned to an employee
type EmployeeBenefit struct {
	ID           string  `json:"id"`
	BenefitID    string  `json:"benefit_id"`
	BenefitName  string  `json:"benefit_name"`
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	Status       string  `json:"status"`
	EnrolledDate string  `json:"enrolled_date"`
	Cost         float64 `json:"cost"`
	Currency     string  `json:"currency"`
}

// GetEmployeeBenefits returns benefits for an employee
func (c *Client) GetEmployeeBenefits(ctx context.Context, employeeID string) ([]EmployeeBenefit, error) {
	path := fmt.Sprintf("/rest/v2/benefits/employees/%s", escapePath(employeeID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	benefits, err := decodeData[[]EmployeeBenefit](resp)
	if err != nil {
		return nil, err
	}
	return *benefits, nil
}
