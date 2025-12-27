package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

var atsCmd = &cobra.Command{
	Use:   "ats",
	Short: "Applicant tracking system",
	Long:  "View ATS offers and candidates.",
}

var (
	atsStatusFlag      string
	atsLimitFlag       int
	atsCursorFlag      string
	atsDepartmentIDFlag string
	atsLocationIDFlag   string
	atsJobIDFlag        string
	atsCandidateIDFlag  string
	atsStageFlag        string
	atsSearchFlag       string
	atsRemoteFlag       bool
	// Job creation flags
	atsJobTitleFlag          string
	atsJobDepartmentIDFlag   string
	atsJobLocationIDFlag     string
	atsJobEmploymentTypeFlag string
	atsJobDescriptionFlag    string
)

var atsOffersCmd = &cobra.Command{
	Use:   "offers",
	Short: "List ATS offers",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		offers, err := client.ListATSOffers(cmd.Context(), api.ATSOffersListParams{
			Status: atsStatusFlag,
			Limit:  atsLimitFlag,
			Cursor: atsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list offers: %v", err)
			return err
		}

		return f.Output(func() {
			if len(offers) == 0 {
				f.PrintText("No offers found.")
				return
			}
			table := f.NewTable("ID", "CANDIDATE", "POSITION", "SALARY", "STATUS")
			for _, o := range offers {
				salary := fmt.Sprintf("%.2f %s", o.Salary, o.Currency)
				table.AddRow(o.ID, o.Candidate, o.Position, salary, o.Status)
			}
			table.Render()
		}, offers)
	},
}

// Jobs commands
var atsJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage ATS jobs",
}

var atsJobsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ATS jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		jobs, err := client.ListATSJobs(cmd.Context(), api.ATSJobsListParams{
			Status:       atsStatusFlag,
			DepartmentID: atsDepartmentIDFlag,
			LocationID:   atsLocationIDFlag,
			Limit:        atsLimitFlag,
			Cursor:       atsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list jobs: %v", err)
			return err
		}

		return f.Output(func() {
			if len(jobs) == 0 {
				f.PrintText("No jobs found.")
				return
			}
			table := f.NewTable("ID", "TITLE", "DEPARTMENT", "LOCATION", "TYPE", "STATUS")
			for _, j := range jobs {
				table.AddRow(j.ID, j.Title, j.Department, j.Location, j.EmploymentType, j.Status)
			}
			table.Render()
		}, jobs)
	},
}

var atsJobsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new ATS job",
	Long:  "Create a new ATS job. Requires --title, --department-id, --location-id, and --employment-type flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if atsJobTitleFlag == "" {
			f.PrintError("--title flag is required")
			return fmt.Errorf("--title flag is required")
		}
		if atsJobDepartmentIDFlag == "" {
			f.PrintError("--department-id flag is required")
			return fmt.Errorf("--department-id flag is required")
		}
		if atsJobLocationIDFlag == "" {
			f.PrintError("--location-id flag is required")
			return fmt.Errorf("--location-id flag is required")
		}
		if atsJobEmploymentTypeFlag == "" {
			f.PrintError("--employment-type flag is required")
			return fmt.Errorf("--employment-type flag is required")
		}

		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		job, err := client.CreateATSJob(cmd.Context(), api.CreateATSJobParams{
			Title:          atsJobTitleFlag,
			DepartmentID:   atsJobDepartmentIDFlag,
			LocationID:     atsJobLocationIDFlag,
			EmploymentType: atsJobEmploymentTypeFlag,
			Description:    atsJobDescriptionFlag,
		})
		if err != nil {
			f.PrintError("Failed to create job: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintSuccess("Job created successfully")
			f.PrintText("ID:          " + job.ID)
			f.PrintText("Title:       " + job.Title)
			f.PrintText("Department:  " + job.Department)
			f.PrintText("Location:    " + job.Location)
			f.PrintText("Type:        " + job.EmploymentType)
			f.PrintText("Status:      " + job.Status)
		}, job)
	},
}

// Job Postings commands
var atsPostingsCmd = &cobra.Command{
	Use:   "postings",
	Short: "Manage job postings",
}

var atsPostingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List job postings",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		postings, err := client.ListATSJobPostings(cmd.Context(), api.ATSJobPostingsListParams{
			Status: atsStatusFlag,
			JobID:  atsJobIDFlag,
			Limit:  atsLimitFlag,
			Cursor: atsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list job postings: %v", err)
			return err
		}

		return f.Output(func() {
			if len(postings) == 0 {
				f.PrintText("No job postings found.")
				return
			}
			table := f.NewTable("ID", "TITLE", "DEPARTMENT", "LOCATION", "STATUS", "POSTED AT")
			for _, p := range postings {
				table.AddRow(p.ID, p.Title, p.Department, p.Location, p.Status, p.PostedAt)
			}
			table.Render()
		}, postings)
	},
}

