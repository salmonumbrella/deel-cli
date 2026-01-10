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

func TestCollectCursorItems_SinglePage(t *testing.T) {
	ctx := context.Background()
	items, page, hasMore, err := collectCursorItems(ctx, false, "", 100, func(ctx context.Context, cursor string, limit int) (CursorListResult[testItem], error) {
		return CursorListResult[testItem]{
			Items: []testItem{{ID: "1", Name: "One"}},
			Page: CursorPage{
				Next: "next-token",
			},
		}, nil
	})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "next-token", page.Next)
	assert.True(t, hasMore)
}

func TestCollectCursorItems_AllPages(t *testing.T) {
	ctx := context.Background()
	calls := 0
	items, page, hasMore, err := collectCursorItems(ctx, true, "", 100, func(ctx context.Context, cursor string, limit int) (CursorListResult[testItem], error) {
		calls++
		if calls == 1 {
			return CursorListResult[testItem]{
				Items: []testItem{{ID: "1", Name: "One"}},
				Page: CursorPage{
					Next: "page-2",
				},
			}, nil
		}
		return CursorListResult[testItem]{
			Items: []testItem{{ID: "2", Name: "Two"}},
			Page: CursorPage{
				Next: "",
			},
		}, nil
	})
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "", page.Next)
	assert.False(t, hasMore)
}

func TestNewCursorListCommand_Flags(t *testing.T) {
	cmd := NewCursorListCommand(CursorListConfig[testItem]{
		Use:          "list",
		Short:        "List items",
		Operation:    "listing items",
		DefaultLimit: 50,
		SupportsAll:  true,
		Headers:      []string{"ID", "NAME"},
		RowFunc:      func(item testItem) []string { return []string{item.ID, item.Name} },
		EmptyMessage: "No items found",
		Fetch: func(ctx context.Context, client *api.Client, cursor string, limit int) (CursorListResult[testItem], error) {
			return CursorListResult[testItem]{}, nil
		},
	})

	require.NotNil(t, cmd)
	limitFlag := cmd.Flags().Lookup("limit")
	require.NotNil(t, limitFlag)
	assert.Equal(t, "50", limitFlag.DefValue)

	cursorFlag := cmd.Flags().Lookup("cursor")
	require.NotNil(t, cursorFlag)

	allFlag := cmd.Flags().Lookup("all")
	require.NotNil(t, allFlag)
}
