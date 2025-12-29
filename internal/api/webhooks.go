package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Webhook represents a webhook subscription
type Webhook struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Secret      string   `json:"secret,omitempty"`
	Status      string   `json:"status"`
	Description string   `json:"description,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

// ListWebhooks returns all webhooks
func (c *Client) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	resp, err := c.Get(ctx, "/rest/v2/webhooks")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []Webhook `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}

// GetWebhook returns a single webhook by ID
func (c *Client) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	path := fmt.Sprintf("/rest/v2/webhooks/%s", escapePath(id))
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Webhook `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// CreateWebhookParams are params for creating a webhook
type CreateWebhookParams struct {
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Description string   `json:"description,omitempty"`
}

// CreateWebhook creates a new webhook
func (c *Client) CreateWebhook(ctx context.Context, params CreateWebhookParams) (*Webhook, error) {
	resp, err := c.Post(ctx, "/rest/v2/webhooks", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Webhook `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// UpdateWebhookParams are params for updating a webhook
type UpdateWebhookParams struct {
	URL         string   `json:"url,omitempty"`
	Events      []string `json:"events,omitempty"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// UpdateWebhook updates an existing webhook
func (c *Client) UpdateWebhook(ctx context.Context, id string, params UpdateWebhookParams) (*Webhook, error) {
	path := fmt.Sprintf("/rest/v2/webhooks/%s", escapePath(id))
	resp, err := c.Patch(ctx, path, params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data Webhook `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

// DeleteWebhook deletes a webhook
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	path := fmt.Sprintf("/rest/v2/webhooks/%s", escapePath(id))
	_, err := c.Delete(ctx, path)
	return err
}

// WebhookEventType represents a webhook event type
type WebhookEventType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListWebhookEventTypes returns available webhook event types
func (c *Client) ListWebhookEventTypes(ctx context.Context) ([]WebhookEventType, error) {
	resp, err := c.Get(ctx, "/rest/v2/webhooks/event-types")
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []WebhookEventType `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return wrapper.Data, nil
}
