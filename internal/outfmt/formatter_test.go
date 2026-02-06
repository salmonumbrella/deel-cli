// internal/outfmt/formatter_test.go
package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deel-cli/internal/dryrun"
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

func TestFormatter_OutputFiltered_TextMode(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatText, "never")

	called := false
	err := f.OutputFiltered(context.Background(), func() {
		called = true
		f.PrintText("hello text")
	}, map[string]interface{}{"id": "c1"})
	require.NoError(t, err)
	assert.True(t, called, "text callback should be invoked in text mode")
	assert.Contains(t, buf.String(), "hello text")
}

func TestEnsureEnvelope_NilData(t *testing.T) {
	result := ensureEnvelope(nil)
	m, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Nil(t, m["data"])
}

func TestEnsureEnvelope_AlreadyWrapped(t *testing.T) {
	data := map[string]any{"data": []any{"a", "b"}}
	result := ensureEnvelope(data)
	// Should return the same map, not double-wrap
	m, ok := result.(map[string]any)
	require.True(t, ok)
	assert.NotNil(t, m["data"])
	arr, ok := m["data"].([]any)
	require.True(t, ok)
	assert.Len(t, arr, 2)
}

func TestExtractData_MapWithDataKey(t *testing.T) {
	data := map[string]any{"data": "payload", "extra": "ignored"}
	extracted, ok := extractData(data)
	assert.True(t, ok)
	assert.Equal(t, "payload", extracted)
}

func TestExtractData_MapWithoutDataKey(t *testing.T) {
	data := map[string]any{"id": "123"}
	extracted, ok := extractData(data)
	assert.False(t, ok)
	assert.Equal(t, data, extracted)
}

func TestExtractData_Struct(t *testing.T) {
	type wrap struct {
		Data string
	}
	data := wrap{Data: "payload"}
	extracted, ok := extractData(data)
	assert.True(t, ok)
	assert.Equal(t, "payload", extracted)
}

func TestExtractData_Nil(t *testing.T) {
	extracted, ok := extractData(nil)
	assert.False(t, ok)
	assert.Nil(t, extracted)
}

func TestFormatter_PrintDryRun_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatJSON, "never")

	err := f.PrintDryRun(&dryrun.Preview{
		Operation:   "CREATE",
		Resource:    "Person",
		Description: "Create person",
		Details:     map[string]string{"Name": "Alice"},
	})
	require.NoError(t, err)

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, true, out["dry_run"])
	assert.NotNil(t, out["preview"])
}

func TestFormatter_PrintDryRun_Text(t *testing.T) {
	var buf bytes.Buffer
	f := New(&buf, &buf, FormatText, "never")

	err := f.PrintDryRun(&dryrun.Preview{
		Operation:   "DELETE",
		Resource:    "Person",
		Description: "Delete person",
		Details:     map[string]string{"ID": "123"},
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "DELETE")
	assert.Contains(t, buf.String(), "Person")
}
