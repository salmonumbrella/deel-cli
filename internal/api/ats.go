package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ATSJob represents an ATS job
type ATSJob struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Department     string `json:"department"`
	DepartmentID   string `json:"department_id"`
	Location       string `json:"location"`
	LocationID     string `json:"location_id"`
	Status         string `json:"status"`
	EmploymentType string `json:"employment_type"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ATSJobPosting represents a public job posting
type ATSJobPosting struct {
	ID             string `json:"id"`
	JobID          string `json:"job_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Department     string `json:"department"`
	Location       string `json:"location"`
	EmploymentType string `json:"employment_type"`
	Status         string `json:"status"`
	PostedAt       string `json:"posted_at"`
	ClosedAt       string `json:"closed_at,omitempty"`
	URL            string `json:"url"`
}

// ATSApplication represents a job application
type ATSApplication struct {
	ID            string `json:"id"`
	CandidateID   string `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	JobID         string `json:"job_id"`
	JobTitle      string `json:"job_title"`
	Status        string `json:"status"`
	Stage         string `json:"stage"`
	AppliedAt     string `json:"applied_at"`
	UpdatedAt     string `json:"updated_at"`
}

// ATSCandidate represents a job candidate
type ATSCandidate struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Location    string `json:"location"`
	LinkedInURL string `json:"linkedin_url,omitempty"`
	ResumeURL   string `json:"resume_url,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// ATSDepartment represents an organizational department
type ATSDepartment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ParentID  string `json:"parent_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

// ATSLocation represents a job location
type ATSLocation struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Remote    bool   `json:"remote"`
	CreatedAt string `json:"created_at"`
}

// RejectionReason represents a candidate rejection reason
type RejectionReason struct {
	ID          string `json:"id"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

// ATSOffer represents an ATS offer
type ATSOffer struct {
	ID          string  `json:"id"`
	CandidateID string  `json:"candidate_id"`
	Candidate   string  `json:"candidate_name"`
	Position    string  `json:"position"`
	Status      string  `json:"status"`
	Salary      float64 `json:"salary"`
	Currency    string  `json:"currency"`
	StartDate   string  `json:"start_date"`
	CreatedAt   string  `json:"created_at"`
}

// ATSJobsListParams are params for listing jobs
type ATSJobsListParams struct {
	Status       string
	DepartmentID string
	LocationID   string
	Limit        int
	Cursor       string
}

// CreateATSJobParams are params for creating a job
type CreateATSJobParams struct {
	Title          string `json:"title"`
	DepartmentID   string `json:"department_id"`
	LocationID     string `json:"location_id"`
	EmploymentType string `json:"employment_type"`
	Description    string `json:"description,omitempty"`
}

// ATSJobPostingsListParams are params for listing job postings
type ATSJobPostingsListParams struct {
	Status string
	JobID  string
	Limit  int
	Cursor string
}

// ATSApplicationsListParams are params for listing applications
type ATSApplicationsListParams struct {
	Status      string
	JobID       string
	CandidateID string
	Stage       string
	Limit       int
	Cursor      string
}

// ATSCandidatesListParams are params for listing candidates
type ATSCandidatesListParams struct {
	Search string
	Limit  int
	Cursor string
}

// ATSDepartmentsListParams are params for listing departments
type ATSDepartmentsListParams struct {
	Limit  int
	Cursor string
}

// ATSLocationsListParams are params for listing locations
type ATSLocationsListParams struct {
	Remote *bool
	Limit  int
	Cursor string
}

// ATSOffersListParams are params for listing offers
type ATSOffersListParams struct {
	Status string
	Limit  int
	Cursor string
}

// ATSOffersListResponse is the response from list offers
type ATSOffersListResponse = ListResponse[ATSOffer]

// ATSJobsListResponse is the response from list jobs
type ATSJobsListResponse = ListResponse[ATSJob]

// ATSJobPostingsListResponse is the response from list job postings
type ATSJobPostingsListResponse = ListResponse[ATSJobPosting]

// ATSApplicationsListResponse is the response from list applications
type ATSApplicationsListResponse = ListResponse[ATSApplication]

// ATSCandidatesListResponse is the response from list candidates
type ATSCandidatesListResponse = ListResponse[ATSCandidate]

// ATSDepartmentsListResponse is the response from list departments
type ATSDepartmentsListResponse = ListResponse[ATSDepartment]

// ATSLocationsListResponse is the response from list locations
type ATSLocationsListResponse = ListResponse[ATSLocation]

// ListATSOffers returns ATS offers
func (c *Client) ListATSOffers(ctx context.Context, params ATSOffersListParams) (*ATSOffersListResponse, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/ats/offers"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ATSOffer](resp)
}

// ListATSJobs returns ATS jobs
func (c *Client) ListATSJobs(ctx context.Context, params ATSJobsListParams) (*ATSJobsListResponse, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.DepartmentID != "" {
		q.Set("department_id", params.DepartmentID)
	}
	if params.LocationID != "" {
		q.Set("location_id", params.LocationID)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/ats/jobs"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ATSJob](resp)
}

// CreateATSJob creates a new ATS job
func (c *Client) CreateATSJob(ctx context.Context, params CreateATSJobParams) (*ATSJob, error) {
	resp, err := c.Post(ctx, "/rest/v2/ats/jobs", params)
	if err != nil {
		return nil, err
	}

	return decodeData[ATSJob](resp)
}

// ListATSJobPostings returns ATS job postings
func (c *Client) ListATSJobPostings(ctx context.Context, params ATSJobPostingsListParams) (*ATSJobPostingsListResponse, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.JobID != "" {
		q.Set("job_id", params.JobID)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/ats/job-postings"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ATSJobPosting](resp)
}

// GetATSJobPosting returns a single job posting by ID
func (c *Client) GetATSJobPosting(ctx context.Context, id string) (*ATSJobPosting, error) {
	path := fmt.Sprintf("/rest/v2/ats/job-postings/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[ATSJobPosting](resp)
}

// ListATSApplications returns ATS applications
func (c *Client) ListATSApplications(ctx context.Context, params ATSApplicationsListParams) (*ATSApplicationsListResponse, error) {
	q := url.Values{}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.JobID != "" {
		q.Set("job_id", params.JobID)
	}
	if params.CandidateID != "" {
		q.Set("candidate_id", params.CandidateID)
	}
	if params.Stage != "" {
		q.Set("stage", params.Stage)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/ats/applications"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ATSApplication](resp)
}

// ListATSCandidates returns ATS candidates
func (c *Client) ListATSCandidates(ctx context.Context, params ATSCandidatesListParams) (*ATSCandidatesListResponse, error) {
	q := url.Values{}
	if params.Search != "" {
		q.Set("search", params.Search)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/ats/candidates"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ATSCandidate](resp)
}

// ListATSDepartments returns ATS departments
func (c *Client) ListATSDepartments(ctx context.Context, params ATSDepartmentsListParams) (*ATSDepartmentsListResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/ats/departments"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ATSDepartment](resp)
}

// ListATSLocations returns ATS locations
func (c *Client) ListATSLocations(ctx context.Context, params ATSLocationsListParams) (*ATSLocationsListResponse, error) {
	q := url.Values{}
	if params.Remote != nil {
		q.Set("remote", strconv.FormatBool(*params.Remote))
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/ats/locations"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[ATSLocation](resp)
}

// ListRejectionReasons returns rejection reasons
func (c *Client) ListRejectionReasons(ctx context.Context) ([]RejectionReason, error) {
	resp, err := c.Get(ctx, "/rest/v2/ats/rejection-reasons")
	if err != nil {
		return nil, err
	}

	reasons, err := decodeData[[]RejectionReason](resp)
	if err != nil {
		return nil, err
	}
	return *reasons, nil
}
