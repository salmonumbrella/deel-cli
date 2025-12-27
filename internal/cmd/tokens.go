package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
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
			f.PrintError("--worker is required")
			return fmt.Errorf("missing required flag")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		token, err := client.CreateWorkerAccessToken(cmd.Context(), api.CreateWorkerAccessTokenParams{
			WorkerID: tokensWorkerFlag,
			Scope:    tokensScopeFlag,
			TTL:      tokensTTLFlag,
		})
		if err != nil {
			f.PrintError("Failed to create token: %v", err)
			return err
		}

		return f.Output(func() {
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
