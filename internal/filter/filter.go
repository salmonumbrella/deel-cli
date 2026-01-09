// internal/filter/filter.go
package filter

import (
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
)

// Apply applies a JQ filter expression to the input data.
// The data is first converted to JSON and back to ensure gojq compatibility.
func Apply(data interface{}, expression string) (interface{}, error) {
	if expression == "" {
		return data, nil
	}

	query, err := gojq.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid filter expression: %w", err)
	}

	// Convert Go structs to JSON-compatible map/slice format
	// gojq expects map[string]interface{} / []interface{}, not typed structs
	jsonData, err := toJSONCompatible(data)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare data for filter: %w", err)
	}

	iter := query.Run(jsonData)

	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("filter error: %w", err)
		}
		results = append(results, v)
	}

	// Return single result unwrapped, multiple as array
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

// toJSONCompatible converts Go structs to map[string]interface{} / []interface{}
// by marshaling to JSON and unmarshaling back. This ensures gojq can traverse the data.
func toJSONCompatible(data interface{}) (interface{}, error) {
	// Marshal to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Unmarshal to generic interface{}
	var result interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}
