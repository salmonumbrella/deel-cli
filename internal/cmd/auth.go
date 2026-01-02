package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/auth"
	"github.com/salmonumbrella/deel-cli/internal/secrets"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Authenticate with Deel and manage stored credentials.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via browser",
	Long:  "Opens a browser window to securely enter your Deel Personal Access Token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		store, err := secrets.OpenDefault()
		if err != nil {
			f.PrintError("Failed to open credential store: %v", err)
			return err
		}

		server, err := auth.NewSetupServer(store)
		if err != nil {
			f.PrintError("Failed to start auth server: %v", err)
			return err
		}

		f.PrintText("Opening browser for authentication...")
		f.PrintText("If the browser doesn't open, navigate to the URL shown.")
		f.PrintText("")

		result, err := server.Start(cmd.Context())
		if err != nil {
			f.PrintError("Authentication failed: %v", err)
			return err
		}

		f.PrintSuccess("Successfully authenticated as %q", result.AccountName)
		f.PrintText("")
		f.PrintText("Test your connection with:")
		f.PrintText("  deel auth test --account " + result.AccountName)
		return nil
	},
}

var authAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add credentials manually",
	Long:  "Add a Deel account by entering your Personal Access Token at the prompt.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		accountName := strings.ToLower(strings.TrimSpace(args[0]))

		if err := auth.ValidateAccountName(accountName); err != nil {
			f.PrintError("Invalid account name: %v", err)
			return err
		}

		// Prompt for token
		f.PrintText("Enter your Deel Personal Access Token:")
		f.PrintText("(Get it from https://app.deel.com/settings/api)")
		f.PrintText("")

		var token string
		if term.IsTerminal(int(os.Stdin.Fd())) {
			// Secure input (no echo)
			tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				f.PrintError("Failed to read token: %v", err)
				return err
			}
			token = auth.SanitizeToken(string(tokenBytes))
			f.PrintText("") // New line after hidden input
		} else {
			// Non-interactive (pipe)
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				f.PrintError("Failed to read token: %v", err)
				return err
			}
			token = auth.SanitizeToken(line)
		}

		if err := auth.ValidateToken(token); err != nil {
			f.PrintError("Invalid token: %v", err)
			return err
		}

		store, err := secrets.OpenDefault()
		if err != nil {
			f.PrintError("Failed to open credential store: %v", err)
			return err
		}

		// Validate token against API before saving
		f.PrintText("Validating token with Deel API...")
		client := api.NewClient(token)
		// Use /rest/v2/contracts with limit=1 as a lightweight validation endpoint
		if _, err := client.Get(cmd.Context(), "/rest/v2/contracts?limit=1"); err != nil {
			f.PrintError("Token validation failed: %v", err)
			return fmt.Errorf("invalid token: failed to authenticate with Deel API")
		}
		f.PrintSuccess("Token validated successfully")

		err = store.Set(accountName, secrets.Credentials{
			Token: token,
		})
		if err != nil {
			f.PrintError("Failed to save credentials: %v", err)
			return err
		}

		f.PrintSuccess("Credentials saved for account %q", accountName)
		return nil
	},
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		store, err := secrets.OpenDefault()
		if err != nil {
			f.PrintError("Failed to open credential store: %v", err)
			return err
		}

		creds, err := store.List()
		if err != nil {
			f.PrintError("Failed to list credentials: %v", err)
			return err
		}

		if len(creds) == 0 {
			f.PrintText("No accounts configured.")
			f.PrintText("")
			f.PrintText("Add an account with:")
			f.PrintText("  deel auth login")
			return nil
		}

		return f.Output(func() {
			table := f.NewTable("NAME", "CREATED")
			for _, c := range creds {
				created := "unknown"
				if !c.CreatedAt.IsZero() {
					created = c.CreatedAt.Format(time.RFC3339)
				}
				table.AddRow(c.Name, created)
			}
			table.Render()
		}, creds)
	},
}

var authRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		accountName := strings.ToLower(strings.TrimSpace(args[0]))

		if err := auth.ValidateAccountName(accountName); err != nil {
			f.PrintError("Invalid account name: %v", err)
			return err
		}

		store, err := secrets.OpenDefault()
		if err != nil {
			f.PrintError("Failed to open credential store: %v", err)
			return err
		}

		if err := store.Delete(accountName); err != nil {
			f.PrintError("Failed to remove account: %v", err)
			return err
		}

		f.PrintSuccess("Removed account %q", accountName)
		return nil
	},
}

var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test authentication",
	Long:  "Verify that your stored credentials work by making a test API call.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		f.PrintText("Testing connection...")

		// Use /rest/v2/contracts with limit=1 as a lightweight validation endpoint
		_, err = client.Get(cmd.Context(), "/rest/v2/contracts?limit=1")
		if err != nil {
			f.PrintError("Connection failed: %v", err)
			return err
		}

		f.PrintSuccess("Connection successful!")
		return nil
	},
}

var authManageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage accounts in browser",
	Long:  "Opens a browser window to view, add, and remove Deel accounts.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		store, err := secrets.OpenDefault()
		if err != nil {
			f.PrintError("Failed to open credential store: %v", err)
			return err
		}

		server, err := auth.NewSetupServerWithMode(store, auth.ModeManage)
		if err != nil {
			f.PrintError("Failed to start server: %v", err)
			return err
		}

		f.PrintText("Opening account manager in browser...")
		f.PrintText("If the browser doesn't open, navigate to the URL shown.")
		f.PrintText("")

		_, err = server.Start(cmd.Context())
		if err != nil {
			// Context cancelled is normal when user closes browser
			if err.Error() == "setup cancelled" || err.Error() == "context canceled" {
				return nil
			}
			f.PrintError("Error: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authAddCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authRemoveCmd)
	authCmd.AddCommand(authTestCmd)
	authCmd.AddCommand(authManageCmd)
}
