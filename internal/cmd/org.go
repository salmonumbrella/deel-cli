package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "View organization information",
	Long:  "View organization details, structure, and legal entities.",
}

var orgGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get organization details",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		org, err := client.GetOrganization(cmd.Context())
		if err != nil {
			f.PrintError("Failed to get organization: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:       " + org.ID)
			f.PrintText("Name:     " + org.Name)
			f.PrintText("Country:  " + org.Country)
			f.PrintText("Industry: " + org.Industry)
			f.PrintText("Size:     " + org.Size)
		}, org)
	},
}

var orgStructuresCmd = &cobra.Command{
	Use:   "structures",
	Short: "View organization structure",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		structures, err := client.GetOrgStructures(cmd.Context())
		if err != nil {
			f.PrintError("Failed to get structures: %v", err)
			return err
		}

		return f.Output(func() {
			if len(structures) == 0 {
				f.PrintText("No org structures found.")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "PARENT")
			for _, s := range structures {
				table.AddRow(s.ID, s.Name, s.Type, s.ParentID)
			}
			table.Render()
		}, structures)
	},
}

var orgEntitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "List legal entities",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		entities, err := client.ListLegalEntities(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list entities: %v", err)
			return err
		}

		return f.Output(func() {
			if len(entities) == 0 {
				f.PrintText("No legal entities found.")
				return
			}
			table := f.NewTable("ID", "NAME", "COUNTRY", "TYPE", "STATUS")
			for _, e := range entities {
				table.AddRow(e.ID, e.Name, e.Country, e.Type, e.Status)
			}
			table.Render()
		}, entities)
	},
}

// Groups commands
var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage organization groups",
	Long:  "List, view, create, update, delete, and clone groups in your Deel organization.",
}

var (
	groupNameFlag        string
	groupDescriptionFlag string
)

var groupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		groups, err := client.ListGroups(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list groups: %v", err)
			return err
		}

		return f.Output(func() {
			if len(groups) == 0 {
				f.PrintText("No groups found.")
				return
			}
			table := f.NewTable("ID", "NAME", "DESCRIPTION", "MEMBERS", "CREATED")
			for _, g := range groups {
				desc := g.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				table.AddRow(g.ID, g.Name, desc, fmt.Sprintf("%d", g.MemberCount), g.CreatedAt)
			}
			table.Render()
		}, groups)
	},
}

var groupsGetCmd = &cobra.Command{
	Use:   "get <group-id>",
	Short: "Get group details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		group, err := client.GetGroup(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get group: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:          " + group.ID)
			f.PrintText("Name:        " + group.Name)
			if group.Description != "" {
				f.PrintText("Description: " + group.Description)
			}
			f.PrintText("Members:     " + fmt.Sprintf("%d", group.MemberCount))
			f.PrintText("Created:     " + group.CreatedAt)
		}, group)
	},
}

var groupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new group",
	Long:  "Create a new group. Requires --name flag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if groupNameFlag == "" {
			f.PrintError("--name flag is required")
			return fmt.Errorf("--name flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Group",
			Description: "Create group",
			Details: map[string]string{
				"Name":        groupNameFlag,
				"Description": groupDescriptionFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		group, err := client.CreateGroup(cmd.Context(), api.CreateGroupParams{
			Name:        groupNameFlag,
			Description: groupDescriptionFlag,
		})
		if err != nil {
			f.PrintError("Failed to create group: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Group created successfully")
			f.PrintText("ID:          " + group.ID)
			f.PrintText("Name:        " + group.Name)
			if group.Description != "" {
				f.PrintText("Description: " + group.Description)
			}
			f.PrintText("Members:     " + fmt.Sprintf("%d", group.MemberCount))
		}, group)
	},
}

var groupsUpdateCmd = &cobra.Command{
	Use:   "update <group-id>",
	Short: "Update a group",
	Long:  "Update an existing group. Optional: --name, --description flags.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") {
			f.PrintError("At least one flag (--name or --description) must be provided")
			return fmt.Errorf("no update flags provided")
		}

		details := map[string]string{
			"ID": args[0],
		}
		if cmd.Flags().Changed("name") {
			details["Name"] = groupNameFlag
		}
		if cmd.Flags().Changed("description") {
			details["Description"] = groupDescriptionFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Group",
			Description: "Update group",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.UpdateGroupParams{}
		if cmd.Flags().Changed("name") {
			params.Name = groupNameFlag
		}
		if cmd.Flags().Changed("description") {
			params.Description = groupDescriptionFlag
		}

		group, err := client.UpdateGroup(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to update group: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Group updated successfully")
			f.PrintText("ID:          " + group.ID)
			f.PrintText("Name:        " + group.Name)
			if group.Description != "" {
				f.PrintText("Description: " + group.Description)
			}
			f.PrintText("Members:     " + fmt.Sprintf("%d", group.MemberCount))
		}, group)
	},
}

