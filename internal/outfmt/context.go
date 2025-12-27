// internal/outfmt/context.go
package outfmt

import "context"

type contextKey string

const (
	formatKey contextKey = "output_format"
	queryKey  contextKey = "query_filter"
)

// WithFormat returns a context with the output format set.
func WithFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, formatKey, format)
}

// GetFormat returns the output format from context, defaulting to "text".
func GetFormat(ctx context.Context) string {
	if v, ok := ctx.Value(formatKey).(string); ok {
		return v
	}
	return "text"
}

// IsJSON returns true if the output format is JSON.
func IsJSON(ctx context.Context) bool {
	return GetFormat(ctx) == "json"
}

// WithQuery returns a context with the JQ query filter set.
func WithQuery(ctx context.Context, query string) context.Context {
	return context.WithValue(ctx, queryKey, query)
}

// GetQuery returns the JQ query from context.
func GetQuery(ctx context.Context) string {
	if v, ok := ctx.Value(queryKey).(string); ok {
		return v
	}
	return ""
}
