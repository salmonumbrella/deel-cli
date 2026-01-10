package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var peopleCmd = &cobra.Command{
	Use:   "people",
	Short: "Manage people and workers",
	Long: `List, search, and view details of people in your Deel organization.

FINDING SOMEONE:
  By name:   deel people search --name "Catherine Song"
  By email:  deel people search --email catherine@example.com
  List all:  deel people list --all

COMMON WORKFLOWS:
  Find person → get their contracts:
    1. deel people search --name "Song"     # Get their HRIS profile ID
    2. deel people get <profile-id>         # View full details
    3. deel contracts list | grep Song      # Find their contracts

  Find person → cancel contract:
    1. deel people search --name "Song"     # Get profile ID
    2. deel contracts list                  # Find contract ID by name
    3. deel contracts cancel <contract-id>  # Cancel the contract`,
}

var (
	peopleEmailFlag string
	peopleNameFlag  string
)

var (
	peopleLimitFlag  int
	peopleCursorFlag string
	peopleAllFlag    bool
)

var peopleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all people",
	Long: `List all people in your organization.

Tip: To find someone by name, use 'deel people search --name "Name"' instead.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, f, err := initClient("listing people")
		if err != nil {
			return err
		}

		people, page, hasMore, err := collectCursorItems(cmd.Context(), peopleAllFlag, peopleCursorFlag, peopleLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.Person], error) {
			resp, err := client.ListPeople(ctx, api.PeopleListParams{
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.Person]{}, err
			}
			return CursorListResult[api.Person]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing people")
		}

		total := page.Total
		if peopleAllFlag {
			total = len(people)
		}

		response := api.PeopleListResponse{
			Data: people,
		}
		response.Page.Next = ""
		response.Page.Total = total

		return outputList(cmd, f, people, hasMore, "No people found.", []string{"ID", "NAME", "EMAIL", "JOB TITLE", "STATUS"}, func(p api.Person) []string {
			return []string{p.HRISProfileID, p.Name, p.Email, p.JobTitle, p.Status}
		}, response)
	},
}

var peoplePersonalFlag bool

var peopleGetCmd = &cobra.Command{
	Use:   "get <hris-profile-id>",
	Short: "Get person details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "getting person")
		}

		// If --personal flag is set, use the /personal endpoint
		if peoplePersonalFlag {
			rawData, err := client.GetPersonPersonal(cmd.Context(), args[0])
			if err != nil {
				return HandleError(f, err, "getting personal info")
			}

			// Parse into a map to access any field
			var data map[string]any
			if err := json.Unmarshal(rawData, &data); err != nil {
				return HandleError(f, err, "parsing personal info")
			}

			return f.OutputFiltered(cmd.Context(), func() {
				if id, ok := data["id"]; ok {
					f.PrintText(fmt.Sprintf("ID:         %v", id))
				}
				if workerID, ok := data["worker_id"]; ok {
					f.PrintText(fmt.Sprintf("Worker ID:  %v", workerID))
				}
				firstName, _ := data["first_name"].(string)
				lastName, _ := data["last_name"].(string)
				f.PrintText("Name:       " + strings.TrimSpace(firstName+" "+lastName))
				if email, ok := data["email"].(string); ok {
					f.PrintText("Email:      " + email)
				}
			}, data)
		}

		person, err := client.GetPerson(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "getting person")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText("Name:       " + person.Name)
			f.PrintText("Email:      " + person.Email)
			f.PrintText("Job Title:  " + person.JobTitle)
			f.PrintText("Department: " + person.Department())
			f.PrintText("Status:     " + person.Status)
			f.PrintText("Country:    " + person.Country)
			f.PrintText("Start Date: " + person.StartDate)
		}, person)
	},
}

var peopleSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for a person",
	Long: `Search for a person by email or name.

Use --email for exact email lookup (faster, uses API).
Use --name for name search (searches first/last name, case-insensitive).

