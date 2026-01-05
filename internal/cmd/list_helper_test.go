// internal/cmd/list_helper_test.go
package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

type testItem struct {
	ID   string
	Name string
}

func TestListCommand_TextOutput(t *testing.T) {
	cfg := ListConfig[testItem]{
		Use:          "list",
		Short:        "List items",
		Headers:      []string{"ID", "NAME"},
		RowFunc:      func(item testItem) []string { return []string{item.ID, item.Name} },
		EmptyMessage: "No items found",
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
			return ListResult[testItem]{
				Items:   []testItem{{ID: "1", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func() (*api.Client, error) {
		return api.NewClient("test-token"), nil
	})

	require.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
}

func TestListCommand_HasExpectedFlags(t *testing.T) {
	cfg := ListConfig[testItem]{
		Use:          "list",
		Short:        "List items",
		Headers:      []string{"ID", "NAME"},
		RowFunc:      func(item testItem) []string { return []string{item.ID, item.Name} },
		EmptyMessage: "No items found",
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
			return ListResult[testItem]{}, nil
		},
	}

	cmd := NewListCommand(cfg, func() (*api.Client, error) {
		return api.NewClient("test-token"), nil
	})

	// Verify pagination flags exist
	pageFlag := cmd.Flags().Lookup("page")
	require.NotNil(t, pageFlag, "expected --page flag")
	assert.Equal(t, "0", pageFlag.DefValue)

	limitFlag := cmd.Flags().Lookup("limit")
	require.NotNil(t, limitFlag, "expected --limit flag")
	assert.Equal(t, "100", limitFlag.DefValue)
}

func TestListCommand_WithLongDescription(t *testing.T) {
	cfg := ListConfig[testItem]{
		Use:          "list",
		Short:        "List items",
		Long:         "This is a longer description for the list command.",
		Example:      "  deel items list --limit 10",
		Headers:      []string{"ID", "NAME"},
		RowFunc:      func(item testItem) []string { return []string{item.ID, item.Name} },
		EmptyMessage: "No items found",
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
			return ListResult[testItem]{}, nil
		},
	}

	cmd := NewListCommand(cfg, func() (*api.Client, error) {
		return api.NewClient("test-token"), nil
	})

	assert.Equal(t, "This is a longer description for the list command.", cmd.Long)
	assert.Equal(t, "  deel items list --limit 10", cmd.Example)
}
