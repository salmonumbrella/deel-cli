package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/outfmt"
)

const moreResultsMessage = "More results available. Use --cursor to paginate or --all to fetch everything."

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
	)

	for {
		result, err := fetch(ctx, cursor, limit)
		if err != nil {
			return nil, CursorPage{}, false, err
		}

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
		cursor = result.Page.Next
	}

	return items, page, hasMore, nil
}
