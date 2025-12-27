package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Candidate represents a candidate in the ATS system
type Candidate struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// AddCandidateParams are params for adding a candidate
type AddCandidateParams struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

// AddCandidate creates a new candidate
func (c *Client) AddCandidate(ctx context.Context, params AddCandidateParams) (*Candidate, error) {
	resp, err := c.Post(ctx, "/rest/v2/candidates", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Candidate `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateCandidateParams are params for updating a candidate
type UpdateCandidateParams struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Status    string `json:"status,omitempty"`
}

// UpdateCandidate updates an existing candidate
func (c *Client) UpdateCandidate(ctx context.Context, id string, params UpdateCandidateParams) (*Candidate, error) {
	path := fmt.Sprintf("/rest/v2/candidates/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Candidate `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
