// internal/filter/filter.go
package filter

import (
	"fmt"

	"github.com/itchyny/gojq"
)

// Apply applies a JQ filter expression to the input data.
func Apply(data interface{}, expression string) (interface{}, error) {
	if expression == "" {
		return data, nil
	}

	query, err := gojq.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid filter expression: %w", err)
	}

	iter := query.Run(data)

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
