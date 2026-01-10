package api

import (
	"context"
	"fmt"
)

// Group represents a Deel group
type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

// CreateGroupParams are params for creating a group
type CreateGroupParams struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateGroupParams are params for updating a group
type UpdateGroupParams struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListGroups returns all groups
func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
	resp, err := c.Get(ctx, "/rest/v2/groups")
	if err != nil {
		return nil, err
	}

	groups, err := decodeData[[]Group](resp)
	if err != nil {
		return nil, err
	}
	return *groups, nil
}

// GetGroup returns a single group by ID
func (c *Client) GetGroup(ctx context.Context, id string) (*Group, error) {
	path := fmt.Sprintf("/rest/v2/groups/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[Group](resp)
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, params CreateGroupParams) (*Group, error) {
	resp, err := c.Post(ctx, "/rest/v2/groups", params)
	if err != nil {
		return nil, err
	}

	return decodeData[Group](resp)
}

// UpdateGroup updates an existing group
func (c *Client) UpdateGroup(ctx context.Context, id string, params UpdateGroupParams) (*Group, error) {
	path := fmt.Sprintf("/rest/v2/groups/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[Group](resp)
}

// DeleteGroup deletes a group
func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/v2/groups/%s", escapePath(id))
	_, err := c.Delete(ctx, path)
	return err
}

// CloneGroup clones an existing group
func (c *Client) CloneGroup(ctx context.Context, id string) (*Group, error) {
	path := fmt.Sprintf("/rest/v2/groups/%s/clone", escapePath(id))
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	return decodeData[Group](resp)
}