var groupsDeleteCmd = &cobra.Command{
	Use:   "delete <group-id>",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		err = client.DeleteGroup(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to delete group: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Group deleted successfully")
		}, map[string]string{"status": "deleted", "id": args[0]})
	},
}

var groupsCloneCmd = &cobra.Command{
	Use:   "clone <group-id>",
	Short: "Clone a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		group, err := client.CloneGroup(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to clone group: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Group cloned successfully")
			f.PrintText("ID:          " + group.ID)
			f.PrintText("Name:        " + group.Name)
			if group.Description != "" {
				f.PrintText("Description: " + group.Description)
			}
			f.PrintText("Members:     " + fmt.Sprintf("%d", group.MemberCount))
		}, group)
	},
}

// Legal entities commands
var legalEntitiesCmd = &cobra.Command{
	Use:   "legal-entities",
	Short: "Manage legal entities",
	Long:  "List, create, update, delete legal entities and view payroll settings.",
}

var (
	entityNameFlag               string
	entityCountryFlag            string
	entityTypeFlag               string
	entityRegistrationNumberFlag string
)

var legalEntitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all legal entities",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		entities, err := client.ListLegalEntities(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list legal entities: %v", err)
			return err
		}

		return f.Output(func() {
			if len(entities) == 0 {
				f.PrintText("No legal entities found.")
				return
			}
			table := f.NewTable("ID", "NAME", "COUNTRY", "TYPE", "STATUS", "REG NUMBER")
			for _, e := range entities {
				table.AddRow(e.ID, e.Name, e.Country, e.Type, e.Status, e.RegistrationNumber)
			}
			table.Render()
		}, entities)
	},
}

var legalEntitiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new legal entity",
	Long:  "Create a new legal entity. Requires --name, --country, and --type flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if entityNameFlag == "" {
			f.PrintError("--name flag is required")
			return fmt.Errorf("--name flag is required")
		}
		if entityCountryFlag == "" {
			f.PrintError("--country flag is required")
			return fmt.Errorf("--country flag is required")
		}
		if entityTypeFlag == "" {
			f.PrintError("--type flag is required")
			return fmt.Errorf("--type flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "LegalEntity",
			Description: "Create legal entity",
			Details: map[string]string{
				"Name":      entityNameFlag,
				"Country":   entityCountryFlag,
				"Type":      entityTypeFlag,
				"RegNumber": entityRegistrationNumberFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		entity, err := client.CreateLegalEntity(cmd.Context(), api.CreateLegalEntityParams{
			Name:               entityNameFlag,
			Country:            entityCountryFlag,
			Type:               entityTypeFlag,
			RegistrationNumber: entityRegistrationNumberFlag,
		})
		if err != nil {
			f.PrintError("Failed to create legal entity: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Legal entity created successfully")
			f.PrintText("ID:       " + entity.ID)
			f.PrintText("Name:     " + entity.Name)
			f.PrintText("Country:  " + entity.Country)
			f.PrintText("Type:     " + entity.Type)
			f.PrintText("Status:   " + entity.Status)
			if entity.RegistrationNumber != "" {
				f.PrintText("Reg Num:  " + entity.RegistrationNumber)
			}
		}, entity)
	},
}

