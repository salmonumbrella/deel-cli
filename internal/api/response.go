package api

import (
	"encoding/json"
	"fmt"
	"strconv"
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

// FlexFloat64 handles JSON number fields that may be strings or numbers.
type FlexFloat64 float64

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexFloat64) UnmarshalJSON(data []byte) error {
	// Try as number first
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = FlexFloat64(num)
		return nil
	}

	// Try as string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			*f = 0
			return nil
		}
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as float64: %w", str, err)
		}
		*f = FlexFloat64(parsed)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexFloat64", string(data))
}
