package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
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

// CursorListConfig defines how a list command behaves.
type CursorListConfig[T any] struct {
	Use       string
	Short     string
	Long      string
	Example   string
	Operation string

	DefaultLimit int
	SupportsAll  bool

	Fetch func(ctx context.Context, client *api.Client, cursor string, limit int) (CursorListResult[T], error)

	Headers      []string
	RowFunc      func(T) []string
	EmptyMessage string

	// BuildResponse constructs the JSON payload for OutputFiltered.
	// If nil, a generic response is used.
	BuildResponse func(items []T, page CursorPage) any
}

// NewCursorListCommand creates a cobra command from CursorListConfig.
func NewCursorListCommand[T any](cfg CursorListConfig[T]) *cobra.Command {
	var cursor string
	var limit int
	var all bool

	defaultLimit := cfg.DefaultLimit
	if defaultLimit <= 0 {
		defaultLimit = 100
	}

	cmd := &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, f, err := initClient(cfg.Operation)
			if err != nil {
				return err
			}

			if limit <= 0 {
				limit = defaultLimit
			}

			items, page, hasMore, err := collectCursorItems(cmd.Context(), all, cursor, limit, func(ctx context.Context, cur string, lim int) (CursorListResult[T], error) {
				return cfg.Fetch(ctx, client, cur, lim)
			})
			if err != nil {
				return HandleError(f, err, cfg.Operation)
			}

			response := buildCursorResponse(cfg, items, page)
			return outputList(cmd, f, items, hasMore, cfg.EmptyMessage, cfg.Headers, cfg.RowFunc, response)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", defaultLimit, "Maximum results")
	cmd.Flags().StringVar(&cursor, "cursor", "", "Pagination cursor")
	if cfg.SupportsAll {
		cmd.Flags().BoolVar(&all, "all", false, "Fetch all results")
	}
	return cmd
}

func buildCursorResponse[T any](cfg CursorListConfig[T], items []T, page CursorPage) any {
	if cfg.BuildResponse != nil {
		return cfg.BuildResponse(items, page)
	}
	return struct {
		Data []T `json:"data"`
		Page struct {
			Next  string `json:"next"`
			Total int    `json:"total"`
		} `json:"page"`
	}{
		Data: items,
		Page: struct {
			Next  string `json:"next"`
			Total int    `json:"total"`
		}{
			Next:  "",
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