Examples:
  deel people search --email catherine@example.com
  deel people search --name "Catherine Song"
  deel people search --name song`,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if peopleEmailFlag == "" && peopleNameFlag == "" {
			f.PrintError("--email or --name is required")
			return fmt.Errorf("--email or --name flag is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "searching people")
		}

		// Email search - use API directly
		if peopleEmailFlag != "" {
			person, err := client.SearchPeopleByEmail(cmd.Context(), peopleEmailFlag)
			if err != nil {
				return HandleError(f, err, "searching people")
			}

			return f.OutputFiltered(cmd.Context(), func() {
				f.PrintText("Found: " + person.Name)
				f.PrintText("ID:    " + person.HRISProfileID)
				f.PrintText("Email: " + person.Email)
			}, person)
		}

		// Name search - fetch all people and filter client-side
		searchName := strings.ToLower(peopleNameFlag)
		var matches []api.Person
		cursor := ""

		for {
			resp, err := client.ListPeople(cmd.Context(), api.PeopleListParams{
				Limit:  100,
				Cursor: cursor,
			})
			if err != nil {
				return HandleError(f, err, "searching people")
			}

			for _, p := range resp.Data {
				fullName := strings.ToLower(p.FirstName + " " + p.LastName)
				firstName := strings.ToLower(p.FirstName)
				lastName := strings.ToLower(p.LastName)

				if strings.Contains(fullName, searchName) ||
					strings.Contains(firstName, searchName) ||
					strings.Contains(lastName, searchName) {
					matches = append(matches, p)
				}
			}

			if resp.Page.Next == "" {
				break
			}
			cursor = resp.Page.Next
		}

		if len(matches) == 0 {
			f.PrintText("No people found matching: " + peopleNameFlag)
			return nil
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintText(fmt.Sprintf("Found %d match(es) for \"%s\":\n", len(matches), peopleNameFlag))
			table := f.NewTable("ID", "NAME", "EMAIL", "JOB TITLE", "STATUS")
			for _, p := range matches {
				table.AddRow(p.HRISProfileID, p.Name, p.Email, p.JobTitle, p.Status)
			}
			table.Render()
		}, matches)
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
			return HandleError(f, err, "initializing client")
		}

		fields, err := client.ListCustomFields(cmd.Context())
		if err != nil {
			return HandleError(f, err, "list custom fields")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
			return HandleError(f, err, "initializing client")
		}

		field, err := client.GetCustomField(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "get custom field")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Person",
			Description: "Create a new person",
			Details: map[string]string{
				"Email":     peopleCreateEmailFlag,
				"FirstName": peopleCreateFirstNameFlag,
				"LastName":  peopleCreateLastNameFlag,
				"Type":      peopleCreateTypeFlag,
				"Country":   peopleCreateCountryFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
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
			return HandleError(f, err, "create person")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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

		details := map[string]string{
			"ID": args[0],
		}
		if peopleUpdateFirstNameFlag != "" {
			details["FirstName"] = peopleUpdateFirstNameFlag
		}
		if peopleUpdateLastNameFlag != "" {
			details["LastName"] = peopleUpdateLastNameFlag
		}
		if peopleUpdatePhoneFlag != "" {
			details["Phone"] = peopleUpdatePhoneFlag
		}
		if peopleUpdateNationalityFlag != "" {
			details["Nationality"] = peopleUpdateNationalityFlag
		}
		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Person",
			Description: "Update personal info",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		info := api.PersonalInfo{
			FirstName:   peopleUpdateFirstNameFlag,
			LastName:    peopleUpdateLastNameFlag,
			Phone:       peopleUpdatePhoneFlag,
			Nationality: peopleUpdateNationalityFlag,
		}

		updated, err := client.UpdatePersonalInfo(cmd.Context(), args[0], info)
		if err != nil {
			return HandleError(f, err, "update personal info")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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

// Flags for set-department command
var setDepartmentIDFlag string

var setDepartmentCmd = &cobra.Command{
	Use:   "set-department <id>",
	Short: "Set person's department",
	Long:  "Set the department for a person. Requires --department-id. Use 'deel org departments list' to see available departments.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if setDepartmentIDFlag == "" {
			f.PrintError("--department-id flag is required")
			return fmt.Errorf("--department-id flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "PersonDepartment",
			Description: "Set person's department",
			Details: map[string]string{
				"PersonID":     args[0],
				"DepartmentID": setDepartmentIDFlag,
			},
		}); ok || err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		dept, err := client.UpdatePersonDepartment(cmd.Context(), args[0], api.UpdatePersonDepartmentParams{
			DepartmentID: setDepartmentIDFlag,
		})
		if err != nil {
			return HandleError(f, err, "set department")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Department updated successfully")
			if dept != nil {
				f.PrintText("Department ID:   " + dept.ID)
				f.PrintText("Department Name: " + dept.Name)
			}
		}, dept)
	},
}

// Flags for working location update command
var (
	peopleLocationCountryFlag    string
	peopleLocationStateFlag      string
	peopleLocationCityFlag       string
	peopleLocationAddressFlag    string
	peopleLocationPostalCodeFlag string
	peopleLocationTimezoneFlag   string
)

var peopleLocationCmd = &cobra.Command{
	Use:   "working-location <id>",
	Short: "Update working location",
	Long:  "Update the working location for a person. Requires --country.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if peopleLocationCountryFlag == "" {
			f.PrintError("--country flag is required")
			return fmt.Errorf("--country flag is required")
		}

		details := map[string]string{
			"ID":      args[0],
			"Country": peopleLocationCountryFlag,
		}
		if peopleLocationStateFlag != "" {
			details["State"] = peopleLocationStateFlag
		}
		if peopleLocationCityFlag != "" {
			details["City"] = peopleLocationCityFlag
		}
		if peopleLocationAddressFlag != "" {
			details["Address"] = peopleLocationAddressFlag
		}
		if peopleLocationPostalCodeFlag != "" {
			details["PostalCode"] = peopleLocationPostalCodeFlag
		}
		if peopleLocationTimezoneFlag != "" {
			details["Timezone"] = peopleLocationTimezoneFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "WorkingLocation",
			Description: "Update working location",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		location := api.WorkingLocation{
			Country:    peopleLocationCountryFlag,
			State:      peopleLocationStateFlag,
			City:       peopleLocationCityFlag,
			Address:    peopleLocationAddressFlag,
			PostalCode: peopleLocationPostalCodeFlag,
			Timezone:   peopleLocationTimezoneFlag,
		}

		updated, err := client.UpdateWorkingLocation(cmd.Context(), args[0], location)
		if err != nil {
			return HandleError(f, err, "update working location")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Working location updated successfully")
			f.PrintText("ID:          " + updated.ID)
			f.PrintText("Country:     " + updated.Country)
			if updated.State != "" {
				f.PrintText("State:       " + updated.State)
			}
			if updated.City != "" {
				f.PrintText("City:        " + updated.City)
			}
			if updated.Address != "" {
				f.PrintText("Address:     " + updated.Address)
			}
			if updated.PostalCode != "" {
				f.PrintText("Postal Code: " + updated.PostalCode)
			}
			if updated.Timezone != "" {
				f.PrintText("Timezone:    " + updated.Timezone)
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
			return HandleError(f, err, "initializing client")
		}

		params := api.ListAdjustmentsParams{
			ContractID: adjustmentsListContractIDFlag,
			CategoryID: adjustmentsListCategoryIDFlag,
		}

		adjustments, err := client.ListAdjustments(cmd.Context(), params)
		if err != nil {
			return HandleError(f, err, "list adjustments")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
	adjustmentsCreateContractIDFlag     string
	adjustmentsCreateCategoryIDFlag     string
	adjustmentsCreateAmountFlag         string
	adjustmentsCreateCurrencyFlag       string
	adjustmentsCreateDescriptionFlag    string
	adjustmentsCreateDateFlag           string
	adjustmentsCreateCycleReferenceFlag string
	adjustmentsCreateMoveNextCycleFlag  bool
	adjustmentsCreateVendorFlag         string
	adjustmentsCreateCountryFlag        string
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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Adjustment",
			Description: "Create adjustment",
			Details: map[string]string{
				"ContractID":  adjustmentsCreateContractIDFlag,
				"CategoryID":  adjustmentsCreateCategoryIDFlag,
				"Amount":      dryrun.FormatAmount(amount, adjustmentsCreateCurrencyFlag),
				"Description": adjustmentsCreateDescriptionFlag,
				"Date":        adjustmentsCreateDateFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.CreateAdjustmentParams{
			ContractID:     adjustmentsCreateContractIDFlag,
			CategoryID:     adjustmentsCreateCategoryIDFlag,
			Amount:         amount,
			Currency:       adjustmentsCreateCurrencyFlag,
			Description:    adjustmentsCreateDescriptionFlag,
			Date:           adjustmentsCreateDateFlag,
			CycleReference: adjustmentsCreateCycleReferenceFlag,
			MoveNextCycle:  adjustmentsCreateMoveNextCycleFlag,
			Vendor:         adjustmentsCreateVendorFlag,
			Country:        adjustmentsCreateCountryFlag,
		}

		adjustment, err := client.CreateAdjustment(cmd.Context(), params)
		if err != nil {
			return HandleError(f, err, "create adjustment")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Adjustment created successfully")
			f.PrintText("ID:          " + adjustment.ID)
			f.PrintText("Contract ID: " + adjustment.ContractID)
			f.PrintText("Category ID: " + adjustment.CategoryID)
			f.PrintText(fmt.Sprintf("Amount:      %.2f %s", adjustment.Amount, adjustment.Currency))
			f.PrintText("Description: " + adjustment.Description)
			f.PrintText("Date:        " + adjustment.Date)
			f.PrintText("Status:      " + adjustment.Status)
			if adjustment.CycleReference != "" {
				f.PrintText("Cycle Ref:   " + adjustment.CycleReference)
			}
			if adjustment.ActualStartCycleDate != "" {
				f.PrintText("Cycle Start: " + adjustment.ActualStartCycleDate)
			}
			if adjustment.ActualEndCycleDate != "" {
				f.PrintText("Cycle End:   " + adjustment.ActualEndCycleDate)
			}
			f.PrintText("Created:     " + adjustment.CreatedAt)
		}, adjustment)
	},
}

var adjustmentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get adjustment",
	Long:  "Get an adjustment by ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		adjustment, err := client.GetAdjustment(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "get adjustment")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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

// Flags for adjustments update command
var (
	adjustmentsUpdateAmountFlag      string
	adjustmentsUpdateDescriptionFlag string
	adjustmentsUpdateDateFlag        string
)

var adjustmentsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update adjustment",
	Long:  "Update an adjustment. Optional flags: --amount, --description, --date.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		var amount float64
		if adjustmentsUpdateAmountFlag != "" {
			parsed, err := strconv.ParseFloat(adjustmentsUpdateAmountFlag, 64)
			if err != nil {
				f.PrintError("Invalid --amount value: %v", err)
				return fmt.Errorf("invalid --amount value: %w", err)
			}
			amount = parsed
		}

		details := map[string]string{
			"ID": args[0],
		}
		if adjustmentsUpdateAmountFlag != "" {
			details["Amount"] = adjustmentsUpdateAmountFlag
		}
		if adjustmentsUpdateDescriptionFlag != "" {
			details["Description"] = adjustmentsUpdateDescriptionFlag
		}
		if adjustmentsUpdateDateFlag != "" {
			details["Date"] = adjustmentsUpdateDateFlag
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "UPDATE",
			Resource:    "Adjustment",
			Description: "Update adjustment",
			Details:     details,
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.UpdateAdjustmentParams{
			Amount:      amount,
			Description: adjustmentsUpdateDescriptionFlag,
			Date:        adjustmentsUpdateDateFlag,
		}

		adjustment, err := client.UpdateAdjustment(cmd.Context(), args[0], params)
		if err != nil {
			return HandleError(f, err, "update adjustment")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Adjustment updated successfully")
			f.PrintText("ID:          " + adjustment.ID)
			f.PrintText("Contract ID: " + adjustment.ContractID)
			f.PrintText("Category ID: " + adjustment.CategoryID)
			f.PrintText(fmt.Sprintf("Amount:      %.2f %s", adjustment.Amount, adjustment.Currency))
			f.PrintText("Description: " + adjustment.Description)
			f.PrintText("Date:        " + adjustment.Date)
			f.PrintText("Status:      " + adjustment.Status)
		}, adjustment)
	},
}

var adjustmentsCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List adjustment categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		categories, err := client.ListAdjustmentCategories(cmd.Context())
		if err != nil {
			return HandleError(f, err, "list adjustment categories")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			if len(categories) == 0 {
				f.PrintText("No adjustment categories found")
				return
			}
			table := f.NewTable("ID", "NAME", "TYPE", "DESCRIPTION")
			for _, c := range categories {
				table.AddRow(c.ID, c.Name, c.Type, c.Description)
			}
			table.Render()
		}, categories)
	},
}

var adjustmentsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete adjustment",
	Long:  "Delete an adjustment by ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "DELETE",
			Resource:    "Adjustment",
			Description: "Delete adjustment",
			Details: map[string]string{
				"ID": args[0],
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		err = client.DeleteAdjustment(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "delete adjustment")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
			return HandleError(f, err, "initializing client")
		}

		managers, err := client.ListManagers(cmd.Context())
		if err != nil {
			return HandleError(f, err, "list managers")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "Manager",
			Description: "Create manager",
			Details: map[string]string{
				"Email":     managersCreateEmailFlag,
				"FirstName": managersCreateFirstNameFlag,
				"LastName":  managersCreateLastNameFlag,
				"Role":      managersCreateRoleFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.CreateManagerParams{
			Email:     managersCreateEmailFlag,
			FirstName: managersCreateFirstNameFlag,
			LastName:  managersCreateLastNameFlag,
			Role:      managersCreateRoleFlag,
		}

		manager, err := client.CreateManager(cmd.Context(), params)
		if err != nil {
			return HandleError(f, err, "create manager")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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

// Flags for assign-manager command
var (
	assignManagerEmailFlag     string
	assignManagerManagerIDFlag string
	assignManagerStartDateFlag string
)

var assignManagerCmd = &cobra.Command{
	Use:   "assign-manager",
	Short: "Assign a manager to a worker by email",
	Long: `Assign a manager to a worker using their email address.
