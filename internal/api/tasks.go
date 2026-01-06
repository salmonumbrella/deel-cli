package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Task represents a contractor task
type Task struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	ContractID  string  `json:"contract_id"`
	CreatedAt   string  `json:"created_at"`
}

// UnmarshalJSON implements custom unmarshaling to handle amount as string from API
func (t *Task) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type TaskAlias Task
	aux := &struct {
		Amount json.RawMessage `json:"amount"`
		*TaskAlias
	}{
		TaskAlias: (*TaskAlias)(t),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle amount as either string or number
	if len(aux.Amount) > 0 {
		// Try to unmarshal as string first (API returns string)
		var amountStr string
		if err := json.Unmarshal(aux.Amount, &amountStr); err == nil {
			if parsed, err := strconv.ParseFloat(amountStr, 64); err == nil {
				t.Amount = parsed
			}
		} else {
			// Fall back to float64 if it's already a number
			var amountFloat float64
			if err := json.Unmarshal(aux.Amount, &amountFloat); err == nil {
				t.Amount = amountFloat
			}
		}
	}

	return nil
}

// TasksListParams are params for listing tasks
type TasksListParams struct {
	ContractID string
	Status     string
	Limit      int
	Cursor     string
}

// TasksListResponse is the response from list tasks
type TasksListResponse struct {
	Data []Task `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"page"`
}

// ListTasks returns tasks for a contract
// ContractID is required - tasks are nested under contracts in the API
func (c *Client) ListTasks(ctx context.Context, params TasksListParams) (*TasksListResponse, error) {
	if params.ContractID == "" {
		return nil, fmt.Errorf("contract_id is required")
	}

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

	// Tasks are nested under contracts: /rest/v2/contracts/{contract_id}/tasks
	path := fmt.Sprintf("/rest/v2/contracts/%s/tasks", escapePath(params.ContractID))
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result TasksListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// CreateTaskParams are params for creating a task
type CreateTaskParams struct {
	ContractID    string  `json:"-"` // Used for URL path, not request body
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	Amount        float64 `json:"amount"`
	DateSubmitted string  `json:"date_submitted,omitempty"` // YYYY-MM-DD format, defaults to today
}

// createTaskRequest is the request body for creating a task (wrapped in data)
type createTaskRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	Amount        string `json:"amount"` // Deel API requires amount as string
	DateSubmitted string `json:"date_submitted"`
}

// CreateTask creates a new task for a contract
// Tasks are nested under contracts: POST /rest/v2/contracts/{contract_id}/tasks
func (c *Client) CreateTask(ctx context.Context, params CreateTaskParams) (*Task, error) {
	if params.ContractID == "" {
		return nil, fmt.Errorf("contract_id is required")
	}

	path := fmt.Sprintf("/rest/v2/contracts/%s/tasks", escapePath(params.ContractID))

	// Default date_submitted to today if not provided
	dateSubmitted := params.DateSubmitted
	if dateSubmitted == "" {
		dateSubmitted = time.Now().Format("2006-01-02")
	}

	// Wrap request in data object as required by Deel API
	reqBody := struct {
		Data createTaskRequest `json:"data"`
	}{
		Data: createTaskRequest{
			Title:         params.Title,
			Description:   params.Description,
			Amount:        fmt.Sprintf("%.2f", params.Amount), // API requires string
			DateSubmitted: dateSubmitted,
		},
	}

	resp, err := c.Post(ctx, path, reqBody)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Task `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateTaskParams are params for updating a task
type UpdateTaskParams struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
}

// UpdateTask updates an existing task
// Tasks are nested under contracts: PATCH /rest/v2/contracts/{contract_id}/tasks/{task_id}
func (c *Client) UpdateTask(ctx context.Context, contractID, taskID string, params UpdateTaskParams) (*Task, error) {
	if contractID == "" {
		return nil, fmt.Errorf("contract_id is required")
	}
	path := fmt.Sprintf("/rest/v2/contracts/%s/tasks/%s", escapePath(contractID), escapePath(taskID))

	// Wrap request in data object as required by Deel API
	reqBody := struct {
		Data UpdateTaskParams `json:"data"`
	}{Data: params}

	resp, err := c.Patch(ctx, path, reqBody)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Task `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteTask deletes a task
// Tasks are nested under contracts: DELETE /rest/v2/contracts/{contract_id}/tasks/{task_id}
func (c *Client) DeleteTask(ctx context.Context, contractID, taskID string) error {
	if contractID == "" {
		return fmt.Errorf("contract_id is required")
	}
	path := fmt.Sprintf("/rest/v2/contracts/%s/tasks/%s", escapePath(contractID), escapePath(taskID))
	_, err := c.Delete(ctx, path)
	return err
}

// ReviewTask approves or rejects a task
// Tasks are nested under contracts: POST /rest/v2/contracts/{contract_id}/tasks/{task_id}/reviews
func (c *Client) ReviewTask(ctx context.Context, contractID, taskID string, status string) error {
	if contractID == "" {
		return fmt.Errorf("contract_id is required")
	}
	path := fmt.Sprintf("/rest/v2/contracts/%s/tasks/%s/reviews", escapePath(contractID), escapePath(taskID))

	// Wrap request in data object as required by Deel API
	reqBody := struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}{}
	reqBody.Data.Status = status

	_, err := c.Post(ctx, path, reqBody)
	return err
}

// ReviewMultipleTasks reviews multiple tasks at once
// Tasks are nested under contracts: POST /rest/v2/contracts/{contract_id}/tasks/many/reviews
func (c *Client) ReviewMultipleTasks(ctx context.Context, contractID string, taskIDs []string, status string) error {
	if contractID == "" {
		return fmt.Errorf("contract_id is required")
	}
	path := fmt.Sprintf("/rest/v2/contracts/%s/tasks/many/reviews", escapePath(contractID))

	// Wrap request in data object as required by Deel API
	reqBody := struct {
		Data struct {
			TaskIDs []string `json:"task_ids"`
			Status  string   `json:"status"`
		} `json:"data"`
	}{}
	reqBody.Data.TaskIDs = taskIDs
	reqBody.Data.Status = status

	_, err := c.Post(ctx, path, reqBody)
	return err
}
