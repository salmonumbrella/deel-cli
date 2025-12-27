package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Manager represents a Deel manager/admin user
type Manager struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Role      string   `json:"role"`
	Status    string   `json:"status"`
	TeamIDs   []string `json:"team_ids,omitempty"`
	CreatedAt string   `json:"created_at"`
}

// ManagerMagicLink represents a magic link for manager authentication
type ManagerMagicLink struct {
	ManagerID string `json:"manager_id"`
	Link      string `json:"link"`
	ExpiresAt string `json:"expires_at"`
}

// CreateManagerParams are params for creating a manager
type CreateManagerParams struct {
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Role      string   `json:"role"`
	TeamIDs   []string `json:"team_ids,omitempty"`
}

// CreateMagicLinkParams are params for creating a magic link
type CreateMagicLinkParams struct {
	Email string `json:"email"`
}

// ListManagers returns a list of managers
func (c *Client) ListManagers(ctx context.Context) ([]Manager, error) {
	resp, err := c.Get(ctx, "/rest/v2/managers")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Manager `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// CreateManager creates a new manager
func (c *Client) CreateManager(ctx context.Context, params CreateManagerParams) (*Manager, error) {
	resp, err := c.Post(ctx, "/rest/v2/managers", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Manager `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// CreateManagerMagicLink creates a magic link for manager authentication
func (c *Client) CreateManagerMagicLink(ctx context.Context, params CreateMagicLinkParams) (*ManagerMagicLink, error) {
	resp, err := c.Post(ctx, "/rest/v2/managers/magic-link", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data ManagerMagicLink `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
