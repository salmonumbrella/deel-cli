package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var peopleCmd = &cobra.Command{
	Use:   "people",
	Short: "Manage people and workers",
	Long:  "List, search, and view details of people in your Deel organization.",
}

var (
	peopleEmailFlag string
)

var peopleListCmd = NewListCommand(ListConfig[api.Person]{
	Use:          "list",
	Short:        "List all people",
	Headers:      []string{"ID", "NAME", "EMAIL", "JOB TITLE", "STATUS"},
	EmptyMessage: "No people found.",
	RowFunc: func(p api.Person) []string {
		return []string{
			p.HRISProfileID,
			p.FirstName + " " + p.LastName,
			p.Email,
			p.JobTitle,
			p.Status,
		}
	},
	Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.Person], error) {
		resp, err := client.ListPeople(ctx, api.PeopleListParams{
			Limit:  pageSize,
			Cursor: "", // TODO: implement cursor-based pagination
		})
		if err != nil {
			return ListResult[api.Person]{}, err
		}
		return ListResult[api.Person]{
			Items:   resp.Data,
			HasMore: resp.Page.Next != "",
		}, nil
	},
}, getClient)

var peopleGetCmd = &cobra.Command{
	Use:   "get <hris-profile-id>",
	Short: "Get person details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		person, err := client.GetPerson(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get person: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Name:       " + person.FirstName + " " + person.LastName)
			f.PrintText("Email:      " + person.Email)
			f.PrintText("Job Title:  " + person.JobTitle)
			f.PrintText("Department: " + person.Department)
			f.PrintText("Status:     " + person.Status)
			f.PrintText("Country:    " + person.Country)
			f.PrintText("Start Date: " + person.StartDate)
		}, person)
	},
}

var peopleSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for a person",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if peopleEmailFlag == "" {
			f.PrintError("--email is required")
			return fmt.Errorf("--email flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		person, err := client.SearchPeopleByEmail(cmd.Context(), peopleEmailFlag)
		if err != nil {
			f.PrintError("Failed to search: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("Found: " + person.FirstName + " " + person.LastName)
			f.PrintText("ID:    " + person.HRISProfileID)
			f.PrintText("Email: " + person.Email)
		}, person)
	},
}

var customFieldsCmd = &cobra.Command{
	Use:   "custom-fields",
	Short: "Manage custom fields",
}

var customFieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom fields",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		fields, err := client.ListCustomFields(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list custom fields: %v", err)
			return err
		}

		return f.Output(func() {
			if len(fields) == 0 {
				f.PrintText("No custom fields found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE")
			for _, field := range fields {
				table.AddRow(field.ID, field.Name, field.Type)
			}
			table.Render()
		}, fields)
	},
}

var customFieldsGetCmd = &cobra.Command{
	Use:   "get <field-id>",
	Short: "Get custom field details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		field, err := client.GetCustomField(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get custom field: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:    " + field.ID)
			f.PrintText("Name:  " + field.Name)
			f.PrintText("Type:  " + field.Type)
			f.PrintText("Value: " + field.Value)
		}, field)
	},
}

// Flags for people create command
var (
	peopleCreateEmailFlag     string
	peopleCreateFirstNameFlag string
	peopleCreateLastNameFlag  string
	peopleCreateTypeFlag      string
	peopleCreateCountryFlag   string
)

var peopleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create person",
	Long:  "Create a new person. Requires --email, --first-name, --last-name, --type, and --country flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if peopleCreateEmailFlag == "" {
			f.PrintError("--email flag is required")
			return fmt.Errorf("--email flag is required")
		}
		if peopleCreateFirstNameFlag == "" {
			f.PrintError("--first-name flag is required")
			return fmt.Errorf("--first-name flag is required")
		}
		if peopleCreateLastNameFlag == "" {
			f.PrintError("--last-name flag is required")
			return fmt.Errorf("--last-name flag is required")
		}
		if peopleCreateTypeFlag == "" {
			f.PrintError("--type flag is required")
			return fmt.Errorf("--type flag is required")
		}
		if peopleCreateCountryFlag == "" {
			f.PrintError("--country flag is required")
			return fmt.Errorf("--country flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreatePersonParams{
			Email:     peopleCreateEmailFlag,
			FirstName: peopleCreateFirstNameFlag,
			LastName:  peopleCreateLastNameFlag,
			Type:      peopleCreateTypeFlag,
			Country:   peopleCreateCountryFlag,
		}

		person, err := client.CreatePerson(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create person: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Person created successfully")
			f.PrintText("ID:         " + person.ID)
			f.PrintText("Email:      " + person.Email)
			f.PrintText("First Name: " + person.FirstName)
			f.PrintText("Last Name:  " + person.LastName)
			f.PrintText("Type:       " + person.Type)
			f.PrintText("Country:    " + person.Country)
			f.PrintText("Status:     " + person.Status)
			f.PrintText("Created:    " + person.CreatedAt)
		}, person)
	},
}

