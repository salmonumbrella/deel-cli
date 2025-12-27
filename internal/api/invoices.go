package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Invoice represents a Deel invoice
type Invoice struct {
	ID          string  `json:"id"`
	Number      string  `json:"number"`
	Status      string  `json:"status"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	DueDate     string  `json:"due_date"`
	PaidDate    string  `json:"paid_date"`
	ContractID  string  `json:"contract_id"`
	WorkerName  string  `json:"worker_name"`
	Description string  `json:"description"`
}

// InvoicesListResponse is the response from list invoices
type InvoicesListResponse struct {
	Data []Invoice `json:"data"`
	Page struct {
		Next  string `json:"next"`
		Total int    `json:"total"`
	} `json:"page"`
}

// InvoicesListParams are params for listing invoices
type InvoicesListParams struct {
	Limit      int
	Cursor     string
	Status     string
	ContractID string
}

// ListInvoices returns a list of invoices
func (c *Client) ListInvoices(ctx context.Context, params InvoicesListParams) (*InvoicesListResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.ContractID != "" {
		q.Set("contract_id", params.ContractID)
	}

	path := "/rest/v2/invoices"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result InvoicesListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetInvoice returns a single invoice
func (c *Client) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	path := fmt.Sprintf("/rest/v2/invoices/%s", escapePath(invoiceID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Invoice `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// InvoiceAdjustment represents an invoice adjustment
type InvoiceAdjustment struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// ListInvoiceAdjustments returns adjustments for an invoice
func (c *Client) ListInvoiceAdjustments(ctx context.Context, invoiceID string) ([]InvoiceAdjustment, error) {
	path := fmt.Sprintf("/rest/v2/invoices/%s/adjustments", escapePath(invoiceID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []InvoiceAdjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// CreateInvoiceAdjustmentParams are params for creating an adjustment
type CreateInvoiceAdjustmentParams struct {
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description,omitempty"`
}

// CreateInvoiceAdjustment creates an adjustment on an invoice
func (c *Client) CreateInvoiceAdjustment(ctx context.Context, invoiceID string, params CreateInvoiceAdjustmentParams) (*InvoiceAdjustment, error) {
	path := fmt.Sprintf("/rest/v2/invoices/%s/adjustments", escapePath(invoiceID))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data InvoiceAdjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// GetInvoicePDF returns the PDF bytes for an invoice
func (c *Client) GetInvoicePDF(ctx context.Context, invoiceID string) ([]byte, error) {
	path := fmt.Sprintf("/rest/v2/invoices/%s/pdf", escapePath(invoiceID))
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/pdf")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	pdfBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	return pdfBytes, nil
}

// DeelInvoice represents a Deel-issued invoice
type DeelInvoice struct {
	ID          string  `json:"id"`
	Number      string  `json:"number"`
	Status      string  `json:"status"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	IssueDate   string  `json:"issue_date"`
	DueDate     string  `json:"due_date"`
	PaidDate    string  `json:"paid_date"`
	Description string  `json:"description"`
}

// DeelInvoicesListResponse is the response from list Deel invoices
type DeelInvoicesListResponse struct {
	Data []DeelInvoice `json:"data"`
	Page struct {
		Next  string `json:"next"`
		Total int    `json:"total"`
	} `json:"page"`
}

// DeelInvoicesListParams are params for listing Deel invoices
type DeelInvoicesListParams struct {
	Limit  int
	Cursor string
	Status string
}

// ListDeelInvoices returns a list of Deel-issued invoices
func (c *Client) ListDeelInvoices(ctx context.Context, params DeelInvoicesListParams) (*DeelInvoicesListResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}

	path := "/rest/v2/deel-invoices"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result DeelInvoicesListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}
