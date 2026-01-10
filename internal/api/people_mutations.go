package api

import (
	"context"
	"fmt"
)

// PersonalInfo represents personal information for a person
type PersonalInfo struct {
	ID          string `json:"id,omitempty"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Nationality string `json:"nationality,omitempty"`
}

// WorkingLocation represents the working location of a person
type WorkingLocation struct {
	ID         string `json:"id,omitempty"`
	Country    string `json:"country"`
	State      string `json:"state,omitempty"`
	City       string `json:"city,omitempty"`
	Address    string `json:"address,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

// CreatePersonParams are parameters for creating a person
type CreatePersonParams struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Type      string `json:"type"`
	Country   string `json:"country"`
}

// CreateDirectEmployeeParams are parameters for creating a direct employee
type CreateDirectEmployeeParams struct {
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Country    string `json:"country"`
	StartDate  string `json:"start_date"`
	JobTitle   string `json:"job_title"`
	Department string `json:"department,omitempty"`
}

// PersonResponse represents the enhanced person response with additional fields
type PersonResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Type        string `json:"type"`
	Country     string `json:"country"`
	State       string `json:"state,omitempty"`
	City        string `json:"city,omitempty"`
	Address     string `json:"address,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	Status      string `json:"status"`
	StartDate   string `json:"start_date,omitempty"`
	JobTitle    string `json:"job_title,omitempty"`
	Department  string `json:"department,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// CreatePerson creates a new person
func (c *Client) CreatePerson(ctx context.Context, params CreatePersonParams) (*PersonResponse, error) {
	resp, err := c.Post(ctx, "/rest/v2/people", params)
	if err != nil {
		return nil, err
	}

	return decodeData[PersonResponse](resp)
}

// UpdatePersonalInfo updates personal information for a person
func (c *Client) UpdatePersonalInfo(ctx context.Context, id string, info PersonalInfo) (*PersonalInfo, error) {
	path := fmt.Sprintf("/rest/v2/people/%s/personal-info", escapePath(id))
	resp, err := c.Patch(ctx, path, info)
	if err != nil {
		return nil, err
	}

	return decodeData[PersonalInfo](resp)
}

// UpdateWorkingLocation updates the working location for a person
func (c *Client) UpdateWorkingLocation(ctx context.Context, id string, location WorkingLocation) (*WorkingLocation, error) {
	path := fmt.Sprintf("/rest/v2/people/%s/working-location", escapePath(id))
	resp, err := c.Put(ctx, path, location)
	if err != nil {
		return nil, err
	}

	return decodeData[WorkingLocation](resp)
}

// CreateDirectEmployee creates a new direct employee
func (c *Client) CreateDirectEmployee(ctx context.Context, params CreateDirectEmployeeParams) (*PersonResponse, error) {
	resp, err := c.Post(ctx, "/rest/v2/people/direct-employee", params)
	if err != nil {
		return nil, err
	}

	return decodeData[PersonResponse](resp)
}
