package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Person represents a Deel person/worker
type Person struct {
	ID           string `json:"id"`
	HRISProfileID string `json:"hris_profile_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	JobTitle     string `json:"job_title"`
	Department   string `json:"department"`
	Status       string `json:"status"`
	StartDate    string `json:"start_date"`
	Country      string `json:"country"`
}

// PeopleListResponse is the response from list people
type PeopleListResponse struct {
	Data []Person `json:"data"`
	Page struct {
		Next  string `json:"next"`
		Total int    `json:"total"`
	} `json:"page"`
}

// PeopleListParams are params for listing people
type PeopleListParams struct {
	Limit  int
	Cursor string
}

// ListPeople returns a list of people
func (c *Client) ListPeople(ctx context.Context, params PeopleListParams) (*PeopleListResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/people"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result PeopleListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetPerson returns a single person by HRIS profile ID
func (c *Client) GetPerson(ctx context.Context, hrisProfileID string) (*Person, error) {
	path := fmt.Sprintf("/rest/v2/people/%s", escapePath(hrisProfileID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Person `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// SearchPeopleByEmail finds a person by email
func (c *Client) SearchPeopleByEmail(ctx context.Context, email string) (*Person, error) {
	q := url.Values{}
	q.Set("email", email)
	path := "/rest/v2/people/search?" + q.Encode()

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Person `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// CustomField represents a custom field
type CustomField struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

// ListCustomFields returns custom fields for people
func (c *Client) ListCustomFields(ctx context.Context) ([]CustomField, error) {
	resp, err := c.Get(ctx, "/rest/v2/people/custom-fields")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []CustomField `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// GetCustomField returns a specific custom field
func (c *Client) GetCustomField(ctx context.Context, fieldID string) (*CustomField, error) {
	path := fmt.Sprintf("/rest/v2/people/custom-fields/%s", escapePath(fieldID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data CustomField `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
