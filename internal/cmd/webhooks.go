package cmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhooks",
	Long:  "List, view, and manage webhook subscriptions in your Deel organization.",
}

var (
	webhooksURLFlag         string
	webhooksEventsFlag      []string
	webhooksDescriptionFlag string
)

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		webhooks, err := client.ListWebhooks(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list webhooks: %v", err)
			return err
		}

		return f.Output(func() {
			if len(webhooks) == 0 {
				f.PrintText("No webhooks found.")
				return
			}
			table := f.NewTable("ID", "URL", "EVENTS", "STATUS", "CREATED")
			for _, w := range webhooks {
				eventsStr := strings.Join(w.Events, ", ")
				if len(eventsStr) > 50 {
					eventsStr = eventsStr[:47] + "..."
				}
				table.AddRow(w.ID, w.URL, eventsStr, w.Status, w.CreatedAt)
			}
			table.Render()
		}, webhooks)
	},
}

var webhooksGetCmd = &cobra.Command{
	Use:   "get <webhook-id>",
	Short: "Get webhook details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		webhook, err := client.GetWebhook(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get webhook: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:          " + webhook.ID)
			f.PrintText("URL:         " + webhook.URL)
			f.PrintText("Status:      " + webhook.Status)
			f.PrintText("Events:      " + strings.Join(webhook.Events, ", "))
			if webhook.Description != "" {
				f.PrintText("Description: " + webhook.Description)
			}
			if webhook.Secret != "" {
				f.PrintText("Secret:      " + webhook.Secret)
			}
			f.PrintText("Created:     " + webhook.CreatedAt)
		}, webhook)
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new webhook",
	Long:  "Create a new webhook subscription. Requires --url and --events flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if webhooksURLFlag == "" {
			f.PrintError("--url flag is required")
			return fmt.Errorf("--url flag is required")
		}
		if len(webhooksEventsFlag) == 0 {
			f.PrintError("--events flag is required (provide at least one event)")
			return fmt.Errorf("--events flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Webhook",
			Description: "Create webhook",
			Details: map[string]string{
				"URL":         webhooksURLFlag,
				"Events":      strings.Join(webhooksEventsFlag, ","),
				"Description": webhooksDescriptionFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		webhook, err := client.CreateWebhook(cmd.Context(), api.CreateWebhookParams{
			URL:         webhooksURLFlag,
			Events:      webhooksEventsFlag,
			Description: webhooksDescriptionFlag,
		})
		if err != nil {
			f.PrintError("Failed to create webhook: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Webhook created successfully")
			f.PrintText("ID:          " + webhook.ID)
			f.PrintText("URL:         " + webhook.URL)
			f.PrintText("Status:      " + webhook.Status)
			f.PrintText("Events:      " + strings.Join(webhook.Events, ", "))
			if webhook.Description != "" {
				f.PrintText("Description: " + webhook.Description)
			}
			if webhook.Secret != "" {
				f.PrintText("Secret:      " + webhook.Secret)
			}
		}, webhook)
	},
}

var webhooksUpdateCmd = &cobra.Command{
	Use:   "update <webhook-id>",
	Short: "Update a webhook",
	Long:  "Update an existing webhook. Optional: --url, --events, --description flags.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Check if at least one flag was provided
		if !cmd.Flags().Changed("url") && !cmd.Flags().Changed("events") && !cmd.Flags().Changed("description") {
			f.PrintError("At least one flag (--url, --events, or --description) must be provided")
			return fmt.Errorf("no update flags provided")
		}

		details := map[string]string{
			"ID": args[0],
		}
		if cmd.Flags().Changed("url") {
			details["URL"] = webhooksURLFlag
		}
		if cmd.Flags().Changed("events") {
			details["Events"] = strings.Join(webhooksEventsFlag, ",")
		}
		if cmd.Flags().Changed("description") {
			details["Description"] = webhooksDescriptionFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Webhook",
			Description: "Update webhook",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.UpdateWebhookParams{}
		if cmd.Flags().Changed("url") {
			params.URL = webhooksURLFlag
		}
		if cmd.Flags().Changed("events") {
			params.Events = webhooksEventsFlag
		}
		if cmd.Flags().Changed("description") {
			params.Description = webhooksDescriptionFlag
		}

		webhook, err := client.UpdateWebhook(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to update webhook: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Webhook updated successfully")
			f.PrintText("ID:          " + webhook.ID)
			f.PrintText("URL:         " + webhook.URL)
			f.PrintText("Status:      " + webhook.Status)
			f.PrintText("Events:      " + strings.Join(webhook.Events, ", "))
			if webhook.Description != "" {
				f.PrintText("Description: " + webhook.Description)
			}
		}, webhook)
	},
}

var webhooksEnableCmd = &cobra.Command{
	Use:   "enable <webhook-id>",
	Short: "Enable a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Webhook",
			Description: "Enable webhook",
			Details: map[string]string{
				"ID":     args[0],
				"Status": "active",
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		webhook, err := client.UpdateWebhook(cmd.Context(), args[0], api.UpdateWebhookParams{
			Status: "active",
		})
		if err != nil {
			f.PrintError("Failed to enable webhook: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Webhook enabled successfully")
			f.PrintText("ID:     " + webhook.ID)
			f.PrintText("Status: " + webhook.Status)
		}, webhook)
	},
}

