package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
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

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <webhook-id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
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
	webhooksCmd.AddCommand(webhooksDeleteCmd)
	webhooksCmd.AddCommand(webhooksEventTypesCmd)
}
