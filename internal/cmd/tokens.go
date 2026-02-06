package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Manage access tokens",
	Long:  "Create and manage worker access tokens.",
}

var (
	tokensWorkerFlag string
	tokensScopeFlag  string
	tokensTTLFlag    int
)

var tokensCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a worker access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if tokensWorkerFlag == "" {
			return failValidation(cmd, f, "--worker is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "WorkerToken",
			Description: "Create worker access token",
			Details: map[string]string{
				"WorkerID": tokensWorkerFlag,
				"Scope":    tokensScopeFlag,
				"TTL":      fmt.Sprintf("%d", tokensTTLFlag),
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		token, err := client.CreateWorkerAccessToken(cmd.Context(), api.CreateWorkerAccessTokenParams{
			WorkerID: tokensWorkerFlag,
			Scope:    tokensScopeFlag,
			TTL:      tokensTTLFlag,
		})
		if err != nil {
			return HandleError(f, err, "create token")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("Token:     " + token.Token)
			f.PrintText("Worker ID: " + token.WorkerID)
			f.PrintText("Scope:     " + token.Scope)
			f.PrintText("Expires:   " + token.ExpiresAt)
		}, token)
	},
}

func init() {
	tokensCreateCmd.Flags().StringVar(&tokensWorkerFlag, "worker", "", "Worker ID (required)")
	tokensCreateCmd.Flags().StringVar(&tokensScopeFlag, "scope", "", "Token scope")
	tokensCreateCmd.Flags().IntVar(&tokensTTLFlag, "ttl", 3600, "Token TTL in seconds")

	tokensCmd.AddCommand(tokensCreateCmd)
}
