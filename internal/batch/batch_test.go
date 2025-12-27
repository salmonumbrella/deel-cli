package batch

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSON_Array(t *testing.T) {
	input := `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`
	reader := strings.NewReader(input)

	items, err := parseJSON(reader)
	require.NoError(t, err)
	require.Len(t, items, 2)

	// Verify first item
	var first map[string]any
	err = items[0].Unmarshal(&first)
	require.NoError(t, err)
	assert.Equal(t, float64(1), first["id"])
	assert.Equal(t, "Alice", first["name"])

	// Verify second item
	var second map[string]any
	err = items[1].Unmarshal(&second)
	require.NoError(t, err)
	assert.Equal(t, float64(2), second["id"])
	assert.Equal(t, "Bob", second["name"])
}

func TestParseJSON_NDJSON(t *testing.T) {
	input := `{"id": 1, "name": "Alice"}
{"id": 2, "name": "Bob"}
{"id": 3, "name": "Charlie"}`
	reader := strings.NewReader(input)

	items, err := parseJSON(reader)
	require.NoError(t, err)
	require.Len(t, items, 3)

	// Verify items
	for i, expected := range []string{"Alice", "Bob", "Charlie"} {
		var item map[string]any
		err = items[i].Unmarshal(&item)
		require.NoError(t, err)
		assert.Equal(t, expected, item["name"])
	}
}

func TestParseJSON_Empty(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"empty array", "[]"},
		{"whitespace only", "   \n\t  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			items, err := parseJSON(reader)
			require.NoError(t, err)
			assert.Empty(t, items)
		})
	}
}

func TestParseJSON_TooManyItems(t *testing.T) {
	// Create input with more than MaxItemCount items
	var buf bytes.Buffer
	buf.WriteString("[")
	for i := 0; i < MaxItemCount+1; i++ {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(`{"id":1}`)
	}
	buf.WriteString("]")

	reader := &buf
	_, err := parseJSON(reader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many items")
}

func TestParseJSON_InvalidJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid json", `{invalid}`},
		{"unclosed array", `[{"id": 1}`},
		{"non-object in array", `[1, 2, 3]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			_, err := parseJSON(reader)
			require.Error(t, err)
		})
	}
}

func TestResult(t *testing.T) {
	// Test success result
	success := Result{
		Index:   0,
		Success: true,
		Data:    map[string]string{"id": "123"},
	}
	assert.True(t, success.Success)
	assert.Nil(t, success.Error)

	// Test error result
	failure := Result{
		Index:   1,
		Success: false,
		Error:   assert.AnError,
	}
	assert.False(t, failure.Success)
	assert.NotNil(t, failure.Error)
}

func TestSummary(t *testing.T) {
	summary := Summary{
		Total:     10,
		Succeeded: 8,
		Failed:    2,
	}

	assert.Equal(t, 10, summary.Total)
	assert.Equal(t, 8, summary.Succeeded)
	assert.Equal(t, 2, summary.Failed)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 10*1024*1024, MaxInputSize) // 10MB
	assert.Equal(t, 10000, MaxItemCount)
}
