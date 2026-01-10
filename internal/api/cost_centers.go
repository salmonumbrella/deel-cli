package api

import "context"

// CostCenter represents a cost center
type CostCenter struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// ListCostCenters returns all cost centers
func (c *Client) ListCostCenters(ctx context.Context) ([]CostCenter, error) {
	resp, err := c.Get(ctx, "/rest/v2/cost-centers")
	if err != nil {
		return nil, err
	}

	centers, err := decodeData[[]CostCenter](resp)
	if err != nil {
		return nil, err
	}
	return *centers, nil
}

// CostCenterInput represents input for creating/updating a cost center
type CostCenterInput struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

// SyncCostCentersParams are params for syncing cost centers
type SyncCostCentersParams struct {
	CostCenters []CostCenterInput `json:"cost_centers"`
}

// SyncCostCenters syncs cost centers
func (c *Client) SyncCostCenters(ctx context.Context, params SyncCostCentersParams) ([]CostCenter, error) {
	resp, err := c.Post(ctx, "/rest/v2/cost-centers/sync", params)
	if err != nil {
		return nil, err
	}

	centers, err := decodeData[[]CostCenter](resp)
	if err != nil {
		return nil, err
	}
	return *centers, nil
}
