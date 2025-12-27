package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var shiftsCmd = &cobra.Command{
	Use:   "shifts",
	Short: "Manage worker shifts",
	Long:  "View shifts and shift rates.",
}

var (
	shiftsWorkerFlag  string
	shiftsStartFlag   string
	shiftsEndFlag     string
	shiftsLimitFlag   int
	shiftsCountryFlag string
)

var shiftsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List shifts",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		shifts, err := client.ListShifts(cmd.Context(), api.ShiftsListParams{
			WorkerID:  shiftsWorkerFlag,
			StartDate: shiftsStartFlag,
			EndDate:   shiftsEndFlag,
			Limit:     shiftsLimitFlag,
		})
		if err != nil {
			f.PrintError("Failed to list shifts: %v", err)
			return err
		}

		return f.Output(func() {
			if len(shifts) == 0 {
				f.PrintText("No shifts found.")
				return
			}
			table := f.NewTable("ID", "WORKER", "DATE", "START", "END", "HOURS", "STATUS")
			for _, s := range shifts {
				table.AddRow(s.ID, s.WorkerName, s.Date, s.StartTime, s.EndTime, fmt.Sprintf("%.1f", s.Hours), s.Status)
			}
			table.Render()
		}, shifts)
	},
}

var shiftsRatesCmd = &cobra.Command{
	Use:   "rates",
	Short: "List shift rates",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if shiftsCountryFlag == "" {
			f.PrintError("--country is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		rates, err := client.ListShiftRates(cmd.Context(), shiftsCountryFlag)
		if err != nil {
			f.PrintError("Failed to list rates: %v", err)
			return err
		}

		return f.Output(func() {
			if len(rates) == 0 {
				f.PrintText("No rates found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "RATE")
			for _, r := range rates {
				rate := fmt.Sprintf("%.2f %s", r.Rate, r.Currency)
				table.AddRow(r.ID, r.Name, r.Type, rate)
			}
			table.Render()
		}, rates)
	},
}

func init() {
	shiftsListCmd.Flags().StringVar(&shiftsWorkerFlag, "worker", "", "Filter by worker ID")
	shiftsListCmd.Flags().StringVar(&shiftsStartFlag, "start", "", "Start date YYYY-MM-DD")
	shiftsListCmd.Flags().StringVar(&shiftsEndFlag, "end", "", "End date YYYY-MM-DD")
	shiftsListCmd.Flags().IntVar(&shiftsLimitFlag, "limit", 50, "Maximum results")

	shiftsRatesCmd.Flags().StringVar(&shiftsCountryFlag, "country", "", "Country code (required)")

	shiftsCmd.AddCommand(shiftsListCmd)
	shiftsCmd.AddCommand(shiftsRatesCmd)
}
