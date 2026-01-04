package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListATSJobs(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":              "job1",
				"title":           "Software Engineer",
				"department":      "Engineering",
				"department_id":   "dept1",
				"location":        "San Francisco",
				"location_id":     "loc1",
				"status":          "open",
				"employment_type": "full-time",
				"created_at":      "2024-01-01T00:00:00Z",
			},
		},
	}
	server := mockServerWithQuery(t, "/rest/v2/ats/jobs", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "open", query["status"])
		assert.Equal(t, "dept1", query["department_id"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListATSJobs(context.Background(), ATSJobsListParams{
		Status:       "open",
		DepartmentID: "dept1",
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "job1", result.Data[0].ID)
	assert.Equal(t, "Software Engineer", result.Data[0].Title)
	assert.Equal(t, "Engineering", result.Data[0].Department)
}

func TestCreateATSJob(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/ats/jobs", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Backend Engineer", body["title"])
		assert.Equal(t, "dept2", body["department_id"])
		assert.Equal(t, "full-time", body["employment_type"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":              "job-new",
			"title":           "Backend Engineer",
			"department_id":   "dept2",
			"location_id":     "loc2",
			"employment_type": "full-time",
			"status":          "draft",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateATSJob(context.Background(), CreateATSJobParams{
		Title:          "Backend Engineer",
		DepartmentID:   "dept2",
		LocationID:     "loc2",
		EmploymentType: "full-time",
	})

	require.NoError(t, err)
	assert.Equal(t, "job-new", result.ID)
	assert.Equal(t, "Backend Engineer", result.Title)
	assert.Equal(t, "draft", result.Status)
}

func TestListATSJobPostings(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":              "posting1",
				"job_id":          "job1",
				"title":           "Senior Developer",
				"description":     "We are looking for a senior developer",
				"department":      "Engineering",
				"location":        "Remote",
				"employment_type": "full-time",
				"status":          "published",
				"posted_at":       "2024-01-15T00:00:00Z",
				"url":             "https://careers.example.com/jobs/posting1",
			},
		},
	}
	server := mockServerWithQuery(t, "/rest/v2/ats/job-postings", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "published", query["status"])
		assert.Equal(t, "job1", query["job_id"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListATSJobPostings(context.Background(), ATSJobPostingsListParams{
		Status: "published",
		JobID:  "job1",
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "posting1", result.Data[0].ID)
	assert.Equal(t, "Senior Developer", result.Data[0].Title)
	assert.Equal(t, "published", result.Data[0].Status)
}

func TestGetATSJobPosting(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":              "posting1",
			"job_id":          "job1",
			"title":           "Frontend Developer",
			"description":     "Build amazing UIs",
			"department":      "Engineering",
			"location":        "New York",
			"employment_type": "contract",
			"status":          "published",
			"posted_at":       "2024-02-01T00:00:00Z",
			"url":             "https://careers.example.com/jobs/posting1",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/ats/job-postings/posting1", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetATSJobPosting(context.Background(), "posting1")

	require.NoError(t, err)
	assert.Equal(t, "posting1", result.ID)
	assert.Equal(t, "Frontend Developer", result.Title)
	assert.Equal(t, "Build amazing UIs", result.Description)
	assert.Equal(t, "contract", result.EmploymentType)
}

func TestListATSApplications(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":             "app1",
				"candidate_id":   "cand1",
				"candidate_name": "John Doe",
				"job_id":         "job1",
				"job_title":      "Software Engineer",
				"status":         "active",
				"stage":          "phone-screen",
				"applied_at":     "2024-01-10T00:00:00Z",
				"updated_at":     "2024-01-11T00:00:00Z",
			},
		},
	}
	server := mockServerWithQuery(t, "/rest/v2/ats/applications", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "active", query["status"])
		assert.Equal(t, "job1", query["job_id"])
		assert.Equal(t, "phone-screen", query["stage"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListATSApplications(context.Background(), ATSApplicationsListParams{
		Status: "active",
		JobID:  "job1",
		Stage:  "phone-screen",
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "app1", result.Data[0].ID)
	assert.Equal(t, "John Doe", result.Data[0].CandidateName)
	assert.Equal(t, "phone-screen", result.Data[0].Stage)
}

func TestListATSCandidates(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":           "cand1",
				"first_name":   "Jane",
				"last_name":    "Smith",
				"email":        "jane.smith@example.com",
				"phone":        "+1234567890",
				"location":     "San Francisco, CA",
				"linkedin_url": "https://linkedin.com/in/janesmith",
				"resume_url":   "https://storage.example.com/resumes/cand1.pdf",
				"created_at":   "2024-01-05T00:00:00Z",
			},
		},
	}
	server := mockServerWithQuery(t, "/rest/v2/ats/candidates", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "Jane Smith", query["search"])
		assert.Equal(t, "50", query["limit"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListATSCandidates(context.Background(), ATSCandidatesListParams{
		Search: "Jane Smith",
		Limit:  50,
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "cand1", result.Data[0].ID)
	assert.Equal(t, "Jane", result.Data[0].FirstName)
	assert.Equal(t, "Smith", result.Data[0].LastName)
	assert.Equal(t, "jane.smith@example.com", result.Data[0].Email)
}

func TestListATSDepartments(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":         "dept1",
				"name":       "Engineering",
				"created_at": "2023-01-01T00:00:00Z",
			},
			{
				"id":         "dept2",
				"name":       "Product",
				"parent_id":  "dept1",
				"created_at": "2023-01-15T00:00:00Z",
			},
		},
	}
	server := mockServerWithQuery(t, "/rest/v2/ats/departments", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "100", query["limit"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListATSDepartments(context.Background(), ATSDepartmentsListParams{
		Limit: 100,
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "dept1", result.Data[0].ID)
	assert.Equal(t, "Engineering", result.Data[0].Name)
	assert.Equal(t, "dept2", result.Data[1].ID)
	assert.Equal(t, "dept1", result.Data[1].ParentID)
}

func TestListATSLocations(t *testing.T) {
	t.Run("filter remote=true", func(t *testing.T) {
		response := map[string]any{
			"data": []map[string]any{
				{
					"id":         "loc2",
					"name":       "Remote - US",
					"city":       "",
					"country":    "USA",
					"remote":     true,
					"created_at": "2023-01-01T00:00:00Z",
				},
			},
		}
		server := mockServerWithQuery(t, "/rest/v2/ats/locations", func(t *testing.T, query map[string]string) {
			assert.Equal(t, "true", query["remote"])
		}, http.StatusOK, response)
		defer server.Close()

		client := testClient(server)
		remoteTrue := true
		result, err := client.ListATSLocations(context.Background(), ATSLocationsListParams{
			Remote: &remoteTrue,
		})

		require.NoError(t, err)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, "loc2", result.Data[0].ID)
		assert.True(t, result.Data[0].Remote)
	})

	t.Run("filter remote=false", func(t *testing.T) {
		response := map[string]any{
			"data": []map[string]any{
				{
					"id":         "loc1",
					"name":       "San Francisco Office",
					"city":       "San Francisco",
					"country":    "USA",
					"remote":     false,
					"created_at": "2023-01-01T00:00:00Z",
				},
			},
		}
		server := mockServerWithQuery(t, "/rest/v2/ats/locations", func(t *testing.T, query map[string]string) {
			assert.Equal(t, "false", query["remote"])
		}, http.StatusOK, response)
		defer server.Close()

		client := testClient(server)
		remoteFalse := false
		result, err := client.ListATSLocations(context.Background(), ATSLocationsListParams{
			Remote: &remoteFalse,
		})

		require.NoError(t, err)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, "loc1", result.Data[0].ID)
		assert.False(t, result.Data[0].Remote)
	})

	t.Run("no filter - all locations", func(t *testing.T) {
		response := map[string]any{
			"data": []map[string]any{
				{
					"id":         "loc1",
					"name":       "San Francisco Office",
					"city":       "San Francisco",
					"country":    "USA",
					"remote":     false,
					"created_at": "2023-01-01T00:00:00Z",
				},
				{
					"id":         "loc2",
					"name":       "Remote - US",
					"city":       "",
					"country":    "USA",
					"remote":     true,
					"created_at": "2023-01-01T00:00:00Z",
				},
			},
		}
		server := mockServerWithQuery(t, "/rest/v2/ats/locations", func(t *testing.T, query map[string]string) {
			_, hasRemote := query["remote"]
			assert.False(t, hasRemote, "remote parameter should not be present")
		}, http.StatusOK, response)
		defer server.Close()

		client := testClient(server)
		result, err := client.ListATSLocations(context.Background(), ATSLocationsListParams{
			Remote: nil,
		})

		require.NoError(t, err)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, "loc1", result.Data[0].ID)
		assert.False(t, result.Data[0].Remote)
		assert.Equal(t, "loc2", result.Data[1].ID)
		assert.True(t, result.Data[1].Remote)
	})
}

