package api

import (
	"context"
	"fmt"
	"net/url"
)

// G2NReport represents a gross-to-net report for a Global Payroll worker
type G2NReport struct {
	ID          string  `json:"id"`
	WorkerID    string  `json:"worker_id"`
	WorkerName  string  `json:"worker_name"`
	Period      string  `json:"period"`
	GrossAmount float64 `json:"gross_amount"`
	NetAmount   float64 `json:"net_amount"`
	Deductions  float64 `json:"deductions"`
	Taxes       float64 `json:"taxes"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// G2NReportDownload represents a downloadable gross-to-net report
type G2NReportDownload struct {
	ReportID    string `json:"report_id"`
	DownloadURL string `json:"download_url"`
	ExpiresAt   string `json:"expires_at"`
	Format      string `json:"format"`
}

// GPTerminationRequest represents a termination request for a GP worker
type GPTerminationRequest struct {
	ID            string `json:"id"`
	WorkerID      string `json:"worker_id"`
	Reason        string `json:"reason"`
	EffectiveDate string `json:"effective_date"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

// ListG2NReportsParams are parameters for listing gross-to-net reports
type ListG2NReportsParams struct {
	WorkerID string
	Period   string
}

// RequestGPTerminationParams are parameters for requesting GP worker termination
type RequestGPTerminationParams struct {
	WorkerID      string `json:"worker_id"`
	Reason        string `json:"reason"`
	EffectiveDate string `json:"effective_date"`
}

// ListG2NReports retrieves gross-to-net reports for Global Payroll workers
func (c *Client) ListG2NReports(ctx context.Context, params ListG2NReportsParams) ([]G2NReport, error) {
	path := "/rest/v2/gp/reports/gross-to-net"

	// Build query parameters
	query := url.Values{}
	if params.WorkerID != "" {
		query.Set("worker_id", params.WorkerID)
	}
	if params.Period != "" {
		query.Set("period", params.Period)
	}

	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	reports, err := decodeData[[]G2NReport](resp)
	if err != nil {
		return nil, err
	}
	return *reports, nil
}

// DownloadG2NReport retrieves a download URL for a specific gross-to-net report
func (c *Client) DownloadG2NReport(ctx context.Context, reportID string) (*G2NReportDownload, error) {
	path := fmt.Sprintf("/rest/v2/gp/reports/gross-to-net/download?report_id=%s", escapePath(reportID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[G2NReportDownload](resp)
}

// RequestGPTermination creates a termination request for a Global Payroll worker
func (c *Client) RequestGPTermination(ctx context.Context, params RequestGPTerminationParams) (*GPTerminationRequest, error) {
	resp, err := c.Post(ctx, "/rest/v2/gp/termination-requests", params)
	if err != nil {
		return nil, err
	}

	return decodeData[GPTerminationRequest](resp)
}
