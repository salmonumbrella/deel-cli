package api

import (
	"context"
	"fmt"
	"net/url"
)

// ITAsset represents an IT asset
type ITAsset struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	SerialNumber string `json:"serial_number"`
	Status       string `json:"status"`
	AssignedTo   string `json:"assigned_to"`
	AssignedDate string `json:"assigned_date"`
	Condition    string `json:"condition"`
}

// ITAssetsListResponse is the response from list IT assets
type ITAssetsListResponse = ListResponse[ITAsset]

// ITAssetsListParams are params for listing IT assets
type ITAssetsListParams struct {
	Status string
	Type   string
	Limit  int
	Cursor string
}

// ListITAssets returns IT assets
func (c *Client) ListITAssets(ctx context.Context, params ITAssetsListParams) (*ITAssetsListResponse, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Type != "" {
		q.Set("type", params.Type)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/it/assets"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ITAsset](resp)
}

// ITOrder represents an IT equipment order
type ITOrder struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Status       string  `json:"status"`
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name"`
	Items        int     `json:"items_count"`
	TotalCost    float64 `json:"total_cost"`
	Currency     string  `json:"currency"`
	OrderDate    string  `json:"order_date"`
}

// ListITOrders returns IT orders
func (c *Client) ListITOrders(ctx context.Context, limit int) ([]ITOrder, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := "/rest/v2/it/orders"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	orders, err := decodeData[[]ITOrder](resp)
	if err != nil {
		return nil, err
	}
	return *orders, nil
}

// HardwarePolicy represents an IT hardware policy
type HardwarePolicy struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Country     string  `json:"country"`
	Budget      float64 `json:"budget"`
	Currency    string  `json:"currency"`
}

// ListHardwarePolicies returns hardware policies
func (c *Client) ListHardwarePolicies(ctx context.Context) ([]HardwarePolicy, error) {
	resp, err := c.Get(ctx, "/rest/v2/it/policies")
	if err != nil {
		return nil, err
	}

	policies, err := decodeData[[]HardwarePolicy](resp)
	if err != nil {
		return nil, err
	}
	return *policies, nil
}
