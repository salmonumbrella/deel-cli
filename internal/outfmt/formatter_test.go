// internal/outfmt/formatter_test.go
package outfmt

import (
	"bytes"
	"context"
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
