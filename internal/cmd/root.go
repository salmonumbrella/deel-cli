package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/climerrors"
	"github.com/salmonumbrella/deel-cli/internal/config"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
	"github.com/salmonumbrella/deel-cli/internal/outfmt"
	"github.com/salmonumbrella/deel-cli/internal/secrets"
)

func emitAgentFlagError(ctx context.Context, message string) {
	if !outfmt.IsAgent(ctx) || AgentErrorEmitted() {
		return
	}
	// Keep output compact and machine-readable.
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(map[string]any{
		"ok": false,
		"error": map[string]any{
			"operation": "validating flags",
			"category":  "validation",
			"message":   message,
		},
	})
	markAgentErrorEmitted()
}

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
	agentFlag          bool
	timeoutFlag        time.Duration
	retriesFlag        int
	retryBaseFlag      time.Duration
	retryMaxFlag       time.Duration
	jsonlFlag          bool
	queryFlag          string
	jqFlag             string
	jsonFlag           bool
	dryRunFlag         bool
	dataOnlyFlag       bool
	rawFlag            bool
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
  deel contracts list # View contracts

JSON output:
  --json              # Output JSON (lists include data/page; single resources are wrapped in data)
  --json --items      # Output only the data array/object (for piping to jq)
  --json --raw        # Output raw JSON without the data envelope
  --json --jq '.data[].name'  # Apply JQ filter to JSON output

