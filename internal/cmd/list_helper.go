package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

const moreResultsMessage = "More results available. Use --cursor to paginate or --all to fetch everything."

// maxPaginationPages is a safety limit to prevent runaway pagination when using --all.
const maxPaginationPages = 100

// CursorPage captures cursor pagination info.
type CursorPage struct {
	Next  string
	Total int
}

// CursorListResult represents a single paginated response.
type CursorListResult[T any] struct {
	Items []T
	Page  CursorPage
}

// makeListResponse builds a ListResponse from items and page info.
// It clears Page.Next (since --all mode has fetched everything) and preserves Total.
func makeListResponse[T any](items []T, page CursorPage) api.ListResponse[T] {
	return api.ListResponse[T]{
		Data: items,
		Page: api.Page{
			Next:  "", // Always clear - we've collected all items
			Total: page.Total,
		},
	}
}

func outputList[T any](cmd *cobra.Command, f *outfmt.Formatter, items []T, hasMore bool, emptyMessage string, headers []string, rowFunc func(T) []string, response any) error {
	return f.OutputFiltered(cmd.Context(), func() {
		if len(items) == 0 {
			f.PrintText(emptyMessage)
			return
		}
		table := f.NewTable(headers...)
		for _, item := range items {
			table.AddRow(rowFunc(item)...)
		}
		table.Render()
		if hasMore {
			f.PrintText("")
			f.PrintText(moreResultsMessage)
		}
	}, response)
}

func collectCursorItems[T any](
	ctx context.Context,
	all bool,
	cursor string,
	limit int,
	fetch func(ctx context.Context, cursor string, limit int) (CursorListResult[T], error),
) ([]T, CursorPage, bool, error) {
	var (
		items   []T
		page    CursorPage
		hasMore bool
		pages   int
	)

	for {
		result, err := fetch(ctx, cursor, limit)
		if err != nil {
			return nil, CursorPage{}, false, err
		}
		pages++

		if !all {
			items = result.Items
			page = result.Page
			hasMore = result.Page.Next != ""
			break
		}

		items = append(items, result.Items...)
		if result.Page.Total > 0 {
			page.Total = result.Page.Total
		}
		if result.Page.Next == "" {
			break
		}
		if pages >= maxPaginationPages {
			return nil, CursorPage{}, false, fmt.Errorf("pagination safety limit reached (%d pages); use --limit and --cursor for manual pagination", maxPaginationPages)
		}
		cursor = result.Page.Next
	}

	return items, page, hasMore, nil
}
