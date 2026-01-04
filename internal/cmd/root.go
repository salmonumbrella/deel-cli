package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/climerrors"
	"github.com/salmonumbrella/deel-cli/internal/config"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
	"github.com/salmonumbrella/deel-cli/internal/outfmt"
	"github.com/salmonumbrella/deel-cli/internal/secrets"
)

// Version info (set by ldflags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Global flags
var (
	accountFlag        string
	outputFlag         string
	colorFlag          string
	debugFlag          bool
	queryFlag          string
	dryRunFlag         bool
	idempotencyKeyFlag string
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "deel",
	Short: "CLI for Deel - manage people, contracts, payroll, and more",
	Long: `Deel CLI provides command-line access to the Deel platform.

Manage your workforce, contracts, time off, payroll, and more from the terminal.

Get started:
  deel auth login     # Authenticate via browser
  deel people list    # List your workforce
  deel contracts list # View contracts`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate output format
		if outputFlag != "" {
			switch outputFlag {
			case "text", "json":
				// Valid
			default:
				return fmt.Errorf("invalid output format %q (must be 'text' or 'json')", outputFlag)
			}
		}
		// Validate color mode
		if colorFlag != "" {
			switch colorFlag {
			case "auto", "always", "never":
				// Valid
			default:
				return fmt.Errorf("invalid color mode %q (must be 'auto', 'always', or 'never')", colorFlag)
			}
		}
		// Set query filter in context
		if queryFlag != "" {
			ctx := outfmt.WithQuery(cmd.Context(), queryFlag)
			cmd.SetContext(ctx)
		}
		// Set dry-run mode in context
		if dryRunFlag {
			ctx := dryrun.WithDryRun(cmd.Context(), true)
			cmd.SetContext(ctx)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&accountFlag, "account", "", "Account to use (overrides DEEL_ACCOUNT)")
	rootCmd.PersistentFlags().StringVar(&outputFlag, "output", "", "Output format: text or json (default: text)")
	rootCmd.PersistentFlags().StringVar(&colorFlag, "color", "", "Color mode: auto, always, or never (default: auto)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVar(&queryFlag, "query", "", "JQ filter for JSON output")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Preview changes without executing")
	rootCmd.PersistentFlags().StringVar(&idempotencyKeyFlag, "idempotency-key", "", "Idempotency key for write requests")

	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(peopleCmd)
	rootCmd.AddCommand(contractsCmd)
	rootCmd.AddCommand(milestonesCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(timeOffCmd)
	rootCmd.AddCommand(payrollCmd)
	rootCmd.AddCommand(invoicesCmd)
	rootCmd.AddCommand(paymentsCmd)
	rootCmd.AddCommand(reportsCmd)
	rootCmd.AddCommand(payoutsCmd)
	rootCmd.AddCommand(benefitsCmd)
	rootCmd.AddCommand(itCmd)
	rootCmd.AddCommand(teamsCmd)
	rootCmd.AddCommand(orgCmd)
	rootCmd.AddCommand(onboardingCmd)
	rootCmd.AddCommand(complianceCmd)
	rootCmd.AddCommand(bgCheckCmd)
	rootCmd.AddCommand(immigrationCmd)
	rootCmd.AddCommand(atsCmd)
	rootCmd.AddCommand(shiftsCmd)
	rootCmd.AddCommand(calcCmd)
	rootCmd.AddCommand(tokensCmd)
	rootCmd.AddCommand(webhooksCmd)
	rootCmd.AddCommand(timesheetsCmd)
	rootCmd.AddCommand(eorCmd)
	rootCmd.AddCommand(gpCmd)
	rootCmd.AddCommand(candidatesCmd)
	rootCmd.AddCommand(screeningsCmd)
	rootCmd.AddCommand(costCentersCmd)
	rootCmd.AddCommand(offboardingCmd)
}

// ExecuteContext runs the root command with context
func ExecuteContext(ctx context.Context, args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}

// getFormatter creates a formatter based on flags and environment
func getFormatter() *outfmt.Formatter {
	format := outfmt.FormatText
	if outputFlag != "" {
		format = outfmt.Format(outputFlag)
	} else if envOutput := os.Getenv(config.EnvOutput); envOutput != "" {
		format = outfmt.Format(envOutput)
	}

	colorMode := "auto"
	if colorFlag != "" {
		colorMode = colorFlag
	} else if envColor := os.Getenv(config.EnvColor); envColor != "" {
		colorMode = envColor
	}

	f := outfmt.New(os.Stdout, os.Stderr, format, colorMode)
	f.SetQuery(queryFlag)
	return f
}

// HandleError wraps an error with context and prints it using the formatter
func HandleError(f *outfmt.Formatter, err error, operation string) error {
	if err == nil {
		return nil
	}
	cliErr := climerrors.Wrap(err, operation)

	var buf bytes.Buffer
	climerrors.FormatError(&buf, cliErr)
	f.PrintError("%s", strings.TrimSpace(buf.String()))

	// Return a simple error with friendly message so cobra doesn't print raw JSON
	friendlyMsg := climerrors.FriendlyMessage(err)
	return fmt.Errorf("failed %s: %s", operation, friendlyMsg)
}

// getClient creates an API client using the configured credentials
func getClient() (*api.Client, error) {
	// First check for direct token in environment
	if token := os.Getenv(config.EnvToken); token != "" {
		client := api.NewClient(token)
		client.SetDebug(debugFlag)
		if idempotencyKeyFlag != "" {
			client.SetIdempotencyKey(idempotencyKeyFlag)
		} else if envKey := os.Getenv(config.EnvIdempotencyKey); envKey != "" {
			client.SetIdempotencyKey(envKey)
		}
		return client, nil
	}

	// Get account name
	account := accountFlag
	if account == "" {
		account = os.Getenv(config.EnvAccount)
	}
	if account == "" {
		// Try to list available accounts for a helpful error message
		hint := "Use --account flag, DEEL_ACCOUNT env, or DEEL_TOKEN for direct auth"
		if store, err := secrets.OpenDefault(); err == nil {
			if creds, err := store.List(); err == nil && len(creds) > 0 {
				names := make([]string, len(creds))
				for i, c := range creds {
					names[i] = c.Name
				}
				hint = fmt.Sprintf("Available accounts: %s. Use --account flag or set DEEL_ACCOUNT env", strings.Join(names, ", "))
			}
		}
		return nil, fmt.Errorf("no account specified. %s", hint)
	}

	// Load from keychain
	store, err := secrets.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to open credential store: %w", err)
	}

	creds, err := store.Get(account)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for account %q: %w", account, err)
	}

	client := api.NewClient(creds.Token)
	client.SetDebug(debugFlag)
	if idempotencyKeyFlag != "" {
		client.SetIdempotencyKey(idempotencyKeyFlag)
	} else if envKey := os.Getenv(config.EnvIdempotencyKey); envKey != "" {
		client.SetIdempotencyKey(envKey)
	}
	return client, nil
}

// versionCmd shows version info
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("deel version %s\n", Version)
		fmt.Printf("  commit: %s\n", Commit)
		fmt.Printf("  built:  %s\n", BuildDate)
	},
}
