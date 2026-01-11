package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