var webhooksDisableCmd = &cobra.Command{
	Use:   "disable <webhook-id>",
	Short: "Disable a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Webhook",
			Description: "Disable webhook",
			Details: map[string]string{
				"ID":     args[0],
				"Status": "disabled",
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		webhook, err := client.UpdateWebhook(cmd.Context(), args[0], api.UpdateWebhookParams{
			Status: "disabled",
		})
		if err != nil {
			f.PrintError("Failed to disable webhook: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Webhook disabled successfully")
			f.PrintText("ID:     " + webhook.ID)
			f.PrintText("Status: " + webhook.Status)
		}, webhook)
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <webhook-id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "Webhook",
			Description: "Delete webhook",
			Details: map[string]string{
				"ID": args[0],
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		err = client.DeleteWebhook(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to delete webhook: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Webhook deleted successfully")
		}, map[string]string{"status": "deleted", "id": args[0]})
	},
}

var webhooksEventTypesCmd = &cobra.Command{
	Use:   "event-types",
	Short: "List available webhook event types",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		eventTypes, err := client.ListWebhookEventTypes(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list webhook event types: %v", err)
			return err
		}

		return f.Output(func() {
			if len(eventTypes) == 0 {
				f.PrintText("No event types found.")
				return
			}
			table := f.NewTable("EVENT NAME", "DESCRIPTION")
			for _, et := range eventTypes {
				table.AddRow(et.Name, et.Description)
			}
			table.Render()
		}, eventTypes)
	},
}

// Flags for verify command
var (
	webhooksVerifySecretFlag        string
	webhooksVerifySignatureFlag     string
	webhooksVerifyPayloadFlag       string
	webhooksVerifyPayloadFileFlag   string
	webhooksVerifySignedPayloadFlag string
)

var webhooksVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a webhook signature",
	Long:  "Verify a webhook signature using HMAC-SHA256. Provide --secret and --signature, plus --payload/--payload-file or --signed-payload.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if webhooksVerifySecretFlag == "" {
			f.PrintError("--secret is required")
			return fmt.Errorf("--secret is required")
		}
		if webhooksVerifySignatureFlag == "" {
			f.PrintError("--signature is required")
			return fmt.Errorf("--signature is required")
		}

		var payload string
		if webhooksVerifySignedPayloadFlag != "" {
			payload = webhooksVerifySignedPayloadFlag
		} else if webhooksVerifyPayloadFileFlag != "" {
			data, err := os.ReadFile(webhooksVerifyPayloadFileFlag)
			if err != nil {
				f.PrintError("Failed to read payload file: %v", err)
				return err
			}
			payload = string(data)
		} else if webhooksVerifyPayloadFlag != "" {
			payload = webhooksVerifyPayloadFlag
		} else {
			f.PrintError("Provide --payload, --payload-file, or --signed-payload")
			return fmt.Errorf("payload is required")
		}

		provided := extractSignatureValue(webhooksVerifySignatureFlag)
		computed := computeHMACSHA256(webhooksVerifySecretFlag, payload)

		match := hmac.Equal([]byte(strings.ToLower(provided)), []byte(computed))
		result := map[string]any{
			"match":              match,
			"computed_signature": computed,
			"provided_signature": provided,
		}

		return f.Output(func() {
			if match {
				f.PrintSuccess("Signature is valid")
			} else {
				f.PrintError("Signature does not match")
			}
			f.PrintText("Computed: " + computed)
			f.PrintText("Provided: " + provided)
		}, result)
	},
}

func init() {
	// Create command flags
	webhooksCreateCmd.Flags().StringVar(&webhooksURLFlag, "url", "", "Webhook URL (required)")
	webhooksCreateCmd.Flags().StringSliceVar(&webhooksEventsFlag, "events", []string{}, "Event types to subscribe to (required, can be specified multiple times)")
	webhooksCreateCmd.Flags().StringVar(&webhooksDescriptionFlag, "description", "", "Webhook description (optional)")

	// Update command flags
	webhooksUpdateCmd.Flags().StringVar(&webhooksURLFlag, "url", "", "Webhook URL")
	webhooksUpdateCmd.Flags().StringSliceVar(&webhooksEventsFlag, "events", []string{}, "Event types to subscribe to (can be specified multiple times)")
	webhooksUpdateCmd.Flags().StringVar(&webhooksDescriptionFlag, "description", "", "Webhook description")

	// Add subcommands
	webhooksCmd.AddCommand(webhooksListCmd)
	webhooksCmd.AddCommand(webhooksGetCmd)
	webhooksCmd.AddCommand(webhooksCreateCmd)
	webhooksCmd.AddCommand(webhooksUpdateCmd)
	webhooksCmd.AddCommand(webhooksEnableCmd)
	webhooksCmd.AddCommand(webhooksDisableCmd)
	webhooksCmd.AddCommand(webhooksDeleteCmd)
	webhooksCmd.AddCommand(webhooksEventTypesCmd)
	webhooksCmd.AddCommand(webhooksVerifyCmd)

	webhooksVerifyCmd.Flags().StringVar(&webhooksVerifySecretFlag, "secret", "", "Webhook secret (required)")
	webhooksVerifyCmd.Flags().StringVar(&webhooksVerifySignatureFlag, "signature", "", "Signature header or value (required)")
	webhooksVerifyCmd.Flags().StringVar(&webhooksVerifyPayloadFlag, "payload", "", "Raw payload string")
	webhooksVerifyCmd.Flags().StringVar(&webhooksVerifyPayloadFileFlag, "payload-file", "", "Path to payload file")
	webhooksVerifyCmd.Flags().StringVar(&webhooksVerifySignedPayloadFlag, "signed-payload", "", "Exact payload string to sign")
}

func computeHMACSHA256(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func extractSignatureValue(signature string) string {
	sig := strings.TrimSpace(signature)
	if strings.Contains(sig, ",") {
		parts := strings.Split(sig, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "v1=") {
				return strings.TrimSpace(strings.TrimPrefix(part, "v1="))
			}
		}
	}
	sig = strings.TrimPrefix(sig, "sha256=")
	sig = strings.TrimPrefix(sig, "v1=")
	return strings.TrimSpace(sig)
}