var legalEntitiesUpdateCmd = &cobra.Command{
	Use:   "update <entity-id>",
	Short: "Update a legal entity",
	Long:  "Update an existing legal entity. Optional: --name, --type, --reg-number flags.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("type") && !cmd.Flags().Changed("reg-number") {
			f.PrintError("At least one flag (--name, --type, or --reg-number) must be provided")
			return fmt.Errorf("no update flags provided")
		}

		details := map[string]string{
			"ID": args[0],
		}
		if cmd.Flags().Changed("name") {
			details["Name"] = entityNameFlag
		}
		if cmd.Flags().Changed("type") {
			details["Type"] = entityTypeFlag
		}
		if cmd.Flags().Changed("reg-number") {
			details["RegNumber"] = entityRegistrationNumberFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "LegalEntity",
			Description: "Update legal entity",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		params := api.UpdateLegalEntityParams{}
		if cmd.Flags().Changed("name") {
			params.Name = entityNameFlag
		}
		if cmd.Flags().Changed("type") {
			params.Type = entityTypeFlag
		}
		if cmd.Flags().Changed("reg-number") {
			params.RegistrationNumber = entityRegistrationNumberFlag
		}

		entity, err := client.UpdateLegalEntity(cmd.Context(), args[0], params)
		if err != nil {
			f.PrintError("Failed to update legal entity: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Legal entity updated successfully")
			f.PrintText("ID:       " + entity.ID)
			f.PrintText("Name:     " + entity.Name)
			f.PrintText("Country:  " + entity.Country)
			f.PrintText("Type:     " + entity.Type)
			f.PrintText("Status:   " + entity.Status)
			if entity.RegistrationNumber != "" {
				f.PrintText("Reg Num:  " + entity.RegistrationNumber)
			}
		}, entity)
	},
}

var legalEntitiesDeleteCmd = &cobra.Command{
	Use:   "delete <entity-id>",
	Short: "Delete a legal entity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "LegalEntity",
			Description: "Delete legal entity",
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

		err = client.DeleteLegalEntity(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to delete legal entity: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Legal entity deleted successfully")
		}, map[string]string{"status": "deleted", "id": args[0]})
	},
}

var legalEntitiesPayrollSettingsCmd = &cobra.Command{
	Use:   "payroll-settings <entity-id>",
	Short: "Get payroll settings for a legal entity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		settings, err := client.GetPayrollSettings(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get payroll settings: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:               " + settings.ID)
			f.PrintText("Legal Entity ID:  " + settings.LegalEntityID)
			f.PrintText("Frequency:        " + settings.PayrollFrequency)
			f.PrintText("Payment Method:   " + settings.PaymentMethod)
			f.PrintText("Currency:         " + settings.Currency)
			if settings.TaxID != "" {
				f.PrintText("Tax ID:           " + settings.TaxID)
			}
			if settings.BankAccount != "" {
				f.PrintText("Bank Account:     " + settings.BankAccount)
			}
			if settings.PayrollProvider != "" {
				f.PrintText("Provider:         " + settings.PayrollProvider)
			}
			f.PrintText("Auto Approval:    " + fmt.Sprintf("%t", settings.AutoApproval))
			if settings.NotificationEmail != "" {
				f.PrintText("Notification:     " + settings.NotificationEmail)
			}
		}, settings)
	},
}

// Lookups commands
var lookupsCmd = &cobra.Command{
	Use:   "lookups",
	Short: "List lookup data",
	Long:  "List currencies, countries, job titles, seniority levels, and time off types.",
}

var lookupsCurrenciesCmd = &cobra.Command{
	Use:   "currencies",
	Short: "List available currencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		currencies, err := client.ListCurrencies(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list currencies: %v", err)
			return err
		}

		return f.Output(func() {
			if len(currencies) == 0 {
				f.PrintText("No currencies found.")
				return
			}
			table := f.NewTable("CODE", "NAME", "SYMBOL")
			for _, c := range currencies {
				table.AddRow(c.Code, c.Name, c.Symbol)
			}
			table.Render()
		}, currencies)
	},
}

var lookupsCountriesCmd = &cobra.Command{
	Use:   "countries",
	Short: "List available countries",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		countries, err := client.ListCountries(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list countries: %v", err)
			return err
		}

		return f.Output(func() {
			if len(countries) == 0 {
				f.PrintText("No countries found.")
				return
			}
			table := f.NewTable("CODE", "NAME")
			for _, c := range countries {
				table.AddRow(c.Code, c.Name)
			}
			table.Render()
		}, countries)
	},
}

var lookupsJobTitlesCmd = &cobra.Command{
	Use:   "job-titles",
	Short: "List available job titles",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		jobTitles, err := client.ListJobTitles(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list job titles: %v", err)
			return err
		}

		return f.Output(func() {
			if len(jobTitles) == 0 {
				f.PrintText("No job titles found.")
				return
			}
			table := f.NewTable("ID", "NAME")
			for _, jt := range jobTitles {
				table.AddRow(jt.ID, jt.Name)
			}
			table.Render()
		}, jobTitles)
	},
}

