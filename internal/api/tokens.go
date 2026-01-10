package api

import "context"

// WorkerAccessToken represents a worker access token
type WorkerAccessToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	WorkerID  string `json:"worker_id"`
	Scope     string `json:"scope"`
}

// CreateWorkerAccessTokenParams are params for creating a token
type CreateWorkerAccessTokenParams struct {
	WorkerID string `json:"worker_id"`
	Scope    string `json:"scope,omitempty"`
	TTL      int    `json:"ttl_seconds,omitempty"`
}

// CreateWorkerAccessToken creates a worker access token
func (c *Client) CreateWorkerAccessToken(ctx context.Context, params CreateWorkerAccessTokenParams) (*WorkerAccessToken, error) {
	resp, err := c.Post(ctx, "/rest/v2/tokens/worker-access", params)
	if err != nil {
		return nil, err
	}

	return decodeData[WorkerAccessToken](resp)
}