This is a convenience command that looks up the worker's profile ID by email
and creates a worker relation using the worker-relations API.

NOTE: The worker must have signed their contract and appear in the People directory
before you can assign a manager. This typically happens after:
  1. Contract is created and signed by client
  2. Worker is invited
  3. Worker signs the contract

Examples:
  deel people assign-manager --email worker@example.com --manager fd470477-d950-47dd-93eb-d31830d6caca
  deel people assign-manager --email worker@example.com --manager fd470477-... --start-date 2026-01-20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if assignManagerEmailFlag == "" {
			f.PrintError("--email flag is required")
			return fmt.Errorf("--email flag is required")
		}
		if assignManagerManagerIDFlag == "" {
			f.PrintError("--manager flag is required")
			return fmt.Errorf("--manager flag is required")
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		// Look up the worker by email
		person, err := client.SearchPeopleByEmail(cmd.Context(), assignManagerEmailFlag)
		if err != nil {
			return HandleError(f, err, "searching worker by email")
		}

		if person.HRISProfileID == "" {
			f.PrintError("Worker found but has no profile ID")
			return fmt.Errorf("worker has no profile ID")
		}

		// Use start date from flag or person's start date
		startDate := assignManagerStartDateFlag
		if startDate == "" {
			startDate = person.StartDate
		}
		// StartDate is required, so if still empty, error out
		if startDate == "" {
			f.PrintError("--start-date is required (worker has no start date on file)")
			return fmt.Errorf("--start-date is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "ASSIGN",
			Resource:    "WorkerRelation",
			Description: "Assign manager to worker",
			Details: map[string]string{
				"WorkerEmail": assignManagerEmailFlag,
				"WorkerName":  person.Name,
				"ProfileID":   person.HRISProfileID,
				"ManagerID":   assignManagerManagerIDFlag,
				"StartDate":   startDate,
			},
		}); ok {
			return err
		}

		// Use the HRIS parent relation endpoint (recommended by Deel support)
		params := api.SetWorkerManagerParams{
			ManagerID: assignManagerManagerIDFlag,
			StartDate: startDate,
		}

		relation, err := client.SetWorkerManager(cmd.Context(), person.HRISProfileID, params)
		if err != nil {
			return HandleError(f, err, "assigning manager")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Manager assigned successfully")
			f.PrintText("Worker:      " + person.Name + " (" + assignManagerEmailFlag + ")")
			f.PrintText("Profile ID:  " + person.HRISProfileID)
			f.PrintText("Manager ID:  " + assignManagerManagerIDFlag)
			if relation != nil {
				if relation.StartDate != "" {
					f.PrintText("Start Date:  " + relation.StartDate)
				}
				if relation.Status != "" {
					f.PrintText("Status:      " + relation.Status)
				}
			}
		}, relation)
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
			return HandleError(f, err, "initializing client")
		}

		relations, err := client.ListWorkerRelations(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "list worker relations")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "WorkerRelation",
			Description: "Create worker relation",
			Details: map[string]string{
				"ProfileID":    relationsCreateProfileIDFlag,
				"ManagerID":    relationsCreateManagerIDFlag,
				"RelationType": relationsCreateRelationTypeFlag,
				"StartDate":    relationsCreateStartDateFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		params := api.CreateWorkerRelationParams{
			ProfileID:    relationsCreateProfileIDFlag,
			ManagerID:    relationsCreateManagerIDFlag,
			RelationType: relationsCreateRelationTypeFlag,
			StartDate:    relationsCreateStartDateFlag,
		}

		relation, err := client.CreateWorkerRelation(cmd.Context(), params)
		if err != nil {
			return HandleError(f, err, "create worker relation")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
			return HandleError(f, err, "initializing client")
		}

		err = client.DeleteWorkerRelation(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "delete worker relation")
		}

		return f.OutputFiltered(cmd.Context(), func() {
			f.PrintSuccess("Worker relation deleted successfully")
			f.PrintText("ID: " + args[0])
		}, map[string]string{"id": args[0], "status": "deleted"})
	},
}