// Flags for people update command
var (
	peopleUpdateFirstNameFlag   string
	peopleUpdateLastNameFlag    string
	peopleUpdatePhoneFlag       string
	peopleUpdateNationalityFlag string
)

var peopleUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update personal info",
	Long:  "Update personal information for a person. Optional flags: --first-name, --last-name, --phone, --nationality.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		info := api.PersonalInfo{
			FirstName:   peopleUpdateFirstNameFlag,
			LastName:    peopleUpdateLastNameFlag,
			Phone:       peopleUpdatePhoneFlag,
			Nationality: peopleUpdateNationalityFlag,
		}

		updated, err := client.UpdatePersonalInfo(cmd.Context(), args[0], info)
		if err != nil {
			f.PrintError("Failed to update personal info: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Personal info updated successfully")
			if updated.ID != "" {
				f.PrintText("ID:          " + updated.ID)
			}
			if updated.FirstName != "" {
				f.PrintText("First Name:  " + updated.FirstName)
			}
			if updated.LastName != "" {
				f.PrintText("Last Name:   " + updated.LastName)
			}
			if updated.Phone != "" {
				f.PrintText("Phone:       " + updated.Phone)
			}
			if updated.Nationality != "" {
				f.PrintText("Nationality: " + updated.Nationality)
			}
		}, updated)
	},
}

// Adjustments subcommand
var adjustmentsCmd = &cobra.Command{
	Use:   "adjustments",
	Short: "Manage adjustments",
	Long:  "List, create, and delete contract adjustments (bonuses, deductions, expenses).",
}

// Flags for adjustments list command
var (
	adjustmentsListContractIDFlag string
	adjustmentsListCategoryIDFlag string
)

var adjustmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List adjustments",
	Long:  "List adjustments with optional filters. Optional flags: --contract-id, --category-id.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.ListAdjustmentsParams{
			ContractID: adjustmentsListContractIDFlag,
			CategoryID: adjustmentsListCategoryIDFlag,
		}

		adjustments, err := client.ListAdjustments(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to list adjustments: %v", err)
			return err
		}

		return f.Output(func() {
			if len(adjustments) == 0 {
				f.PrintText("No adjustments found")
				return
			}
			table := f.NewTable("ID", "CONTRACT ID", "CATEGORY ID", "AMOUNT", "CURRENCY", "DATE", "STATUS")
			for _, adj := range adjustments {
				table.AddRow(
					adj.ID,
					adj.ContractID,
					adj.CategoryID,
					fmt.Sprintf("%.2f", adj.Amount),
					adj.Currency,
					adj.Date,
					adj.Status,
				)
			}
			table.Render()
		}, adjustments)
	},
}

// Flags for adjustments create command
var (
	adjustmentsCreateContractIDFlag  string
	adjustmentsCreateCategoryIDFlag  string
	adjustmentsCreateAmountFlag      string
	adjustmentsCreateCurrencyFlag    string
	adjustmentsCreateDescriptionFlag string
	adjustmentsCreateDateFlag        string
)

var adjustmentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create adjustment",
	Long:  "Create a new adjustment. Requires --contract-id, --category-id, --amount, --currency, --description, and --date flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if adjustmentsCreateContractIDFlag == "" {
			f.PrintError("--contract-id flag is required")
			return fmt.Errorf("--contract-id flag is required")
		}
		if adjustmentsCreateCategoryIDFlag == "" {
			f.PrintError("--category-id flag is required")
			return fmt.Errorf("--category-id flag is required")
		}
		if adjustmentsCreateAmountFlag == "" {
			f.PrintError("--amount flag is required")
			return fmt.Errorf("--amount flag is required")
		}
		if adjustmentsCreateCurrencyFlag == "" {
			f.PrintError("--currency flag is required")
			return fmt.Errorf("--currency flag is required")
		}
		if adjustmentsCreateDescriptionFlag == "" {
			f.PrintError("--description flag is required")
			return fmt.Errorf("--description flag is required")
		}
		if adjustmentsCreateDateFlag == "" {
			f.PrintError("--date flag is required")
			return fmt.Errorf("--date flag is required")
		}

		// Parse amount
		amount, err := strconv.ParseFloat(adjustmentsCreateAmountFlag, 64)
		if err != nil {
			f.PrintError("Invalid --amount value: %v", err)
			return fmt.Errorf("invalid --amount value: %w", err)
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateAdjustmentParams{
			ContractID:  adjustmentsCreateContractIDFlag,
			CategoryID:  adjustmentsCreateCategoryIDFlag,
			Amount:      amount,
			Currency:    adjustmentsCreateCurrencyFlag,
			Description: adjustmentsCreateDescriptionFlag,
			Date:        adjustmentsCreateDateFlag,
		}

		adjustment, err := client.CreateAdjustment(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create adjustment: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Adjustment created successfully")
			f.PrintText("ID:          " + adjustment.ID)
			f.PrintText("Contract ID: " + adjustment.ContractID)
			f.PrintText("Category ID: " + adjustment.CategoryID)
			f.PrintText(fmt.Sprintf("Amount:      %.2f %s", adjustment.Amount, adjustment.Currency))
			f.PrintText("Description: " + adjustment.Description)
			f.PrintText("Date:        " + adjustment.Date)
			f.PrintText("Status:      " + adjustment.Status)
			f.PrintText("Created:     " + adjustment.CreatedAt)
		}, adjustment)
	},
}

var adjustmentsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete adjustment",
	Long:  "Delete an adjustment by ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		err = client.DeleteAdjustment(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to delete adjustment: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Adjustment deleted successfully")
			f.PrintText("ID: " + args[0])
		}, map[string]string{"id": args[0], "status": "deleted"})
	},
}

// Managers subcommand
var managersCmd = &cobra.Command{
	Use:   "managers",
	Short: "Manage managers",
	Long:  "List and create managers in your Deel organization.",
}

var managersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List managers",
	Long:  "List all managers in your organization.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		managers, err := client.ListManagers(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list managers: %v", err)
			return err
		}

		return f.Output(func() {
			if len(managers) == 0 {
				f.PrintText("No managers found")
				return
			}
			table := f.NewTable("ID", "EMAIL", "FIRST NAME", "LAST NAME", "ROLE", "STATUS")
			for _, mgr := range managers {
				table.AddRow(
					mgr.ID,
					mgr.Email,
					mgr.FirstName,
					mgr.LastName,
					mgr.Role,
					mgr.Status,
				)
			}
			table.Render()
		}, managers)
	},
}

// Flags for managers create command
var (
	managersCreateEmailFlag     string
	managersCreateFirstNameFlag string
	managersCreateLastNameFlag  string
	managersCreateRoleFlag      string
)

var managersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create manager",
	Long:  "Create a new manager. Requires --email, --first-name, --last-name, and --role flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if managersCreateEmailFlag == "" {
			f.PrintError("--email flag is required")
			return fmt.Errorf("--email flag is required")
		}
		if managersCreateFirstNameFlag == "" {
			f.PrintError("--first-name flag is required")
			return fmt.Errorf("--first-name flag is required")
		}
		if managersCreateLastNameFlag == "" {
			f.PrintError("--last-name flag is required")
			return fmt.Errorf("--last-name flag is required")
		}
		if managersCreateRoleFlag == "" {
			f.PrintError("--role flag is required")
			return fmt.Errorf("--role flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateManagerParams{
			Email:     managersCreateEmailFlag,
			FirstName: managersCreateFirstNameFlag,
			LastName:  managersCreateLastNameFlag,
			Role:      managersCreateRoleFlag,
		}

		manager, err := client.CreateManager(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create manager: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Manager created successfully")
			f.PrintText("ID:         " + manager.ID)
			f.PrintText("Email:      " + manager.Email)
			f.PrintText("First Name: " + manager.FirstName)
			f.PrintText("Last Name:  " + manager.LastName)
			f.PrintText("Role:       " + manager.Role)
			f.PrintText("Status:     " + manager.Status)
			f.PrintText("Created:    " + manager.CreatedAt)
		}, manager)
	},
}

