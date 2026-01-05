package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var costCentersCmd = &cobra.Command{
	Use:     "cost-centers",
	Aliases: []string{"cost-center", "cc"},
	Short:   "Manage cost centers",
	Long:    "List and sync cost centers for your organization.",
}

var (
	costCenterFileFlag   string
	costCentersLimitFlag int
)

var costCentersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cost centers",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		centers, err := client.ListCostCenters(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list cost centers: %v", err)
			return err
		}

		// Apply client-side limit
		if costCentersLimitFlag > 0 && len(centers) > costCentersLimitFlag {
			centers = centers[:costCentersLimitFlag]
		}

		return f.Output(func() {
			if len(centers) == 0 {
				f.PrintText("No cost centers found.")
				return
			}
			table := f.NewTable("ID", "CODE", "NAME", "STATUS", "CREATED")
			for _, c := range centers {
				table.AddRow(c.ID, c.Code, c.Name, c.Status, c.CreatedAt)
			}
			table.Render()
		}, centers)
	},
}

var costCentersSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync cost centers from a JSON file",
	Long: `Sync cost centers from a JSON file. Requires --file flag.

The JSON file should contain an array of cost centers:
[
  {
    "name": "Engineering",
    "code": "ENG",
    "description": "Engineering department"
  },
  {
    "name": "Sales",
    "code": "SALES"
  }
]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if costCenterFileFlag == "" {
			f.PrintError("--file is required")
			return fmt.Errorf("missing required flag")
		}

		// Read the file
		data, err := os.ReadFile(costCenterFileFlag)
		if err != nil {
			f.PrintError("Failed to read file: %v", err)
			return err
		}

		// Parse the JSON
		var centers []api.CostCenterInput
		if err := json.Unmarshal(data, &centers); err != nil {
			f.PrintError("Failed to parse JSON: %v", err)
			return err
		}

		if len(centers) == 0 {
			f.PrintError("No cost centers found in file")
			return fmt.Errorf("empty cost centers array")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "SYNC",
			Resource:    "CostCenters",
			Description: "Sync cost centers",
			Details: map[string]string{
				"File":       costCenterFileFlag,
				"TotalItems": fmt.Sprintf("%d", len(centers)),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		synced, err := client.SyncCostCenters(cmd.Context(), api.SyncCostCentersParams{
			CostCenters: centers,
		})
		if err != nil {
			f.PrintError("Failed to sync cost centers: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Synced %d cost centers successfully", len(synced))
			table := f.NewTable("ID", "CODE", "NAME", "STATUS")
			for _, c := range synced {
				table.AddRow(c.ID, c.Code, c.Name, c.Status)
			}
			table.Render()
		}, synced)
	},
}

func init() {
	// List command flags
	costCentersListCmd.Flags().IntVar(&costCentersLimitFlag, "limit", 100, "Maximum results")

	// Sync command flags
	costCentersSyncCmd.Flags().StringVar(&costCenterFileFlag, "file", "", "JSON file containing cost centers (required)")

	// Add subcommands
	costCentersCmd.AddCommand(costCentersListCmd)
	costCentersCmd.AddCommand(costCentersSyncCmd)
}
