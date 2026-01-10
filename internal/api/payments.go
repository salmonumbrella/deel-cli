package api

import (
	"context"
	"fmt"
	"net/url"
)

// OffCyclePayment represents an off-cycle payment
type OffCyclePayment struct {
	ID          string  `json:"id"`
	ContractID  string  `json:"contract_id"`
	WorkerName  string  `json:"worker_name"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
	PaymentDate string  `json:"payment_date"`
	CreatedAt   string  `json:"created_at"`
}

// OffCyclePaymentsListResponse is the response from list off-cycle payments
type OffCyclePaymentsListResponse = ListResponse[OffCyclePayment]

// OffCyclePaymentsListParams are params for listing off-cycle payments
type OffCyclePaymentsListParams struct {
	ContractID string
	Status     string
	Limit      int
	Cursor     string
}

// ListOffCyclePayments returns off-cycle payments
func (c *Client) ListOffCyclePayments(ctx context.Context, params OffCyclePaymentsListParams) (*OffCyclePaymentsListResponse, error) {
	q := url.Values{}
	if params.ContractID != "" {
		q.Set("contract_id", params.ContractID)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/payments/off-cycle"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[OffCyclePayment](resp)
}

// CreateOffCyclePaymentParams are params for creating an off-cycle payment
type CreateOffCyclePaymentParams struct {
	ContractID  string  `json:"contract_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"`
	Description string  `json:"description,omitempty"`
	PaymentDate string  `json:"payment_date"`
}

// CreateOffCyclePayment creates a new off-cycle payment
func (c *Client) CreateOffCyclePayment(ctx context.Context, params CreateOffCyclePaymentParams) (*OffCyclePayment, error) {
	resp, err := c.Post(ctx, "/rest/v2/payments/off-cycle", params)
	if err != nil {
		return nil, err
	}

	return decodeData[OffCyclePayment](resp)
}

// IndividualPaymentBreakdown represents a detailed breakdown of a single payment
type IndividualPaymentBreakdown struct {
	PaymentID       string  `json:"payment_id"`
	GrossAmount     float64 `json:"gross_amount"`
	NetAmount       float64 `json:"net_amount"`
	Currency        string  `json:"currency"`
	DeelFee         float64 `json:"deel_fee"`
	WithholdingTax  float64 `json:"withholding_tax"`
	OtherDeductions float64 `json:"other_deductions"`
	Reimbursements  float64 `json:"reimbursements"`
	LineItems       []struct {
		Description string  `json:"description"`
		Amount      float64 `json:"amount"`
		Type        string  `json:"type"`
	} `json:"line_items"`
}

// GetIndividualPaymentBreakdown returns detailed breakdown for a single payment
func (c *Client) GetIndividualPaymentBreakdown(ctx context.Context, paymentID string) (*IndividualPaymentBreakdown, error) {
	path := fmt.Sprintf("/rest/v2/payments/%s/breakdown", escapePath(paymentID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[IndividualPaymentBreakdown](resp)
}

// DetailedPaymentReceipt represents a detailed payment receipt
type DetailedPaymentReceipt struct {
	ID          string  `json:"id"`
	PaymentID   string  `json:"payment_id"`
	ReceiptURL  string  `json:"receipt_url"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	IssueDate   string  `json:"issue_date"`
	ContractID  string  `json:"contract_id"`
	WorkerName  string  `json:"worker_name"`
	Description string  `json:"description"`
}

// DetailedPaymentReceiptsListResponse is the response from list detailed payment receipts
type DetailedPaymentReceiptsListResponse = ListResponse[DetailedPaymentReceipt]

// DetailedPaymentReceiptsListParams are params for listing detailed payment receipts
type DetailedPaymentReceiptsListParams struct {
	Limit      int
	Cursor     string
	ContractID string
	PaymentID  string
}

// ListDetailedPaymentReceipts returns a list of detailed payment receipts
func (c *Client) ListDetailedPaymentReceipts(ctx context.Context, params DetailedPaymentReceiptsListParams) (*DetailedPaymentReceiptsListResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}
	if params.ContractID != "" {
		q.Set("contract_id", params.ContractID)
	}
	if params.PaymentID != "" {
		q.Set("payment_id", params.PaymentID)
	}

	path := "/rest/v2/payment-receipts"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[DetailedPaymentReceipt](resp)
}

// DetailedPaymentsReport represents a detailed payments report
type DetailedPaymentsReport struct {
	ReportID    string  `json:"report_id"`
	GeneratedAt string  `json:"generated_at"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`
	Payments    []struct {
		PaymentID   string  `json:"payment_id"`
		ContractID  string  `json:"contract_id"`
		WorkerName  string  `json:"worker_name"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		PaymentDate string  `json:"payment_date"`
		Status      string  `json:"status"`
		Type        string  `json:"type"`
	} `json:"payments"`
}

// DetailedPaymentsReportParams are params for generating a detailed payments report
type DetailedPaymentsReportParams struct {
	StartDate  string
	EndDate    string
	ContractID string
	Status     string
}

// GetDetailedPaymentsReport returns a detailed payments report
func (c *Client) GetDetailedPaymentsReport(ctx context.Context, params DetailedPaymentsReportParams) (*DetailedPaymentsReport, error) {
	q := url.Values{}
	if params.StartDate != "" {
		q.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		q.Set("end_date", params.EndDate)
	}
	if params.ContractID != "" {
		q.Set("contract_id", params.ContractID)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}

	path := "/rest/v2/reports/payments/detailed"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[DetailedPaymentsReport](resp)
}
