package api

import (
	"context"
	"fmt"
	"net/url"
)

// Payslip represents a payslip
type Payslip struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Status string `json:"status"`
}

// GetEORWorkerPayslips returns payslips for an EOR worker
func (c *Client) GetEORWorkerPayslips(ctx context.Context, workerID string) ([]Payslip, error) {
	path := fmt.Sprintf("/rest/v2/eor/workers/%s/payslips", escapePath(workerID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	payslips, err := decodeData[[]Payslip](resp)
	if err != nil {
		return nil, err
	}
	return *payslips, nil
}

// GetGPWorkerPayslips returns payslips for a Global Payroll worker
func (c *Client) GetGPWorkerPayslips(ctx context.Context, workerID string) ([]Payslip, error) {
	path := fmt.Sprintf("/rest/v2/gp/workers/%s/payslips", escapePath(workerID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	payslips, err := decodeData[[]Payslip](resp)
	if err != nil {
		return nil, err
	}
	return *payslips, nil
}

// PayslipDownload represents a payslip download URL response
type PayslipDownload struct {
	URL string `json:"url"`
}

// GetGPPayslipDownloadURL returns a pre-signed download URL for a GP payslip PDF
func (c *Client) GetGPPayslipDownloadURL(ctx context.Context, workerID, payslipID string) (string, error) {
	path := fmt.Sprintf("/rest/v2/gp/workers/%s/payslips/%s/download", escapePath(workerID), escapePath(payslipID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return "", err
	}

	data, err := decodeData[PayslipDownload](resp)
	if err != nil {
		return "", err
	}
	return data.URL, nil
}

// PaymentBreakdown represents a payment breakdown
type PaymentBreakdown struct {
	CycleID     string  `json:"cycle_id"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`
	Workers     int     `json:"worker_count"`
	Status      string  `json:"status"`
}

// GetPaymentBreakdown returns payment breakdown for a cycle
func (c *Client) GetPaymentBreakdown(ctx context.Context, cycleID string) (*PaymentBreakdown, error) {
	path := fmt.Sprintf("/rest/v2/payments/cycles/%s/breakdown", escapePath(cycleID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[PaymentBreakdown](resp)
}

// PaymentReceipt represents a payment receipt
type PaymentReceipt struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Date      string  `json:"date"`
	Reference string  `json:"reference"`
}

// ListPaymentReceipts returns payment receipts
func (c *Client) ListPaymentReceipts(ctx context.Context, limit int) ([]PaymentReceipt, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := "/rest/v2/payments/receipts"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	receipts, err := decodeData[[]PaymentReceipt](resp)
	if err != nil {
		return nil, err
	}
	return *receipts, nil
}
