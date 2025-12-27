package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Team represents a team
type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ManagerID   string `json:"manager_id"`
	ManagerName string `json:"manager_name"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

// TeamsListResponse is the response from list teams
type TeamsListResponse struct {
	Data []Team `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"page"`
}

// ListTeams returns teams
func (c *Client) ListTeams(ctx context.Context, limit int, cursor string) (*TeamsListResponse, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}

	path := "/rest/v2/teams"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result TeamsListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetTeam returns a single team
func (c *Client) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	path := fmt.Sprintf("/rest/v2/teams/%s", escapePath(teamID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Team `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
