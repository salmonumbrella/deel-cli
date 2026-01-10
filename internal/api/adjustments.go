package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
)

// Adjustment represents a contract adjustment (bonus, deduction, expense)
type Adjustment struct {
	ID                   string  `json:"id"`
	ContractID           string  `json:"contract_id"`
	CategoryID           string  `json:"category_id"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	Description          string  `json:"description"`
	Date                 string  `json:"date"`
	Status               string  `json:"status"`
	CreatedAt            string  `json:"created_at"`
	CycleReference       string  `json:"cycle_reference,omitempty"`
	MoveNextCycle        bool    `json:"move_next_cycle,omitempty"`
	ActualStartCycleDate string  `json:"actual_start_cycle_date,omitempty"`
	ActualEndCycleDate   string  `json:"actual_end_cycle_date,omitempty"`
}

// AdjustmentCategory represents a category of adjustments
type AdjustmentCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"` // "bonus", "deduction", "expense"
}

// CreateAdjustmentParams are parameters for creating an adjustment
type CreateAdjustmentParams struct {
	ContractID     string  `json:"contract_id"`
	CategoryID     string  `json:"category_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Description    string  `json:"description"`
	Date           string  `json:"date"`
	CycleReference string  `json:"cycle_reference,omitempty"`
	MoveNextCycle  bool    `json:"move_next_cycle,omitempty"`
	Vendor         string  `json:"vendor,omitempty"`         // Vendor name (defaults to "Company" if empty)
	Country        string  `json:"country,omitempty"`        // ISO 3166-1 alpha-2 country code (defaults to "CA" if empty)
}

// UpdateAdjustmentParams are parameters for updating an adjustment
type UpdateAdjustmentParams struct {
	Amount      float64 `json:"amount,omitempty"`
	Description string  `json:"description,omitempty"`
	Date        string  `json:"date,omitempty"`
}

// ListAdjustmentsParams are parameters for listing adjustments
type ListAdjustmentsParams struct {
	ContractID string
	CategoryID string
}

// CreateAdjustment creates a new adjustment using multipart/form-data
func (c *Client) CreateAdjustment(ctx context.Context, params CreateAdjustmentParams) (*Adjustment, error) {
	// Build multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Helper to track WriteField errors
	var writeErr error
	writeField := func(key, value string) {
		if writeErr != nil {
			return
		}
		writeErr = writer.WriteField(key, value)
	}

	// Required fields
	writeField("contract_id", params.ContractID)
	writeField("adjustment_category_id", params.CategoryID)
	writeField("amount", fmt.Sprintf("%.2f", params.Amount))
	writeField("currency", params.Currency)
	writeField("description", params.Description)
	writeField("date_of_adjustment", params.Date)

	// Required by API: vendor and country
	vendor := params.Vendor
	if vendor == "" {
		vendor = "Company" // Default vendor placeholder
	}
	writeField("vendor", vendor)

	country := params.Country
	if country == "" {
		country = "CA" // Default to Canada
	}
	writeField("country", country)

	// Create placeholder file part (required by API)
	// Some adjustment types require a file, we provide a minimal placeholder with correct MIME type
	if writeErr == nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="file"; filename="adjustment.txt"`)
		h.Set("Content-Type", "text/plain")
		part, err := writer.CreatePart(h)
		if err != nil {
			writeErr = err
		} else {
			// Write placeholder content
			_, writeErr = part.Write([]byte("Adjustment record"))
		}
	}

	// Optional fields
	if params.CycleReference != "" {
		writeField("cycle_reference", params.CycleReference)
	}
	if params.MoveNextCycle {
		writeField("move_next_cycle", "true")
	}

	// Title field (use description as title)
	writeField("title", params.Description)

	if writeErr != nil {
		return nil, fmt.Errorf("failed to build form: %w", writeErr)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize form: %w", err)
	}

	// Use doMultipart for retry logic, circuit breaker, and error handling.
	// Pass bytes.NewReader so the body can be re-read on retries.
	resp, err := c.doMultipart(ctx, http.MethodPost, "/rest/v2/adjustments", bytes.NewReader(buf.Bytes()), writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// GetAdjustment returns a single adjustment
func (c *Client) GetAdjustment(ctx context.Context, id string) (*Adjustment, error) {
	path := fmt.Sprintf("/rest/v2/adjustments/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateAdjustment updates an existing adjustment
func (c *Client) UpdateAdjustment(ctx context.Context, id string, params UpdateAdjustmentParams) (*Adjustment, error) {
	path := fmt.Sprintf("/rest/v2/adjustments/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteAdjustment deletes an adjustment
func (c *Client) DeleteAdjustment(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/v2/adjustments/%s", escapePath(id))
	_, err := c.Delete(ctx, path)
	return err
}

// ListAdjustments returns adjustments with optional filters
func (c *Client) ListAdjustments(ctx context.Context, params ListAdjustmentsParams) ([]Adjustment, error) {
	q := url.Values{}
	if params.ContractID != "" {
		q.Set("contract_id", params.ContractID)
	}
	if params.CategoryID != "" {
		q.Set("category_id", params.CategoryID)
	}

	path := "/rest/v2/adjustments"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Adjustment `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// ListAdjustmentCategories returns all available adjustment categories
func (c *Client) ListAdjustmentCategories(ctx context.Context) ([]AdjustmentCategory, error) {
	resp, err := c.Get(ctx, "/rest/v2/adjustments/categories")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []AdjustmentCategory `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}
