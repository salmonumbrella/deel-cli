package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var itCmd = &cobra.Command{
	Use:   "it",
	Short: "Manage IT assets and equipment",
	Long:  "View IT assets, orders, and hardware policies.",
}

var (
	itAssetsStatusFlag string
	itAssetsTypeFlag   string
	itAssetsLimitFlag  int
	itAssetsCursorFlag string
	itAssetsAllFlag    bool
	itOrdersLimitFlag  int
)

var itAssetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "List IT assets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, f, err := initClient("listing it assets")
		if err != nil {
			return err
		}

		assets, _, hasMore, err := collectCursorItems(cmd.Context(), itAssetsAllFlag, itAssetsCursorFlag, itAssetsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ITAsset], error) {
			resp, err := client.ListITAssets(ctx, api.ITAssetsListParams{
				Status: itAssetsStatusFlag,
				Type:   itAssetsTypeFlag,
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.ITAsset]{}, err
			}
			return CursorListResult[api.ITAsset]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing it assets")
		}

		response := api.ITAssetsListResponse{
			Data: assets,
		}
		response.Page.Next = ""

		return outputList(cmd, f, assets, hasMore, "No IT assets found.", []string{"ID", "NAME", "TYPE", "SERIAL", "STATUS", "ASSIGNED TO"}, func(a api.ITAsset) []string {
			return []string{a.ID, a.Name, a.Type, a.SerialNumber, a.Status, a.AssignedTo}
		}, response)
	},
}

var itOrdersCmd = &cobra.Command{
	Use:   "orders",
	Short: "List IT orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		orders, err := client.ListITOrders(cmd.Context(), itOrdersLimitFlag)
		if err != nil {
			return HandleError(f, err, "list orders")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(orders) == 0 {
				f.PrintText("No IT orders found.")
				return
			}
			table := f.NewTable("ID", "EMPLOYEE", "TYPE", "ITEMS", "COST", "STATUS")
			for _, o := range orders {
				cost := fmt.Sprintf("%.2f %s", o.TotalCost, o.Currency)
				table.AddRow(o.ID, o.EmployeeName, o.Type, fmt.Sprintf("%d", o.Items), cost, o.Status)
			}
			table.Render()
		}, orders)
	},
}

var itPoliciesCmd = &cobra.Command{
	Use:   "policies",
	Short: "List hardware policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		policies, err := client.ListHardwarePolicies(cmd.Context())
		if err != nil {
			return HandleError(f, err, "list policies")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(policies) == 0 {
				f.PrintText("No hardware policies found.")
				return
			}
			table := f.NewTable("ID", "NAME", "COUNTRY", "BUDGET")
			for _, p := range policies {
				budget := fmt.Sprintf("%.2f %s", p.Budget, p.Currency)
				table.AddRow(p.ID, p.Name, p.Country, budget)
			}
			table.Render()
		}, policies)
	},
}

func init() {
	itAssetsCmd.Flags().StringVar(&itAssetsStatusFlag, "status", "", "Filter by status")
	itAssetsCmd.Flags().StringVar(&itAssetsTypeFlag, "type", "", "Filter by type")
	itAssetsCmd.Flags().IntVar(&itAssetsLimitFlag, "limit", 100, "Maximum results")
	itAssetsCmd.Flags().StringVar(&itAssetsCursorFlag, "cursor", "", "Pagination cursor")
	itAssetsCmd.Flags().BoolVar(&itAssetsAllFlag, "all", false, "Fetch all pages")

	itOrdersCmd.Flags().IntVar(&itOrdersLimitFlag, "limit", 100, "Maximum results")

	itCmd.AddCommand(itAssetsCmd)
	itCmd.AddCommand(itOrdersCmd)
	itCmd.AddCommand(itPoliciesCmd)
}