Agent mode:
  --agent             # Force JSON, disable color, and emit compact JSON for tools/agents`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if jsonFlag {
			if outputFlag != "" && outputFlag != "json" {
				emitAgentFlagError(ctx, fmt.Sprintf("cannot use --json with --output %q", outputFlag))
				return fmt.Errorf("cannot use --json with --output %q", outputFlag)
			}
			outputFlag = "json"
		}
		if jqFlag != "" {
			if queryFlag != "" && queryFlag != jqFlag {
				emitAgentFlagError(ctx, "cannot use --jq and --query with different values")
				return fmt.Errorf("cannot use --jq and --query with different values")
			}
			queryFlag = jqFlag
		}

		if jsonlFlag {
			if rawFlag {
				emitAgentFlagError(ctx, "cannot use --jsonl with --raw")
				return fmt.Errorf("cannot use --jsonl with --raw")
			}
			if dataOnlyFlag {
				emitAgentFlagError(ctx, "cannot use --jsonl with --items/--data-only (JSONL already streams items)")
				return fmt.Errorf("cannot use --jsonl with --items/--data-only (JSONL already streams items)")
			}
			if outputFlag != "" && outputFlag != "json" {
				emitAgentFlagError(ctx, fmt.Sprintf("cannot use --jsonl with --output %q (JSONL requires JSON output)", outputFlag))
				return fmt.Errorf("cannot use --jsonl with --output %q (JSONL requires JSON output)", outputFlag)
			}
			outputFlag = "json"
			jsonFlag = true
			ctx = outfmt.WithPrettyJSON(ctx, false)
			ctx = outfmt.WithJSONL(ctx, true)
		} else {
			ctx = outfmt.WithJSONL(ctx, false)
		}

		// Agent mode forces JSON output + no color + compact JSON.
		if agentFlag {
			// Some commands intentionally emit non-JSON bytes to stdout.
			if cmd.Name() == "completion" {
				emitAgentFlagError(ctx, "--agent is not supported for completion scripts")
				return fmt.Errorf("--agent is not supported for completion scripts")
			}
			if outputFlag != "" && outputFlag != "json" {
				emitAgentFlagError(ctx, fmt.Sprintf("cannot use --agent with --output %q (agent mode requires JSON output)", outputFlag))
				return fmt.Errorf("cannot use --agent with --output %q (agent mode requires JSON output)", outputFlag)
			}
			outputFlag = "json"
			jsonFlag = true
			colorFlag = "never"
			ctx = outfmt.WithAgent(ctx, true)
			// Don't override PrettyJSON if JSONL has already forced compact output.
			if !jsonlFlag {
				ctx = outfmt.WithPrettyJSON(ctx, false)
			}
		} else {
			ctx = outfmt.WithAgent(ctx, false)
			if !jsonlFlag {
				ctx = outfmt.WithPrettyJSON(ctx, true)
			}
		}

		// Validate output format
		if outputFlag != "" {
			switch outputFlag {
			case "text", "json":
				// Valid
			default:
				emitAgentFlagError(ctx, fmt.Sprintf("invalid output format %q (must be 'text' or 'json')", outputFlag))
				return fmt.Errorf("invalid output format %q (must be 'text' or 'json')", outputFlag)
			}
		}
		// Validate color mode
		if colorFlag != "" {
			switch colorFlag {
			case "auto", "always", "never":
				// Valid
			default:
				emitAgentFlagError(ctx, fmt.Sprintf("invalid color mode %q (must be 'auto', 'always', or 'never')", colorFlag))
				return fmt.Errorf("invalid color mode %q (must be 'auto', 'always', or 'never')", colorFlag)
			}
		}

		// Set output format in context (used by helpers that need to know if we're in JSON mode).
		format := "text"
		if outputFlag != "" {
			format = outputFlag
		} else if envOutput := os.Getenv(config.EnvOutput); envOutput != "" {
			format = envOutput
		}
		ctx = outfmt.WithFormat(ctx, format)

		// Set query filter in context
		if queryFlag != "" {
			ctx = outfmt.WithQuery(ctx, queryFlag)
		}
		// Set data-only mode in context
		if dataOnlyFlag {
			ctx = outfmt.WithDataOnly(ctx, true)
		}
		if rawFlag {
			ctx = outfmt.WithRaw(ctx, true)
		}
		// Set dry-run mode in context
		if dryRunFlag {
			ctx = dryrun.WithDryRun(ctx, true)
		}
		cmd.SetContext(ctx)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&accountFlag, "account", "", "Account to use (overrides DEEL_ACCOUNT)")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "Output format: text or json (default: text)")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON (alias for --output json)")
	rootCmd.PersistentFlags().BoolVar(&agentFlag, "agent", agentEnabledFromEnv(), "Agent mode: force JSON output, disable color, emit compact JSON")
	rootCmd.PersistentFlags().BoolVar(&jsonlFlag, "jsonl", false, "Stream JSON lines output (one JSON value per line; implies JSON output)")
	rootCmd.PersistentFlags().StringVar(&colorFlag, "color", "", "Color mode: auto, always, or never (default: auto)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVar(&queryFlag, "query", "", "JQ filter for JSON output")
	rootCmd.PersistentFlags().StringVar(&jqFlag, "jq", "", "JQ filter for JSON output (alias for --query)")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Preview changes without executing")
	rootCmd.PersistentFlags().BoolVar(&dataOnlyFlag, "data-only", false, "Output only the data array/object (use with --json)")
	rootCmd.PersistentFlags().BoolVar(&dataOnlyFlag, "data", false, "Alias for --data-only")
	rootCmd.PersistentFlags().BoolVar(&dataOnlyFlag, "items", false, "Alias for --data-only")
	rootCmd.PersistentFlags().BoolVar(&rawFlag, "raw", false, "Output raw JSON without the data envelope (use with --json)")
	rootCmd.PersistentFlags().StringVar(&idempotencyKeyFlag, "idempotency-key", "", "Idempotency key for write requests")
	rootCmd.PersistentFlags().DurationVar(&timeoutFlag, "timeout", 30*time.Second, "HTTP request timeout")
	rootCmd.PersistentFlags().IntVar(&retriesFlag, "retries", 3, "Max retry attempts for transient failures")
	rootCmd.PersistentFlags().DurationVar(&retryBaseFlag, "retry-base", 1*time.Second, "Base backoff for retries")
	rootCmd.PersistentFlags().DurationVar(&retryMaxFlag, "retry-max", 30*time.Second, "Max backoff for retries")

	// Agent-mode help: emit JSON schema instead of human help text.
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if agentFlag || agentEnabledFromEnv() {
			info := buildCommandInfo(cmd)
			enc := json.NewEncoder(os.Stdout)
			_ = enc.Encode(map[string]any{
				"ok":     true,
				"result": info,
			})
			return
		}
		defaultHelp(cmd, args)
	})

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
	f.SetAgentMode(agentFlag)
	if agentFlag {
		f.SetPrettyJSON(false)
	}
	f.SetQuery(queryFlag)
	f.SetDataOnly(dataOnlyFlag)
	f.SetRaw(rawFlag)
	return f
}

