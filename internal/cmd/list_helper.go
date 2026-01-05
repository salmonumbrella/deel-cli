// internal/cmd/list_helper.go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

// ListResult represents the result of a paginated list operation.
type ListResult[T any] struct {
	Items   []T
	HasMore bool
}

// ListConfig defines how a list command behaves.
type ListConfig[T any] struct {
	Use     string
	Short   string
	Long    string
	Example string

	// Fetch function - called with page/pageSize, returns items and hasMore
	Fetch func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[T], error)

	// Output configuration
	Headers      []string
	RowFunc      func(T) []string
	EmptyMessage string
}

// NewListCommand creates a cobra command from ListConfig.
func NewListCommand[T any](cfg ListConfig[T], getClientFn func() (*api.Client, error)) *cobra.Command {
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageSize < 10 {
				pageSize = 10
			}

			client, err := getClientFn()
			if err != nil {
				return err
			}

			result, err := cfg.Fetch(cmd.Context(), client, page, pageSize)
			if err != nil {
				return err
			}

			f := getFormatter()

			// Handle empty results
			if len(result.Items) == 0 {
				return f.Output(func() {
					f.PrintText(cfg.EmptyMessage)
				}, map[string]interface{}{
					"items":    result.Items,
					"has_more": result.HasMore,
				})
			}

			// Handle output
			return f.Output(func() {
				table := f.NewTable(cfg.Headers...)
				for _, item := range result.Items {
					table.AddRow(cfg.RowFunc(item)...)
				}
				table.Render()
				if result.HasMore {
					fmt.Fprintln(os.Stderr, "# More results available")
				}
			}, map[string]interface{}{
				"items":    result.Items,
				"has_more": result.HasMore,
			})
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "limit", 100, "Max results (min 10)")
	return cmd
}
