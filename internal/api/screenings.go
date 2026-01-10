package api

import (
	"context"
	"fmt"
)

// VeriffSession represents a Veriff verification session
type VeriffSession struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
}

// CreateVeriffSessionParams are params for creating a Veriff session
type CreateVeriffSessionParams struct {
	WorkerID string `json:"worker_id"`
	Callback string `json:"callback,omitempty"`
}

// CreateVeriffSession creates a new Veriff verification session
func (c *Client) CreateVeriffSession(ctx context.Context, params CreateVeriffSessionParams) (*VeriffSession, error) {
	resp, err := c.Post(ctx, "/rest/v2/screenings/veriff/session", params)
	if err != nil {
		return nil, err
	}

	return decodeData[VeriffSession](resp)
}

// KYCDetails represents KYC verification details for a worker
type KYCDetails struct {
	WorkerID   string `json:"worker_id"`
	Status     string `json:"status"`
	VerifiedAt string `json:"verified_at,omitempty"`
	Provider   string `json:"provider,omitempty"`
	Details    struct {
		FirstName   string `json:"first_name,omitempty"`
		LastName    string `json:"last_name,omitempty"`
		DateOfBirth string `json:"date_of_birth,omitempty"`
		Country     string `json:"country,omitempty"`
	} `json:"details,omitempty"`
}

// GetKYCDetails returns KYC verification details for a worker
func (c *Client) GetKYCDetails(ctx context.Context, workerID string) (*KYCDetails, error) {
	path := fmt.Sprintf("/rest/v2/screenings/kyc/%s", escapePath(workerID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[KYCDetails](resp)
}

// AMLData represents anti-money laundering screening data
type AMLData struct {
	Results []struct {
		Name         string `json:"name"`
		Country      string `json:"country"`
		MatchType    string `json:"match_type"`
		RiskLevel    string `json:"risk_level"`
		ScreenedAt   string `json:"screened_at"`
		ListName     string `json:"list_name,omitempty"`
		MatchDetails string `json:"match_details,omitempty"`
	} `json:"results"`
	Summary struct {
		TotalMatches int    `json:"total_matches"`
		HighestRisk  string `json:"highest_risk"`
	} `json:"summary"`
}

// GetAMLData returns anti-money laundering screening data
func (c *Client) GetAMLData(ctx context.Context) (*AMLData, error) {
	resp, err := c.Get(ctx, "/rest/v2/screenings/aml")
	if err != nil {
		return nil, err
	}

	return decodeData[AMLData](resp)
}

// SubmitExternalKYCParams are params for submitting external KYC verification
type SubmitExternalKYCParams struct {
	WorkerID       string `json:"worker_id"`
	Provider       string `json:"provider"`
	VerifiedAt     string `json:"verified_at"`
	DocumentType   string `json:"document_type"`
	DocumentID     string `json:"document_id,omitempty"`
	ExpirationDate string `json:"expiration_date,omitempty"`
}

// ExternalKYCSubmission represents the response from submitting external KYC
type ExternalKYCSubmission struct {
	ID          string `json:"id"`
	WorkerID    string `json:"worker_id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submitted_at"`
}

// SubmitExternalKYC submits external KYC verification data
func (c *Client) SubmitExternalKYC(ctx context.Context, params SubmitExternalKYCParams) (*ExternalKYCSubmission, error) {
	resp, err := c.Post(ctx, "/rest/v2/screenings/kyc/external", params)
	if err != nil {
		return nil, err
	}

	return decodeData[ExternalKYCSubmission](resp)
}

// CreateManualVerificationParams are params for creating a manual verification
type CreateManualVerificationParams struct {
	WorkerID     string   `json:"worker_id"`
	VerifiedBy   string   `json:"verified_by"`
	Notes        string   `json:"notes,omitempty"`
	DocumentURLs []string `json:"document_urls,omitempty"`
}

// ManualVerification represents a manual verification record
type ManualVerification struct {
	ID           string   `json:"id"`
	WorkerID     string   `json:"worker_id"`
	VerifiedBy   string   `json:"verified_by"`
	Status       string   `json:"status"`
	Notes        string   `json:"notes,omitempty"`
	CreatedAt    string   `json:"created_at"`
	DocumentURLs []string `json:"document_urls,omitempty"`
}

// CreateManualVerification creates a manual verification record
func (c *Client) CreateManualVerification(ctx context.Context, params CreateManualVerificationParams) (*ManualVerification, error) {
	resp, err := c.Post(ctx, "/rest/v2/screenings/verification/manual", params)
	if err != nil {
		return nil, err
	}

	return decodeData[ManualVerification](resp)
}
