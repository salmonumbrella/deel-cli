package api

import (
	"context"
	"fmt"
)

// ComplianceDoc represents a compliance document
type ComplianceDoc struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Country    string `json:"country"`
	Status     string `json:"status"`
	Required   bool   `json:"required"`
	ExpiresAt  string `json:"expires_at"`
	UploadedAt string `json:"uploaded_at"`
}

// ListComplianceDocs returns compliance documents
func (c *Client) ListComplianceDocs(ctx context.Context, contractID string) ([]ComplianceDoc, error) {
	path := fmt.Sprintf("/rest/v2/compliance/contracts/%s/documents", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	docs, err := decodeData[[]ComplianceDoc](resp)
	if err != nil {
		return nil, err
	}
	return *docs, nil
}

// ComplianceTemplate represents a document template
type ComplianceTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Country     string `json:"country"`
	Description string `json:"description"`
}

// ListComplianceTemplates returns available templates
func (c *Client) ListComplianceTemplates(ctx context.Context, country string) ([]ComplianceTemplate, error) {
	path := fmt.Sprintf("/rest/v2/compliance/templates?country=%s", escapePath(country))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	templates, err := decodeData[[]ComplianceTemplate](resp)
	if err != nil {
		return nil, err
	}
	return *templates, nil
}

// ComplianceValidation represents a validation result
type ComplianceValidation struct {
	ContractID string `json:"contract_id"`
	Status     string `json:"status"`
	Issues     int    `json:"issues_count"`
	LastCheck  string `json:"last_check"`
}

// GetComplianceValidations returns validation status
func (c *Client) GetComplianceValidations(ctx context.Context, contractID string) (*ComplianceValidation, error) {
	path := fmt.Sprintf("/rest/v2/compliance/contracts/%s/validations", escapePath(contractID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[ComplianceValidation](resp)
}