var lookupsSeniorityLevelsCmd = &cobra.Command{
	Use:   "seniority-levels",
	Short: "List available seniority levels",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		levels, err := client.ListSeniorityLevels(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list seniority levels: %v", err)
			return err
		}

		return f.Output(func() {
			if len(levels) == 0 {
				f.PrintText("No seniority levels found.")
				return
			}
			table := f.NewTable("ID", "NAME")
			for _, l := range levels {
				table.AddRow(l.ID, l.Name)
			}
			table.Render()
		}, levels)
	},
}

var lookupsTimeOffTypesCmd = &cobra.Command{
	Use:   "time-off-types",
	Short: "List available time off types",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		types, err := client.ListTimeOffTypes(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list time off types: %v", err)
			return err
		}

		return f.Output(func() {
			if len(types) == 0 {
				f.PrintText("No time off types found.")
				return
			}
			table := f.NewTable("ID", "NAME")
			for _, t := range types {
				table.AddRow(t.ID, t.Name)
			}
			table.Render()
		}, types)
	},
}

func init() {
	// Groups command flags
	groupsCreateCmd.Flags().StringVar(&groupNameFlag, "name", "", "Group name (required)")
	groupsCreateCmd.Flags().StringVar(&groupDescriptionFlag, "description", "", "Group description (optional)")

	groupsUpdateCmd.Flags().StringVar(&groupNameFlag, "name", "", "Group name")
	groupsUpdateCmd.Flags().StringVar(&groupDescriptionFlag, "description", "", "Group description")

	// Add groups subcommands
	groupsCmd.AddCommand(groupsListCmd)
	groupsCmd.AddCommand(groupsGetCmd)
	groupsCmd.AddCommand(groupsCreateCmd)
	groupsCmd.AddCommand(groupsUpdateCmd)
	groupsCmd.AddCommand(groupsDeleteCmd)
	groupsCmd.AddCommand(groupsCloneCmd)

	// Legal entities command flags
	legalEntitiesCreateCmd.Flags().StringVar(&entityNameFlag, "name", "", "Entity name (required)")
	legalEntitiesCreateCmd.Flags().StringVar(&entityCountryFlag, "country", "", "Country code (required)")
	legalEntitiesCreateCmd.Flags().StringVar(&entityTypeFlag, "type", "", "Entity type (required)")
	legalEntitiesCreateCmd.Flags().StringVar(&entityRegistrationNumberFlag, "reg-number", "", "Registration number (optional)")

	legalEntitiesUpdateCmd.Flags().StringVar(&entityNameFlag, "name", "", "Entity name")
	legalEntitiesUpdateCmd.Flags().StringVar(&entityTypeFlag, "type", "", "Entity type")
	legalEntitiesUpdateCmd.Flags().StringVar(&entityRegistrationNumberFlag, "reg-number", "", "Registration number")

	// Add legal entities subcommands
	legalEntitiesCmd.AddCommand(legalEntitiesListCmd)
	legalEntitiesCmd.AddCommand(legalEntitiesCreateCmd)
	legalEntitiesCmd.AddCommand(legalEntitiesUpdateCmd)
	legalEntitiesCmd.AddCommand(legalEntitiesDeleteCmd)
	legalEntitiesCmd.AddCommand(legalEntitiesPayrollSettingsCmd)

	// Add lookups subcommands
	lookupsCmd.AddCommand(lookupsCurrenciesCmd)
	lookupsCmd.AddCommand(lookupsCountriesCmd)
	lookupsCmd.AddCommand(lookupsJobTitlesCmd)
	lookupsCmd.AddCommand(lookupsSeniorityLevelsCmd)
	lookupsCmd.AddCommand(lookupsTimeOffTypesCmd)

	// Add all commands to org
	orgCmd.AddCommand(orgGetCmd)
	orgCmd.AddCommand(orgStructuresCmd)
	orgCmd.AddCommand(orgEntitiesCmd)
	orgCmd.AddCommand(groupsCmd)
	orgCmd.AddCommand(legalEntitiesCmd)
	orgCmd.AddCommand(lookupsCmd)
}