func init() {
	peopleListCmd.Flags().IntVar(&peopleLimitFlag, "limit", 100, "Maximum results")
	peopleListCmd.Flags().StringVar(&peopleCursorFlag, "cursor", "", "Pagination cursor")
	peopleListCmd.Flags().BoolVar(&peopleAllFlag, "all", false, "Fetch all pages")

	peopleSearchCmd.Flags().StringVar(&peopleEmailFlag, "email", "", "Email to search for (exact match)")
	peopleSearchCmd.Flags().StringVar(&peopleNameFlag, "name", "", "Name to search for (partial match, case-insensitive)")

	// People get command flags
	peopleGetCmd.Flags().BoolVar(&peoplePersonalFlag, "personal", false, "Get personal info including numeric worker_id")

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

	// Set-department command flags
	setDepartmentCmd.Flags().StringVar(&setDepartmentIDFlag, "department-id", "", "Department ID (required)")

	// Working location command flags
	peopleLocationCmd.Flags().StringVar(&peopleLocationCountryFlag, "country", "", "Country code (required)")
	peopleLocationCmd.Flags().StringVar(&peopleLocationStateFlag, "state", "", "State (optional)")
	peopleLocationCmd.Flags().StringVar(&peopleLocationCityFlag, "city", "", "City (optional)")
	peopleLocationCmd.Flags().StringVar(&peopleLocationAddressFlag, "address", "", "Address (optional)")
	peopleLocationCmd.Flags().StringVar(&peopleLocationPostalCodeFlag, "postal-code", "", "Postal code (optional)")
	peopleLocationCmd.Flags().StringVar(&peopleLocationTimezoneFlag, "timezone", "", "Timezone (optional)")

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
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateCycleReferenceFlag, "cycle-reference", "", "Payroll cycle reference (optional)")
	adjustmentsCreateCmd.Flags().BoolVar(&adjustmentsCreateMoveNextCycleFlag, "move-next-cycle", false, "Move adjustment to next payroll cycle (optional)")
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateVendorFlag, "vendor", "", "Vendor name (optional)")
	adjustmentsCreateCmd.Flags().StringVar(&adjustmentsCreateCountryFlag, "country", "", "Country code ISO 3166-1 alpha-2 (optional, defaults to CA)")

	// Adjustments update command flags
	adjustmentsUpdateCmd.Flags().StringVar(&adjustmentsUpdateAmountFlag, "amount", "", "Amount (optional)")
	adjustmentsUpdateCmd.Flags().StringVar(&adjustmentsUpdateDescriptionFlag, "description", "", "Description (optional)")
	adjustmentsUpdateCmd.Flags().StringVar(&adjustmentsUpdateDateFlag, "date", "", "Date YYYY-MM-DD (optional)")

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

	// Assign-manager command flags
	assignManagerCmd.Flags().StringVar(&assignManagerEmailFlag, "email", "", "Worker email address (required)")
	assignManagerCmd.Flags().StringVar(&assignManagerManagerIDFlag, "manager", "", "Manager ID (required)")
	assignManagerCmd.Flags().StringVar(&assignManagerStartDateFlag, "start-date", "", "Start date YYYY-MM-DD (default: worker's start date)")

	// Add subcommands to adjustments
	adjustmentsCmd.AddCommand(adjustmentsListCmd)
	adjustmentsCmd.AddCommand(adjustmentsGetCmd)
	adjustmentsCmd.AddCommand(adjustmentsCreateCmd)
	adjustmentsCmd.AddCommand(adjustmentsUpdateCmd)
	adjustmentsCmd.AddCommand(adjustmentsDeleteCmd)
	adjustmentsCmd.AddCommand(adjustmentsCategoriesCmd)

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
	peopleCmd.AddCommand(peopleLocationCmd)
	peopleCmd.AddCommand(setDepartmentCmd)
	peopleCmd.AddCommand(customFieldsCmd)
	peopleCmd.AddCommand(adjustmentsCmd)
	peopleCmd.AddCommand(managersCmd)
	peopleCmd.AddCommand(relationsCmd)
	peopleCmd.AddCommand(assignManagerCmd)
}
