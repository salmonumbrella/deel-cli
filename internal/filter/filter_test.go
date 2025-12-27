// internal/filter/filter_test.go
package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply_SimpleField(t *testing.T) {
	data := map[string]interface{}{"name": "test", "value": 42}
	result, err := Apply(data, ".name")
	require.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestApply_Array(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}
	result, err := Apply(data, ".items[]")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"a", "b", "c"}, result)
}

func TestApply_EmptyExpression(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	result, err := Apply(data, "")
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestApply_InvalidExpression(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	_, err := Apply(data, ".[invalid")
	require.Error(t, err)
}
