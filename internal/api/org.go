package api

import (
	"context"
	"fmt"
)

// Organization represents organization info
type Organization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	Industry  string `json:"industry"`
	Size      string `json:"size"`
	CreatedAt string `json:"created_at"`
}

// GetOrganization returns the organization
func (c *Client) GetOrganization(ctx context.Context) (*Organization, error) {
	resp, err := c.Get(ctx, "/rest/v2/organization")
	if err != nil {
		return nil, err
	}

	return decodeData[Organization](resp)
}

// OrgStructure represents an org structure node
type OrgStructure struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	ParentID string         `json:"parent_id"`
	Children []OrgStructure `json:"children"`
}

// GetOrgStructures returns the org structure
func (c *Client) GetOrgStructures(ctx context.Context) ([]OrgStructure, error) {
	resp, err := c.Get(ctx, "/rest/v2/hris/organization_structures")
	if err != nil {
		return nil, err
	}

	structures, err := decodeData[[]OrgStructure](resp)
	if err != nil {
		return nil, err
	}
	return *structures, nil
}

// LegalEntity represents a legal entity
type LegalEntity struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Country            string `json:"country"`
	Type               string `json:"type"`
	Status             string `json:"status"`
	RegistrationNumber string `json:"registration_number"`
}

// ListLegalEntities returns legal entities
func (c *Client) ListLegalEntities(ctx context.Context) ([]LegalEntity, error) {
	resp, err := c.Get(ctx, "/rest/v2/legal-entities")
	if err != nil {
		return nil, err
	}

	entities, err := decodeData[[]LegalEntity](resp)
	if err != nil {
		return nil, err
	}
	return *entities, nil
}

// CreateLegalEntityParams are params for creating a legal entity
type CreateLegalEntityParams struct {
	Name               string `json:"name"`
	Country            string `json:"country"`
	Type               string `json:"type"`
	RegistrationNumber string `json:"registration_number,omitempty"`
}

// CreateLegalEntity creates a new legal entity
func (c *Client) CreateLegalEntity(ctx context.Context, params CreateLegalEntityParams) (*LegalEntity, error) {
	resp, err := c.Post(ctx, "/rest/v2/legal-entities", params)
	if err != nil {
		return nil, err
	}

	return decodeData[LegalEntity](resp)
}

// UpdateLegalEntityParams are params for updating a legal entity
type UpdateLegalEntityParams struct {
	Name               string `json:"name,omitempty"`
	Type               string `json:"type,omitempty"`
	RegistrationNumber string `json:"registration_number,omitempty"`
}

// UpdateLegalEntity updates an existing legal entity
func (c *Client) UpdateLegalEntity(ctx context.Context, id string, params UpdateLegalEntityParams) (*LegalEntity, error) {
	path := fmt.Sprintf("/rest/v2/legal-entities/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	return decodeData[LegalEntity](resp)
}

// DeleteLegalEntity deletes a legal entity
func (c *Client) DeleteLegalEntity(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/v2/legal-entities/%s", escapePath(id))
	_, err := c.Delete(ctx, path)
	return err
}

// PayrollSettings represents payroll settings for a legal entity
type PayrollSettings struct {
	ID                string `json:"id"`
	LegalEntityID     string `json:"legal_entity_id"`
	PayrollFrequency  string `json:"payroll_frequency"`
	PaymentMethod     string `json:"payment_method"`
	Currency          string `json:"currency"`
	TaxID             string `json:"tax_id,omitempty"`
	BankAccount       string `json:"bank_account,omitempty"`
	PayrollProvider   string `json:"payroll_provider,omitempty"`
	AutoApproval      bool   `json:"auto_approval"`
	NotificationEmail string `json:"notification_email,omitempty"`
}

// GetPayrollSettings returns payroll settings for a legal entity
func (c *Client) GetPayrollSettings(ctx context.Context, id string) (*PayrollSettings, error) {
	path := fmt.Sprintf("/rest/v2/legal-entities/%s/payroll-settings", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[PayrollSettings](resp)
}
