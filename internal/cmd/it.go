package cmd

import (
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
	itOrdersLimitFlag  int
)

var itAssetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "List IT assets",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		resp, err := client.ListITAssets(cmd.Context(), api.ITAssetsListParams{
			Status: itAssetsStatusFlag,
			Type:   itAssetsTypeFlag,
			Limit:  itAssetsLimitFlag,
			Cursor: itAssetsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list assets: %v", err)
			return err
		}

		return f.Output(func() {
			if len(resp.Data) == 0 {
				f.PrintText("No IT assets found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "SERIAL", "STATUS", "ASSIGNED TO")
			for _, a := range resp.Data {
				table.AddRow(a.ID, a.Name, a.Type, a.SerialNumber, a.Status, a.AssignedTo)
			}
			table.Render()
		}, resp)
	},
}

var itOrdersCmd = &cobra.Command{
	Use:   "orders",
	Short: "List IT orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		orders, err := client.ListITOrders(cmd.Context(), itOrdersLimitFlag)
		if err != nil {
			f.PrintError("Failed to list orders: %v", err)
			return err
		}

		return f.Output(func() {
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
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		policies, err := client.ListHardwarePolicies(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list policies: %v", err)
			return err
		}

		return f.Output(func() {
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
	itAssetsCmd.Flags().IntVar(&itAssetsLimitFlag, "limit", 50, "Maximum results")
	itAssetsCmd.Flags().StringVar(&itAssetsCursorFlag, "cursor", "", "Pagination cursor")

	itOrdersCmd.Flags().IntVar(&itOrdersLimitFlag, "limit", 50, "Maximum results")

	itCmd.AddCommand(itAssetsCmd)
	itCmd.AddCommand(itOrdersCmd)
	itCmd.AddCommand(itPoliciesCmd)
}
