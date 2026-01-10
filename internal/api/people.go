package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Department can be a string or an object from the API
type Department struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Employment represents an employment record for a person
type Employment struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ContractStatus string `json:"contract_status"`
	HiringType     string `json:"hiring_type"` // contractor, eor, gp, direct
	HiringStatus   string `json:"hiring_status"`
	JobTitle       string `json:"job_title"`
	Country        string `json:"country"`
	IsEnded        bool   `json:"is_ended"`
}

// Person represents a Deel person/worker
type Person struct {
	ID            string       `json:"id"`
	HRISProfileID string       `json:"hris_profile_id"`
	FirstName     string       `json:"first_name"`
	LastName      string       `json:"last_name"`
	Name          string       `json:"name"` // Computed: FirstName + LastName
	Email         string       `json:"email"`
	JobTitle      string       `json:"job_title"`
	DepartmentRaw any          `json:"department"` // API returns string or object
	Status        string       `json:"status"`
	StartDate     string       `json:"start_date"`
	Country       string       `json:"country"`
	HiringType    string       `json:"hiring_type,omitempty"`
	Employments   []Employment `json:"employments,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling to compute the Name field
func (p *Person) UnmarshalJSON(data []byte) error {
	type Alias Person
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Compute Name from FirstName and LastName
	p.Name = p.FirstName
	if p.LastName != "" {
		if p.Name != "" {
			p.Name += " "
		}
		p.Name += p.LastName
	}
	return nil
}

// Department returns the department name, handling both string and object formats
func (p *Person) Department() string {
	if p.DepartmentRaw == nil {
		return ""
	}
	// Handle string case
	if s, ok := p.DepartmentRaw.(string); ok {
		return s
	}
	// Handle object case
	if m, ok := p.DepartmentRaw.(map[string]any); ok {
		if name, ok := m["name"].(string); ok {
			return name
		}
	}
	return ""
}

// PeopleListResponse is the response from list people
type PeopleListResponse = ListResponse[Person]

// PeopleListParams are params for listing people
type PeopleListParams struct {
	Limit  int
	Cursor string
}

// ListPeople returns a list of people
func (c *Client) ListPeople(ctx context.Context, params PeopleListParams) (*PeopleListResponse, error) {
	q := url.Values{}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/rest/v2/people"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeList[Person](resp)
}

// GetPerson returns a single person by HRIS profile ID
func (c *Client) GetPerson(ctx context.Context, hrisProfileID string) (*Person, error) {
	path := fmt.Sprintf("/rest/v2/people/%s", escapePath(hrisProfileID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[Person](resp)
}

// GetPersonPersonal returns personal info including numeric worker_id
// Returns raw JSON to allow flexible handling of varied response shapes
func (c *Client) GetPersonPersonal(ctx context.Context, hrisProfileID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/rest/v2/people/%s/personal", escapePath(hrisProfileID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	data, err := decodeData[json.RawMessage](resp)
	if err != nil {
		return nil, err
	}
	return *data, nil
}

// SearchPeopleByEmail finds a person by email
func (c *Client) SearchPeopleByEmail(ctx context.Context, email string) (*Person, error) {
	q := url.Values{}
	q.Set("email", email)
	path := "/rest/v2/people/search?" + q.Encode()

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[Person](resp)
}

// CustomField represents a custom field
type CustomField struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

// ListCustomFields returns custom fields for people
func (c *Client) ListCustomFields(ctx context.Context) ([]CustomField, error) {
	resp, err := c.Get(ctx, "/rest/v2/people/custom-fields")
	if err != nil {
		return nil, err
	}

	fields, err := decodeData[[]CustomField](resp)
	if err != nil {
		return nil, err
	}
	return *fields, nil
}

// GetCustomField returns a specific custom field
func (c *Client) GetCustomField(ctx context.Context, fieldID string) (*CustomField, error) {
	path := fmt.Sprintf("/rest/v2/people/custom-fields/%s", escapePath(fieldID))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeData[CustomField](resp)
}

// DepartmentsListResponse is the response from list departments
type DepartmentsListResponse = ListResponse[Department]

// ListDepartments returns a list of departments in the organization
func (c *Client) ListDepartments(ctx context.Context) (*DepartmentsListResponse, error) {
	resp, err := c.Get(ctx, "/rest/v2/departments")
	if err != nil {
		return nil, err
	}

	return decodeList[Department](resp)
}

// UpdatePersonDepartmentParams contains parameters for updating a person's department
type UpdatePersonDepartmentParams struct {
	DepartmentID string `json:"department_id"`
}

// UpdatePersonDepartment updates the department for a person
// PUT /rest/v2/people/{id}/department
func (c *Client) UpdatePersonDepartment(ctx context.Context, personID string, params UpdatePersonDepartmentParams) (*Department, error) {
	path := fmt.Sprintf("/rest/v2/people/%s/department", escapePath(personID))

	resp, err := c.Put(ctx, path, wrapData(params))
	if err != nil {
		return nil, err
	}

	return decodeData[Department](resp)
}
