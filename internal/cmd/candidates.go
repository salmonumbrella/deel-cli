package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var candidatesCmd = &cobra.Command{
	Use:   "candidates",
	Short: "Manage candidates",
	Long:  "Add and update candidates in the ATS system.",
}

var (
	candidateFirstNameFlag string
	candidateLastNameFlag  string
	candidateEmailFlag     string
	candidatePhoneFlag     string
	candidateStatusFlag    string
)

var candidatesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new candidate",
	Long:  "Create a new candidate in the ATS. Requires --first-name, --last-name, and --email flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if candidateFirstNameFlag == "" || candidateLastNameFlag == "" || candidateEmailFlag == "" {
			f.PrintError("--first-name, --last-name, and --email are required")
			return fmt.Errorf("missing required flags")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Candidate",
			Description: "Add candidate",
			Details: map[string]string{
				"FirstName": candidateFirstNameFlag,
				"LastName":  candidateLastNameFlag,
				"Email":     candidateEmailFlag,
				"Phone":     candidatePhoneFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		candidate, err := client.AddCandidate(cmd.Context(), api.AddCandidateParams{
			FirstName: candidateFirstNameFlag,
			LastName:  candidateLastNameFlag,
			Email:     candidateEmailFlag,
			Phone:     candidatePhoneFlag,
		})
		if err != nil {
			return HandleError(f, err, "add candidate")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Candidate added successfully")
			f.PrintText("ID:         " + candidate.ID)
			f.PrintText("Name:       " + candidate.FirstName + " " + candidate.LastName)
			f.PrintText("Email:      " + candidate.Email)
			if candidate.Phone != "" {
				f.PrintText("Phone:      " + candidate.Phone)
			}
			f.PrintText("Status:     " + candidate.Status)
			f.PrintText("Created:    " + candidate.CreatedAt)
		}, candidate)
	},
}

var candidatesUpdateCmd = &cobra.Command{
	Use:   "update <candidate-id>",
	Short: "Update a candidate",
	Long:  "Update an existing candidate. Optional: --first-name, --last-name, --email, --phone, --status flags.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		// Check if at least one flag was provided
		if !cmd.Flags().Changed("first-name") &&
			!cmd.Flags().Changed("last-name") &&
			!cmd.Flags().Changed("email") &&
			!cmd.Flags().Changed("phone") &&
			!cmd.Flags().Changed("status") {
			f.PrintError("At least one flag (--first-name, --last-name, --email, --phone, or --status) must be provided")
			return fmt.Errorf("no update flags provided")
		}

		details := map[string]string{
			"ID": args[0],
		}
		if cmd.Flags().Changed("first-name") {
			details["FirstName"] = candidateFirstNameFlag
		}
		if cmd.Flags().Changed("last-name") {
			details["LastName"] = candidateLastNameFlag
		}
		if cmd.Flags().Changed("email") {
			details["Email"] = candidateEmailFlag
		}
		if cmd.Flags().Changed("phone") {
			details["Phone"] = candidatePhoneFlag
		}
		if cmd.Flags().Changed("status") {
			details["Status"] = candidateStatusFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Candidate",
			Description: "Update candidate",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.UpdateCandidateParams{}
		if cmd.Flags().Changed("first-name") {
			params.FirstName = candidateFirstNameFlag
		}
		if cmd.Flags().Changed("last-name") {
			params.LastName = candidateLastNameFlag
		}
		if cmd.Flags().Changed("email") {
			params.Email = candidateEmailFlag
		}
		if cmd.Flags().Changed("phone") {
			params.Phone = candidatePhoneFlag
		}
		if cmd.Flags().Changed("status") {
			params.Status = candidateStatusFlag
		}

		candidate, err := client.UpdateCandidate(cmd.Context(), args[0], params)
		if err != nil {
			return HandleError(f, err, "update candidate")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Candidate updated successfully")
			f.PrintText("ID:         " + candidate.ID)
			f.PrintText("Name:       " + candidate.FirstName + " " + candidate.LastName)
			f.PrintText("Email:      " + candidate.Email)
			if candidate.Phone != "" {
				f.PrintText("Phone:      " + candidate.Phone)
			}
			f.PrintText("Status:     " + candidate.Status)
		}, candidate)
	},
}

func init() {
	// Add command flags
	candidatesAddCmd.Flags().StringVar(&candidateFirstNameFlag, "first-name", "", "First name (required)")
	candidatesAddCmd.Flags().StringVar(&candidateLastNameFlag, "last-name", "", "Last name (required)")
	candidatesAddCmd.Flags().StringVar(&candidateEmailFlag, "email", "", "Email address (required)")
	candidatesAddCmd.Flags().StringVar(&candidatePhoneFlag, "phone", "", "Phone number (optional)")

	// Update command flags
	candidatesUpdateCmd.Flags().StringVar(&candidateFirstNameFlag, "first-name", "", "First name")
	candidatesUpdateCmd.Flags().StringVar(&candidateLastNameFlag, "last-name", "", "Last name")
	candidatesUpdateCmd.Flags().StringVar(&candidateEmailFlag, "email", "", "Email address")
	candidatesUpdateCmd.Flags().StringVar(&candidatePhoneFlag, "phone", "", "Phone number")
	candidatesUpdateCmd.Flags().StringVar(&candidateStatusFlag, "status", "", "Candidate status")

	// Add subcommands
	candidatesCmd.AddCommand(candidatesAddCmd)
	candidatesCmd.AddCommand(candidatesUpdateCmd)
}
