// internal/outfmt/formatter_test.go
package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_OutputWithQuery(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")
	ctx := context.Background()
	ctx = WithQuery(ctx, ".name")

	data := map[string]interface{}{"name": "test", "value": 42}
	err := f.OutputFiltered(ctx, func() {}, data)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "test")
	assert.NotContains(t, buf.String(), "42")
}

func TestFormatter_OutputWithoutQuery(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")
	ctx := context.Background()

	data := map[string]interface{}{"name": "test"}
	err := f.OutputFiltered(ctx, func() {}, data)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "test")
}

func TestFormatter_OutputDataOnly(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")
	ctx := WithDataOnly(context.Background(), true)

	data := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": "c1"},
			map[string]interface{}{"id": "c2"},
		},
		"page": map[string]interface{}{"next": "cursor"},
	}
	err := f.OutputFiltered(ctx, func() {}, data)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "c1")
	assert.NotContains(t, buf.String(), "page")
}

func TestFormatter_OutputFiltered_EnvelopesNonList(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")

	data := map[string]interface{}{"id": "c1"}
	err := f.OutputFiltered(context.Background(), func() {}, data)
	require.NoError(t, err)

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	require.Contains(t, out, "data")
	payload, ok := out["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "c1", payload["id"])
}

func TestFormatter_OutputFiltered_DataOnlySkipsEnvelope(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")
	ctx := WithDataOnly(context.Background(), true)

	data := map[string]interface{}{"id": "c1"}
	err := f.OutputFiltered(ctx, func() {}, data)
	require.NoError(t, err)

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	_, hasData := out["data"]
	assert.False(t, hasData)
	assert.Equal(t, "c1", out["id"])
}

func TestFormatter_OutputFiltered_RawSkipsEnvelope(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")
	f.SetRaw(true)

	data := map[string]interface{}{"id": "c1"}
	err := f.OutputFiltered(context.Background(), func() {}, data)
	require.NoError(t, err)

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	_, hasData := out["data"]
	assert.False(t, hasData)
	assert.Equal(t, "c1", out["id"])
}

func TestFormatter_OutputFiltered_CompactJSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")
	ctx := WithPrettyJSON(context.Background(), false)

	data := map[string]interface{}{"id": "c1", "name": "test"}
	err := f.OutputFiltered(ctx, func() {}, data)
	require.NoError(t, err)

	// Pretty JSON would include indentation/newlines; compact should not.
	assert.NotContains(t, buf.String(), "\n  ")

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	payload, ok := out["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "c1", payload["id"])
	assert.Equal(t, "test", payload["name"])
}

func TestFormatter_OutputFiltered_JSONL(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "auto")

	ctx := context.Background()
	ctx = WithJSONL(ctx, true)
	ctx = WithPrettyJSON(ctx, false)

	data := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": "c1"},
			map[string]interface{}{"id": "c2"},
		},
	}

	err := f.OutputFiltered(ctx, func() {}, data)
	require.NoError(t, err)

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	require.Len(t, lines, 2)

	var out1 map[string]interface{}
	require.NoError(t, json.Unmarshal(lines[0], &out1))
	assert.Equal(t, "c1", out1["id"])

	var out2 map[string]interface{}
	require.NoError(t, json.Unmarshal(lines[1], &out2))
	assert.Equal(t, "c2", out2["id"])
}