var atsPostingsGetCmd = &cobra.Command{
	Use:   "get <posting-id>",
	Short: "Get job posting details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		posting, err := client.GetATSJobPosting(cmd.Context(), args[0])
		if err != nil {
			f.PrintError("Failed to get job posting: %v", err)
			return err
		}

		return f.Output(func() {
			f.PrintText("ID:          " + posting.ID)
			f.PrintText("Job ID:      " + posting.JobID)
			f.PrintText("Title:       " + posting.Title)
			f.PrintText("Department:  " + posting.Department)
			f.PrintText("Location:    " + posting.Location)
			f.PrintText("Type:        " + posting.EmploymentType)
			f.PrintText("Status:      " + posting.Status)
			f.PrintText("Posted At:   " + posting.PostedAt)
			if posting.ClosedAt != "" {
				f.PrintText("Closed At:   " + posting.ClosedAt)
			}
			f.PrintText("URL:         " + posting.URL)
			if posting.Description != "" {
				f.PrintText("\nDescription:")
				f.PrintText(posting.Description)
			}
		}, posting)
	},
}

// Applications command
var atsApplicationsCmd = &cobra.Command{
	Use:   "applications",
	Short: "Manage applications",
}

var atsApplicationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		apps, err := client.ListATSApplications(cmd.Context(), api.ATSApplicationsListParams{
			Status:      atsStatusFlag,
			JobID:       atsJobIDFlag,
			CandidateID: atsCandidateIDFlag,
			Stage:       atsStageFlag,
			Limit:       atsLimitFlag,
			Cursor:      atsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list applications: %v", err)
			return err
		}

		return f.Output(func() {
			if len(apps) == 0 {
				f.PrintText("No applications found.")
				return
			}
			table := f.NewTable("ID", "CANDIDATE", "JOB", "STATUS", "STAGE", "APPLIED AT")
			for _, a := range apps {
				table.AddRow(a.ID, a.CandidateName, a.JobTitle, a.Status, a.Stage, a.AppliedAt)
			}
			table.Render()
		}, apps)
	},
}

// Candidates command
var atsCandidatesCmd = &cobra.Command{
	Use:   "candidates",
	Short: "Manage candidates",
}

var atsCandidatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List candidates",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		candidates, err := client.ListATSCandidates(cmd.Context(), api.ATSCandidatesListParams{
			Search: atsSearchFlag,
			Limit:  atsLimitFlag,
			Cursor: atsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list candidates: %v", err)
			return err
		}

		return f.Output(func() {
			if len(candidates) == 0 {
				f.PrintText("No candidates found.")
				return
			}
			table := f.NewTable("ID", "NAME", "EMAIL", "PHONE", "LOCATION")
			for _, c := range candidates {
				name := fmt.Sprintf("%s %s", c.FirstName, c.LastName)
				table.AddRow(c.ID, name, c.Email, c.Phone, c.Location)
			}
			table.Render()
		}, candidates)
	},
}

// Departments command
var atsDepartmentsCmd = &cobra.Command{
	Use:   "departments",
	Short: "Manage departments",
}

var atsDepartmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List departments",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		departments, err := client.ListATSDepartments(cmd.Context(), api.ATSDepartmentsListParams{
			Limit:  atsLimitFlag,
			Cursor: atsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list departments: %v", err)
			return err
		}

		return f.Output(func() {
			if len(departments) == 0 {
				f.PrintText("No departments found.")
				return
			}
			table := f.NewTable("ID", "NAME", "PARENT ID", "CREATED AT")
			for _, d := range departments {
				table.AddRow(d.ID, d.Name, d.ParentID, d.CreatedAt)
			}
			table.Render()
		}, departments)
	},
}

// Locations command
var atsLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage locations",
}

var atsLocationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		var remotePtr *bool
		if cmd.Flags().Changed("remote") {
			remotePtr = &atsRemoteFlag
		}

		locations, err := client.ListATSLocations(cmd.Context(), api.ATSLocationsListParams{
			Remote: remotePtr,
			Limit:  atsLimitFlag,
			Cursor: atsCursorFlag,
		})
		if err != nil {
			f.PrintError("Failed to list locations: %v", err)
			return err
		}

		return f.Output(func() {
			if len(locations) == 0 {
				f.PrintText("No locations found.")
				return
			}
			table := f.NewTable("ID", "NAME", "CITY", "COUNTRY", "REMOTE")
			for _, l := range locations {
				remote := "No"
				if l.Remote {
					remote = "Yes"
				}
				table.AddRow(l.ID, l.Name, l.City, l.Country, remote)
			}
			table.Render()
		}, locations)
	},
}