func TestListRejectionReasons(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":          "reason1",
				"reason":      "Not qualified",
				"description": "Candidate does not meet minimum requirements",
			},
			{
				"id":          "reason2",
				"reason":      "Position filled",
				"description": "The position has been filled by another candidate",
			},
		},
	}
	server := mockServer(t, "GET", "/rest/v2/ats/rejection-reasons", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListRejectionReasons(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "reason1", result[0].ID)
	assert.Equal(t, "Not qualified", result[0].Reason)
	assert.Equal(t, "reason2", result[1].ID)
	assert.Equal(t, "Position filled", result[1].Reason)
}

func TestListATSOffers(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":             "offer1",
				"candidate_id":   "cand1",
				"candidate_name": "Alice Johnson",
				"position":       "Senior Engineer",
				"status":         "pending",
				"salary":         150000.00,
				"currency":       "USD",
				"start_date":     "2024-03-01",
				"created_at":     "2024-01-20T00:00:00Z",
			},
		},
	}
	server := mockServerWithQuery(t, "/rest/v2/ats/offers", func(t *testing.T, query map[string]string) {
		assert.Equal(t, "pending", query["status"])
	}, http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListATSOffers(context.Background(), ATSOffersListParams{
		Status: "pending",
	})

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "offer1", result.Data[0].ID)
	assert.Equal(t, "Alice Johnson", result.Data[0].Candidate)
	assert.Equal(t, "pending", result.Data[0].Status)
	assert.Equal(t, 150000.00, result.Data[0].Salary)
}
