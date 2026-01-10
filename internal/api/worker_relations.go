package api

import (
	"context"
	"fmt"
)

// WorkerRelation represents a reporting relationship between a worker and manager
type WorkerRelation struct {
	ID           string `json:"id"`
	ProfileID    string `json:"profile_id"`
	ManagerID    string `json:"manager_id"`
	RelationType string `json:"relation_type"` // "direct_report", "dotted_line"
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date,omitempty"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// CreateWorkerRelationParams contains parameters for creating a worker relation
type CreateWorkerRelationParams struct {
	ProfileID    string `json:"profile_id"`
	ManagerID    string `json:"manager_id"`
	RelationType string `json:"relation_type"`
	StartDate    string `json:"start_date"`
}

// ListWorkerRelations retrieves all worker relations for a given profile
func (c *Client) ListWorkerRelations(ctx context.Context, profileID string) ([]WorkerRelation, error) {
	path := fmt.Sprintf("/rest/v2/profiles/%s/worker-relations", escapePath(profileID))

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	relations, err := decodeData[[]WorkerRelation](resp)
	if err != nil {
		return nil, err
	}

	return *relations, nil
}

// CreateWorkerRelation creates a new worker relation
func (c *Client) CreateWorkerRelation(ctx context.Context, params CreateWorkerRelationParams) (*WorkerRelation, error) {
	resp, err := c.Post(ctx, "/rest/v2/worker-relations", params)
	if err != nil {
		return nil, err
	}

	return decodeData[WorkerRelation](resp)
}

// DeleteWorkerRelation deletes a worker relation by ID
func (c *Client) DeleteWorkerRelation(ctx context.Context, relationID string) error {
	path := fmt.Sprintf("/rest/v2/worker-relations/%s", escapePath(relationID))
	_, err := c.Delete(ctx, path)
	return err
}

// SetWorkerManagerParams contains parameters for setting a worker's manager
type SetWorkerManagerParams struct {
	ManagerID string `json:"manager_id"`
	StartDate string `json:"start_date,omitempty"`
}

// SetWorkerManager assigns a manager to a worker using the HRIS parent relation endpoint.
// This is the recommended endpoint per Deel support (Khizar) for assigning managers.
// PUT /rest/v2/hris/worker_relations/profile/{hrisProfileOid}/parent
func (c *Client) SetWorkerManager(ctx context.Context, hrisProfileID string, params SetWorkerManagerParams) (*WorkerRelation, error) {
	path := fmt.Sprintf("/rest/v2/hris/worker_relations/profile/%s/parent", escapePath(hrisProfileID))

	resp, err := c.Put(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[WorkerRelation](resp)
}
