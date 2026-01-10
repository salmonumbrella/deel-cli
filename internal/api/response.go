package api

import (
	"encoding/json"
	"fmt"
)

// Page represents cursor pagination metadata.
type Page struct {
	Next  string `json:"next"`
	Total int    `json:"total,omitempty"`
}

// DataResponse wraps a single data payload.
type DataResponse[T any] struct {
	Data T `json:"data"`
}

// ListResponse wraps a list payload with pagination.
type ListResponse[T any] struct {
	Data []T  `json:"data"`
	Page Page `json:"page"`
}

// DataRequest wraps a request body in a data envelope.
type DataRequest[T any] struct {
	Data T `json:"data"`
}

func decodeJSON[T any](raw json.RawMessage) (*T, error) {
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &out, nil
}

func decodeData[T any](raw json.RawMessage) (*T, error) {
	var wrapper DataResponse[T]
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper.Data, nil
}

func decodeList[T any](raw json.RawMessage) (*ListResponse[T], error) {
	var wrapper ListResponse[T]
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &wrapper, nil
}

func wrapData[T any](data T) DataRequest[T] {
	return DataRequest[T]{Data: data}
}