// Relations subcommand
var relationsCmd = &cobra.Command{
	Use:   "relations",
	Short: "Manage worker relations",
	Long:  "List, create, and delete worker-manager relations.",
}

var relationsListCmd = &cobra.Command{
	Use:   "list <profile-id>",
	Short: "List worker relations",
	Long:  "List all worker relations for a given profile.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		relations, err := client.ListWorkerRelations(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to list worker relations: %v", err)
			return err
		}

		return f.Output(func() {
			if len(relations) == 0 {
				f.PrintText("No worker relations found")
				return
			}
			table := f.NewTable("ID", "PROFILE ID", "MANAGER ID", "RELATION TYPE", "START DATE", "STATUS")
			for _, rel := range relations {
				table.AddRow(
					rel.ID,
					rel.ProfileID,
					rel.ManagerID,
					rel.RelationType,
					rel.StartDate,
					rel.Status,
				)
			}
			table.Render()
		}, relations)
	},
}

// Flags for relations create command
var (
	relationsCreateProfileIDFlag    string
	relationsCreateManagerIDFlag    string
	relationsCreateRelationTypeFlag string
	relationsCreateStartDateFlag    string
)

var relationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create relation",
	Long:  "Create a new worker relation. Requires --profile-id, --manager-id, --relation-type, and --start-date flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if relationsCreateProfileIDFlag == "" {
			f.PrintError("--profile-id flag is required")
			return fmt.Errorf("--profile-id flag is required")
		}
		if relationsCreateManagerIDFlag == "" {
			f.PrintError("--manager-id flag is required")
			return fmt.Errorf("--manager-id flag is required")
		}
		if relationsCreateRelationTypeFlag == "" {
			f.PrintError("--relation-type flag is required")
			return fmt.Errorf("--relation-type flag is required")
		}
		if relationsCreateStartDateFlag == "" {
			f.PrintError("--start-date flag is required")
			return fmt.Errorf("--start-date flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.CreateWorkerRelationParams{
			ProfileID:    relationsCreateProfileIDFlag,
			ManagerID:    relationsCreateManagerIDFlag,
			RelationType: relationsCreateRelationTypeFlag,
			StartDate:    relationsCreateStartDateFlag,
		}

		relation, err := client.CreateWorkerRelation(cmd.Context(), params)
		if err != nil {
			f.PrintError("Failed to create worker relation: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Worker relation created successfully")
			f.PrintText("ID:            " + relation.ID)
			f.PrintText("Profile ID:    " + relation.ProfileID)
			f.PrintText("Manager ID:    " + relation.ManagerID)
			f.PrintText("Relation Type: " + relation.RelationType)
			f.PrintText("Start Date:    " + relation.StartDate)
			f.PrintText("Status:        " + relation.Status)
			f.PrintText("Created:       " + relation.CreatedAt)
		}, relation)
	},
}

var relationsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete relation",
	Long:  "Delete a worker relation by ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		err = client.DeleteWorkerRelation(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to delete worker relation: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Worker relation deleted successfully")
			f.PrintText("ID: " + args[0])
		}, map[string]string{"id": args[0], "status": "deleted"})
	},
}