func categoryString(c climerrors.Category) string {
	switch c {
	case climerrors.CategoryAuth:
		return "auth"
	case climerrors.CategoryForbidden:
		return "forbidden"
	case climerrors.CategoryNotFound:
		return "not_found"
	case climerrors.CategoryValidation:
		return "validation"
	case climerrors.CategoryRateLimit:
		return "rate_limit"
	case climerrors.CategoryServer:
		return "server"
	case climerrors.CategoryNetwork:
		return "network"
	case climerrors.CategoryConfig:
		return "config"
	default:
		return "unknown"
	}
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

	// In agent mode, emit a structured JSON error on stdout so tools can parse it.
	// Only emit the first error object to avoid breaking stdout with multiple JSON blobs.
	if f.IsJSON() && f.IsAgentMode() && !AgentErrorEmitted() {
		_ = f.PrintJSON(map[string]any{
			"ok": false,
			"error": map[string]any{
				"operation":   cliErr.Operation,
				"category":    categoryString(cliErr.Category),
				"message":     climerrors.FriendlyMessage(cliErr.Err),
				"suggestions": cliErr.Suggestions,
			},
		})
		markAgentErrorEmitted()
	}

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
		client.SetTimeout(timeoutFlag)
		client.SetRetryConfig(retriesFlag, retryBaseFlag, retryMaxFlag)
		if idempotencyKeyFlag != "" {
			client.SetIdempotencyKey(idempotencyKeyFlag)
		} else if envKey := os.Getenv(config.EnvIdempotencyKey); envKey != "" {
			client.SetIdempotencyKey(envKey)
		}
		return client, nil
	}

	var store secrets.Store
	var storeErr error

	// Get account name
	account := accountFlag
	if account == "" {
		account = os.Getenv(config.EnvAccount)
	}
	if account == "" {
		var hint string
		store, storeErr = secrets.OpenDefault()
		if storeErr == nil {
			if creds, err := store.List(); err == nil {
				if len(creds) == 1 {
					account = creds[0].Name
				} else if len(creds) > 1 {
					names := make([]string, len(creds))
					for i, c := range creds {
						names[i] = c.Name
					}
					hint = fmt.Sprintf("Available accounts: %s. Use --account flag or set DEEL_ACCOUNT env", strings.Join(names, ", "))
				}
			}
		}
		if account == "" {
			if hint == "" {
				hint = "Use --account flag, DEEL_ACCOUNT env, or DEEL_TOKEN for direct auth"
			}
			return nil, fmt.Errorf("no account specified. %s", hint)
		}
	}

	// Load from keychain
	if store == nil {
		store, storeErr = secrets.OpenDefault()
	}
	if storeErr != nil {
		return nil, fmt.Errorf("failed to open credential store: %w", storeErr)
	}

	creds, err := store.Get(account)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for account %q: %w", account, err)
	}

	client := api.NewClient(creds.Token)
	client.SetDebug(debugFlag)
	client.SetTimeout(timeoutFlag)
	client.SetRetryConfig(retriesFlag, retryBaseFlag, retryMaxFlag)
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
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText(fmt.Sprintf("deel version %s", Version))
			f.PrintText("  commit: " + Commit)
			f.PrintText("  built:  " + BuildDate)
		}, map[string]string{
			"version":   Version,
			"commit":    Commit,
			"buildDate": BuildDate,
		})
	},
}