// Rejection Reasons command
var atsRejectionReasonsCmd = &cobra.Command{
	Use:   "rejection-reasons",
	Short: "Manage rejection reasons",
}

var atsRejectionReasonsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rejection reasons",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()
		client, err := getClient()
		if err != nil {
			f.PrintError("Failed to get client: %v", err)
			return err
		}

		reasons, err := client.ListRejectionReasons(cmd.Context())
		if err != nil {
			f.PrintError("Failed to list rejection reasons: %v", err)
			return err
		}

		return f.Output(func() {
			if len(reasons) == 0 {
				f.PrintText("No rejection reasons found.")
				return
			}
			table := f.NewTable("ID", "REASON", "DESCRIPTION")
			for _, r := range reasons {
				desc := r.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				table.AddRow(r.ID, r.Reason, desc)
			}
			table.Render()
		}, reasons)
	},
}

func init() {
	// Offers command flags
	atsOffersCmd.Flags().StringVar(&atsStatusFlag, "status", "", "Filter by status")
	atsOffersCmd.Flags().IntVar(&atsLimitFlag, "limit", 50, "Maximum results")
	atsOffersCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")

	// Jobs list command flags
	atsJobsListCmd.Flags().StringVar(&atsStatusFlag, "status", "", "Filter by status")
	atsJobsListCmd.Flags().StringVar(&atsDepartmentIDFlag, "department-id", "", "Filter by department ID")
	atsJobsListCmd.Flags().StringVar(&atsLocationIDFlag, "location-id", "", "Filter by location ID")
	atsJobsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 50, "Maximum results")
	atsJobsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")

	// Jobs create command flags
	atsJobsCreateCmd.Flags().StringVar(&atsJobTitleFlag, "title", "", "Job title (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobDepartmentIDFlag, "department-id", "", "Department ID (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobLocationIDFlag, "location-id", "", "Location ID (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobEmploymentTypeFlag, "employment-type", "", "Employment type (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobDescriptionFlag, "description", "", "Job description (optional)")

	// Job postings list command flags
	atsPostingsListCmd.Flags().StringVar(&atsStatusFlag, "status", "", "Filter by status")
	atsPostingsListCmd.Flags().StringVar(&atsJobIDFlag, "job-id", "", "Filter by job ID")
	atsPostingsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 50, "Maximum results")
	atsPostingsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")

	// Applications list command flags
	atsApplicationsListCmd.Flags().StringVar(&atsStatusFlag, "status", "", "Filter by status")
	atsApplicationsListCmd.Flags().StringVar(&atsJobIDFlag, "job-id", "", "Filter by job ID")
	atsApplicationsListCmd.Flags().StringVar(&atsCandidateIDFlag, "candidate-id", "", "Filter by candidate ID")
	atsApplicationsListCmd.Flags().StringVar(&atsStageFlag, "stage", "", "Filter by stage")
	atsApplicationsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 50, "Maximum results")
	atsApplicationsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")

	// Candidates list command flags
	atsCandidatesListCmd.Flags().StringVar(&atsSearchFlag, "search", "", "Search candidates by name or email")
	atsCandidatesListCmd.Flags().IntVar(&atsLimitFlag, "limit", 50, "Maximum results")
	atsCandidatesListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")

	// Departments list command flags
	atsDepartmentsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 50, "Maximum results")
	atsDepartmentsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")

	// Locations list command flags
	atsLocationsListCmd.Flags().BoolVar(&atsRemoteFlag, "remote", false, "Filter by remote locations")
	atsLocationsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 50, "Maximum results")
	atsLocationsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")

	// Add subcommands
	atsJobsCmd.AddCommand(atsJobsListCmd)
	atsJobsCmd.AddCommand(atsJobsCreateCmd)

	atsPostingsCmd.AddCommand(atsPostingsListCmd)
	atsPostingsCmd.AddCommand(atsPostingsGetCmd)

	atsApplicationsCmd.AddCommand(atsApplicationsListCmd)

	atsCandidatesCmd.AddCommand(atsCandidatesListCmd)

	atsDepartmentsCmd.AddCommand(atsDepartmentsListCmd)

	atsLocationsCmd.AddCommand(atsLocationsListCmd)

	atsRejectionReasonsCmd.AddCommand(atsRejectionReasonsListCmd)

	// Add all commands to ats root command
	atsCmd.AddCommand(atsOffersCmd)
	atsCmd.AddCommand(atsJobsCmd)
	atsCmd.AddCommand(atsPostingsCmd)
	atsCmd.AddCommand(atsApplicationsCmd)
	atsCmd.AddCommand(atsCandidatesCmd)
	atsCmd.AddCommand(atsDepartmentsCmd)
	atsCmd.AddCommand(atsLocationsCmd)
	atsCmd.AddCommand(atsRejectionReasonsCmd)
}
