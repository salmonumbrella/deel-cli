package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// ImmigrationCase represents an immigration case
type ImmigrationCase struct {
	ID         string `json:"id"`
	ContractID string `json:"contract_id"`
	WorkerName string `json:"worker_name"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Country    string `json:"country"`
	StartDate  string `json:"start_date"`
	ExpiryDate string `json:"expiry_date"`
	CaseNumber string `json:"case_number"`
}

// GetImmigrationCaseDetails returns case details
func (c *Client) GetImmigrationCaseDetails(ctx context.Context, caseID string) (*ImmigrationCase, error) {
	path := fmt.Sprintf("/rest/v2/immigration/cases/%s", escapePath(caseID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data ImmigrationCase `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// ImmigrationDoc represents an immigration document
type ImmigrationDoc struct {
	ID        string `json:"id"`
	CaseID    string `json:"case_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
}

// ListImmigrationDocs returns documents for a case
func (c *Client) ListImmigrationDocs(ctx context.Context, caseID string) ([]ImmigrationDoc, error) {
	path := fmt.Sprintf("/rest/v2/immigration/cases/%s/documents", escapePath(caseID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []ImmigrationDoc `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// VisaType represents a visa type
type VisaType struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Country      string `json:"country"`
	Category     string `json:"category"`
	Duration     string `json:"max_duration"`
	Requirements string `json:"requirements_summary"`
}

// ListVisaTypes returns visa types for a country
func (c *Client) ListVisaTypes(ctx context.Context, country string) ([]VisaType, error) {
	path := fmt.Sprintf("/rest/v2/immigration/visa-types?country=%s", escapePath(country))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []VisaType `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// VisaRequirement represents a visa requirement check
type VisaRequirement struct {
	Required bool   `json:"visa_required"`
	Type     string `json:"suggested_type"`
	Duration string `json:"max_stay"`
	Notes    string `json:"notes"`
}

// CheckVisaRequirement checks if visa is required
func (c *Client) CheckVisaRequirement(ctx context.Context, fromCountry, toCountry string) (*VisaRequirement, error) {
	path := fmt.Sprintf("/rest/v2/immigration/check?from=%s&to=%s", escapePath(fromCountry), escapePath(toCountry))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data VisaRequirement `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// CreateImmigrationCaseParams are params for creating an immigration case
type CreateImmigrationCaseParams struct {
	ContractID string `json:"contract_id"`
	Type       string `json:"type"`
	Country    string `json:"country"`
	StartDate  string `json:"start_date"`
}

// CreateImmigrationCase creates a new immigration case
func (c *Client) CreateImmigrationCase(ctx context.Context, params CreateImmigrationCaseParams) (*ImmigrationCase, error) {
	resp, err := c.Post(ctx, "/rest/v2/immigration/cases", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data ImmigrationCase `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UploadImmigrationDocumentParams are params for uploading an immigration document
type UploadImmigrationDocumentParams struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DocumentURL string `json:"document_url"`
}

// UploadImmigrationDocument uploads a document to an immigration case
func (c *Client) UploadImmigrationDocument(ctx context.Context, caseID string, params UploadImmigrationDocumentParams) (*ImmigrationDoc, error) {
	path := fmt.Sprintf("/rest/v2/immigration/cases/%s/documents", escapePath(caseID))
	resp, err := c.Post(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data ImmigrationDoc `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}
