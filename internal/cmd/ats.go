package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/deel-cli/internal/api"
	"github.com/salmonumbrella/deel-cli/internal/dryrun"
)

var atsCmd = &cobra.Command{
	Use:   "ats",
	Short: "Applicant tracking system",
	Long:  "View ATS offers and candidates.",
}

var (
	atsStatusFlag       string
	atsLimitFlag        int
	atsCursorFlag       string
	atsAllFlag          bool
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
		client, f, err := initClient("listing ats offers")
		if err != nil {
			return err
		}

		offers, page, hasMore, err := collectCursorItems(cmd.Context(), atsAllFlag, atsCursorFlag, atsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ATSOffer], error) {
			resp, err := client.ListATSOffers(ctx, api.ATSOffersListParams{
				Status: atsStatusFlag,
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.ATSOffer]{}, err
			}
			return CursorListResult[api.ATSOffer]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing ats offers")
		}

		response := makeListResponse(offers, page)

		return outputList(cmd, f, offers, hasMore, "No offers found.", []string{"ID", "CANDIDATE", "POSITION", "SALARY", "STATUS"}, func(o api.ATSOffer) []string {
			salary := fmt.Sprintf("%.2f %s", o.Salary, o.Currency)
			return []string{o.ID, o.Candidate, o.Position, salary, o.Status}
		}, response)
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
		client, f, err := initClient("listing ats jobs")
		if err != nil {
			return err
		}

		jobs, page, hasMore, err := collectCursorItems(cmd.Context(), atsAllFlag, atsCursorFlag, atsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ATSJob], error) {
			resp, err := client.ListATSJobs(ctx, api.ATSJobsListParams{
				Status:       atsStatusFlag,
				DepartmentID: atsDepartmentIDFlag,
				LocationID:   atsLocationIDFlag,
				Limit:        limit,
				Cursor:       cursor,
			})
			if err != nil {
				return CursorListResult[api.ATSJob]{}, err
			}
			return CursorListResult[api.ATSJob]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing ats jobs")
		}

		response := makeListResponse(jobs, page)

		return outputList(cmd, f, jobs, hasMore, "No jobs found.", []string{"ID", "TITLE", "DEPARTMENT", "LOCATION", "TYPE", "STATUS"}, func(j api.ATSJob) []string {
			return []string{j.ID, j.Title, j.Department, j.Location, j.EmploymentType, j.Status}
		}, response)
	},
}

var atsJobsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new ATS job",
	Long:  "Create a new ATS job. Requires --title, --department-id, --location-id, and --employment-type flags.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := getFormatter()

		if atsJobTitleFlag == "" {
			return failValidation(cmd, f, "--title flag is required")
		}
		if atsJobDepartmentIDFlag == "" {
			return failValidation(cmd, f, "--department-id flag is required")
		}
		if atsJobLocationIDFlag == "" {
			return failValidation(cmd, f, "--location-id flag is required")
		}
		if atsJobEmploymentTypeFlag == "" {
			return failValidation(cmd, f, "--employment-type flag is required")
		}

		if ok, err := handleDryRun(cmd, f, &dryrun.Preview{
			Operation:   "CREATE",
			Resource:    "ATSJob",
			Description: "Create ATS job",
			Details: map[string]string{
				"Title":          atsJobTitleFlag,
				"DepartmentID":   atsJobDepartmentIDFlag,
				"LocationID":     atsJobLocationIDFlag,
				"EmploymentType": atsJobEmploymentTypeFlag,
				"Description":    atsJobDescriptionFlag,
			},
		}); ok {
			return err
		}

		client, err := getClient()
		if err != nil {
			return HandleError(f, err, "initializing client")
		}

		job, err := client.CreateATSJob(cmd.Context(), api.CreateATSJobParams{
			Title:          atsJobTitleFlag,
			DepartmentID:   atsJobDepartmentIDFlag,
			LocationID:     atsJobLocationIDFlag,
			EmploymentType: atsJobEmploymentTypeFlag,
			Description:    atsJobDescriptionFlag,
		})
		if err != nil {
			return HandleError(f, err, "create job")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
		client, f, err := initClient("listing ats job postings")
		if err != nil {
			return err
		}

		postings, page, hasMore, err := collectCursorItems(cmd.Context(), atsAllFlag, atsCursorFlag, atsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ATSJobPosting], error) {
			resp, err := client.ListATSJobPostings(ctx, api.ATSJobPostingsListParams{
				Status: atsStatusFlag,
				JobID:  atsJobIDFlag,
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.ATSJobPosting]{}, err
			}
			return CursorListResult[api.ATSJobPosting]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing ats job postings")
		}

		response := makeListResponse(postings, page)

		return outputList(cmd, f, postings, hasMore, "No job postings found.", []string{"ID", "TITLE", "DEPARTMENT", "LOCATION", "STATUS", "POSTED AT"}, func(p api.ATSJobPosting) []string {
			return []string{p.ID, p.Title, p.Department, p.Location, p.Status, p.PostedAt}
		}, response)
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
			return HandleError(f, err, "initializing client")
		}

		posting, err := client.GetATSJobPosting(cmd.Context(), args[0])
		if err != nil {
			return HandleError(f, err, "get job posting")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
			return HandleError(f, err, "initializing client")
		}

		apps, page, hasMore, err := collectCursorItems(cmd.Context(), atsAllFlag, atsCursorFlag, atsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ATSApplication], error) {
			resp, err := client.ListATSApplications(ctx, api.ATSApplicationsListParams{
				Status:      atsStatusFlag,
				JobID:       atsJobIDFlag,
				CandidateID: atsCandidateIDFlag,
				Stage:       atsStageFlag,
				Limit:       limit,
				Cursor:      cursor,
			})
			if err != nil {
				return CursorListResult[api.ATSApplication]{}, err
			}
			return CursorListResult[api.ATSApplication]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing ats applications")
		}

		response := makeListResponse(apps, page)

		return outputList(cmd, f, apps, hasMore, "No applications found.", []string{"ID", "CANDIDATE", "JOB", "STATUS", "STAGE", "APPLIED AT"}, func(a api.ATSApplication) []string {
			return []string{a.ID, a.CandidateName, a.JobTitle, a.Status, a.Stage, a.AppliedAt}
		}, response)
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
			return HandleError(f, err, "initializing client")
		}

		candidates, page, hasMore, err := collectCursorItems(cmd.Context(), atsAllFlag, atsCursorFlag, atsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ATSCandidate], error) {
			resp, err := client.ListATSCandidates(ctx, api.ATSCandidatesListParams{
				Search: atsSearchFlag,
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.ATSCandidate]{}, err
			}
			return CursorListResult[api.ATSCandidate]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing ats candidates")
		}

		response := makeListResponse(candidates, page)

		return outputList(cmd, f, candidates, hasMore, "No candidates found.", []string{"ID", "NAME", "EMAIL", "PHONE", "LOCATION"}, func(c api.ATSCandidate) []string {
			name := fmt.Sprintf("%s %s", c.FirstName, c.LastName)
			return []string{c.ID, name, c.Email, c.Phone, c.Location}
		}, response)
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
			return HandleError(f, err, "initializing client")
		}

		departments, page, hasMore, err := collectCursorItems(cmd.Context(), atsAllFlag, atsCursorFlag, atsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ATSDepartment], error) {
			resp, err := client.ListATSDepartments(ctx, api.ATSDepartmentsListParams{
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.ATSDepartment]{}, err
			}
			return CursorListResult[api.ATSDepartment]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing ats departments")
		}

		response := makeListResponse(departments, page)

		return outputList(cmd, f, departments, hasMore, "No departments found.", []string{"ID", "NAME", "PARENT ID", "CREATED AT"}, func(d api.ATSDepartment) []string {
			return []string{d.ID, d.Name, d.ParentID, d.CreatedAt}
		}, response)
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
			return HandleError(f, err, "initializing client")
		}

		var remotePtr *bool
		if cmd.Flags().Changed("remote") {
			remotePtr = &atsRemoteFlag
		}

		locations, page, hasMore, err := collectCursorItems(cmd.Context(), atsAllFlag, atsCursorFlag, atsLimitFlag, func(ctx context.Context, cursor string, limit int) (CursorListResult[api.ATSLocation], error) {
			resp, err := client.ListATSLocations(ctx, api.ATSLocationsListParams{
				Remote: remotePtr,
				Limit:  limit,
				Cursor: cursor,
			})
			if err != nil {
				return CursorListResult[api.ATSLocation]{}, err
			}
			return CursorListResult[api.ATSLocation]{
				Items: resp.Data,
				Page: CursorPage{
					Next:  resp.Page.Next,
					Total: resp.Page.Total,
				},
			}, nil
		})
		if err != nil {
			return HandleError(f, err, "listing ats locations")
		}

		response := makeListResponse(locations, page)

		return outputList(cmd, f, locations, hasMore, "No locations found.", []string{"ID", "NAME", "CITY", "COUNTRY", "REMOTE"}, func(l api.ATSLocation) []string {
			remote := "No"
			if l.Remote {
				remote = "Yes"
			}
			return []string{l.ID, l.Name, l.City, l.Country, remote}
		}, response)
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
			return HandleError(f, err, "initializing client")
		}

		reasons, err := client.ListRejectionReasons(cmd.Context())
		if err != nil {
			return HandleError(f, err, "list rejection reasons")
		}

		return f.OutputFiltered(cmd.Context(), func() {
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
	atsOffersCmd.Flags().IntVar(&atsLimitFlag, "limit", 100, "Maximum results")
	atsOffersCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")
	atsOffersCmd.Flags().BoolVar(&atsAllFlag, "all", false, "Fetch all pages")

	// Jobs list command flags
	atsJobsListCmd.Flags().StringVar(&atsStatusFlag, "status", "", "Filter by status")
	atsJobsListCmd.Flags().StringVar(&atsDepartmentIDFlag, "department-id", "", "Filter by department ID")
	atsJobsListCmd.Flags().StringVar(&atsLocationIDFlag, "location-id", "", "Filter by location ID")
	atsJobsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 100, "Maximum results")
	atsJobsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")
	atsJobsListCmd.Flags().BoolVar(&atsAllFlag, "all", false, "Fetch all pages")

	// Jobs create command flags
	atsJobsCreateCmd.Flags().StringVar(&atsJobTitleFlag, "title", "", "Job title (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobDepartmentIDFlag, "department-id", "", "Department ID (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobLocationIDFlag, "location-id", "", "Location ID (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobEmploymentTypeFlag, "employment-type", "", "Employment type (required)")
	atsJobsCreateCmd.Flags().StringVar(&atsJobDescriptionFlag, "description", "", "Job description (optional)")

	// Job postings list command flags
	atsPostingsListCmd.Flags().StringVar(&atsStatusFlag, "status", "", "Filter by status")
	atsPostingsListCmd.Flags().StringVar(&atsJobIDFlag, "job-id", "", "Filter by job ID")
	atsPostingsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 100, "Maximum results")
	atsPostingsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")
	atsPostingsListCmd.Flags().BoolVar(&atsAllFlag, "all", false, "Fetch all pages")

	// Applications list command flags
	atsApplicationsListCmd.Flags().StringVar(&atsStatusFlag, "status", "", "Filter by status")
	atsApplicationsListCmd.Flags().StringVar(&atsJobIDFlag, "job-id", "", "Filter by job ID")
	atsApplicationsListCmd.Flags().StringVar(&atsCandidateIDFlag, "candidate-id", "", "Filter by candidate ID")
	atsApplicationsListCmd.Flags().StringVar(&atsStageFlag, "stage", "", "Filter by stage")
	atsApplicationsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 100, "Maximum results")
	atsApplicationsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")
	atsApplicationsListCmd.Flags().BoolVar(&atsAllFlag, "all", false, "Fetch all pages")

	// Candidates list command flags
	atsCandidatesListCmd.Flags().StringVar(&atsSearchFlag, "search", "", "Search candidates by name or email")
	atsCandidatesListCmd.Flags().IntVar(&atsLimitFlag, "limit", 100, "Maximum results")
	atsCandidatesListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")
	atsCandidatesListCmd.Flags().BoolVar(&atsAllFlag, "all", false, "Fetch all pages")

	// Departments list command flags
	atsDepartmentsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 100, "Maximum results")
	atsDepartmentsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")
	atsDepartmentsListCmd.Flags().BoolVar(&atsAllFlag, "all", false, "Fetch all pages")

	// Locations list command flags
	atsLocationsListCmd.Flags().BoolVar(&atsRemoteFlag, "remote", false, "Filter by remote locations")
	atsLocationsListCmd.Flags().IntVar(&atsLimitFlag, "limit", 100, "Maximum results")
	atsLocationsListCmd.Flags().StringVar(&atsCursorFlag, "cursor", "", "Pagination cursor")
	atsLocationsListCmd.Flags().BoolVar(&atsAllFlag, "all", false, "Fetch all pages")

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