func init() {
	// Note: peopleListCmd flags are added by NewListCommand (--page, --limit)
	peopleSearchCmd.Flags().StringVar(&peopleEmailFlag, "email", "", "Email to search for")

	// People create command flags
	peopleCreateCmd.Flags().StringVar(&peopleCreateEmailFlag, "email", "", "Email (required)")
	peopleCreateCmd.Flags().StringVar(&peopleCreateFirstNameFlag, "first-name", "", "First name (required)")
	peopleCreateCmd.Flags().StringVar(&peopleCreateLastNameFlag, "last-name", "", "Last name (required)")
	peopleCreateCmd.Flags().StringVar(&peopleCreateTypeFlag, "type", "", "Person type (required)")
	peopleCreateCmd.Flags().StringVar(&peopleCreateCountryFlag, "country", "", "Country code (required)")

	// People update command flags
	peopleUpdateCmd.Flags().StringVar(&peopleUpdateFirstNameFlag, "first-name", "", "First name (optional)")
	peopleUpdateCmd.Flags().StringVar(&peopleUpdateLastNameFlag, "last-name", "", "Last name (optional)")
	peopleUpdateCmd.Flags().StringVar(&peopleUpdatePhoneFlag, "phone", "", "Phone number (optional)")
	peopleUpdateCmd.Flags().StringVar(&peopleUpdateNationalityFlag, "nationality", "", "Nationality (optional)")

	// Adjustments list command flags
	adjustmentsListCmd.Flags().StringVar(&adjustmentsListContractIDFlag, "contract-id", "", "Contract ID (optional)")
	adjustmentsListCmd.Flags().StringVar(&adjustmentsListCategoryIDFlag, "category-id", "", "Category ID (optional)")

	// Adjustments create command flags
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateContractIDFlag, "contract-id", "", "Contract ID (required)")
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateCategoryIDFlag, "category-id", "", "Category ID (required)")
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateAmountFlag, "amount", "", "Amount (required)")
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateCurrencyFlag, "currency", "", "Currency code (required)")
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateDescriptionFlag, "description", "", "Description (required)")
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateDateFlag, "date", "", "Date YYYY-MM-DD (required)")

	// Managers create command flags
	managersCreateCmd.Flags().StringVar(&managersCreateEmailFlag, "email", "", "Email (required)")
	managersCreateCmd.Flags().StringVar(&managersCreateFirstNameFlag, "first-name", "", "First name (required)")
	managersCreateCmd.Flags().StringVar(&managersCreateLastNameFlag, "last-name", "", "Last name (required)")
	managersCreateCmd.Flags().StringVar(&managersCreateRoleFlag, "role", "", "Role (required)")

	// Relations create command flags
	relationsCreateCmd.Flags().StringVar(&relationsCreateProfileIDFlag, "profile-id", "", "Profile ID (required)")
	relationsCreateCmd.Flags().StringVar(&relationsCreateManagerIDFlag, "manager-id", "", "Manager ID (required)")
	relationsCreateCmd.Flags().StringVar(&relationsCreateRelationTypeFlag, "relation-type", "", "Relation type (required)")
	relationsCreateCmd.Flags().StringVar(&relationsCreateStartDateFlag, "start-date", "", "Start date YYYY-MM-DD (required)")

	// Add subcommands to adjustments
	adjustmentsCmd.AddCommand(adjustmentsListCmd)
	adjustmentsCmd.AddCommand(adjustmentsCreateCmd)
	adjustmentsCmd.AddCommand(adjustmentsDeleteCmd)

	// Add subcommands to managers
	managersCmd.AddCommand(managersListCmd)
	managersCmd.AddCommand(managersCreateCmd)

	// Add subcommands to relations
	relationsCmd.AddCommand(relationsListCmd)
	relationsCmd.AddCommand(relationsCreateCmd)
	relationsCmd.AddCommand(relationsDeleteCmd)

	customFieldsCmd.AddCommand(customFieldsListCmd)
	customFieldsCmd.AddCommand(customFieldsGetCmd)

	peopleCmd.AddCommand(peopleListCmd)
	peopleCmd.AddCommand(peopleGetCmd)
	peopleCmd.AddCommand(peopleSearchCmd)
	peopleCmd.AddCommand(peopleCreateCmd)
	peopleCmd.AddCommand(peopleUpdateCmd)
	peopleCmd.AddCommand(customFieldsCmd)
	peopleCmd.AddCommand(adjustmentsCmd)
	peopleCmd.AddCommand(managersCmd)
	peopleCmd.AddCommand(relationsCmd)
}
