package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
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

// ListTasks returns tasks
func (c *Client) ListTasks(ctx context.Context, params TasksListParams) (*TasksListResponse, error) {
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

	path := "/rest/v2/tasks"
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
	ContractID  string  `json:"contract_id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount"`
}

// CreateTask creates a new task
func (c *Client) CreateTask(ctx context.Context, params CreateTaskParams) (*Task, error) {
	resp, err := c.Post(ctx, "/rest/v2/tasks", params)
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
func (c *Client) UpdateTask(ctx context.Context, taskID string, params UpdateTaskParams) (*Task, error) {
	path := fmt.Sprintf("/rest/v2/tasks/%s", escapePath(taskID))
	resp, err := c.Patch(ctx, path, params)
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
func (c *Client) DeleteTask(ctx context.Context, taskID string) error {
	path := fmt.Sprintf("/rest/v2/tasks/%s", escapePath(taskID))
	_, err := c.Delete(ctx, path)
	return err
}

// ReviewTask approves or rejects a task
func (c *Client) ReviewTask(ctx context.Context, taskID string, status string) error {
	path := fmt.Sprintf("/rest/v2/tasks/%s/review", escapePath(taskID))
	_, err := c.Post(ctx, path, map[string]string{"status": status})
	return err
}

// ReviewMultipleTasks reviews multiple tasks at once
func (c *Client) ReviewMultipleTasks(ctx context.Context, taskIDs []string, status string) error {
	_, err := c.Post(ctx, "/rest/v2/tasks/review-many", map[string]any{
		"task_ids": taskIDs,
		"status":   status,
	})
	return err
}
