// internal/outfmt/context_test.go
package outfmt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithFormat(t *testing.T) {
	ctx := context.Background()
	ctx = WithFormat(ctx, "json")
	assert.Equal(t, "json", GetFormat(ctx))
}

func TestIsJSON(t *testing.T) {
	ctx := context.Background()
	assert.False(t, IsJSON(ctx))

	ctx = WithFormat(ctx, "json")
	assert.True(t, IsJSON(ctx))
}

func TestWithQuery(t *testing.T) {
	ctx := context.Background()
	ctx = WithQuery(ctx, ".data[]")
	assert.Equal(t, ".data[]", GetQuery(ctx))
}

func TestGetFormat_Default(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "text", GetFormat(ctx))
}
