package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGPContract(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/gp/contracts", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "jane.smith@example.com", body["worker_email"])
		assert.Equal(t, "Jane Smith", body["worker_name"])
		assert.Equal(t, "GB", body["country"])
		assert.Equal(t, "2024-03-01", body["start_date"])
		assert.Equal(t, "Payroll Manager", body["job_title"])
		assert.Equal(t, "Finance", body["department"])
		assert.Equal(t, 80000.0, body["salary"])
		assert.Equal(t, "GBP", body["currency"])
		assert.Equal(t, "monthly", body["pay_frequency"])
	}, 201, map[string]any{
		"data": map[string]any{
			"id":            "gp-789",
			"worker_id":     "w-789",
			"worker_name":   "Jane Smith",
			"worker_email":  "jane.smith@example.com",
			"country":       "GB",
			"start_date":    "2024-03-01",
			"status":        "active",
			"job_title":     "Payroll Manager",
			"department":    "Finance",
			"salary":        80000.0,
			"currency":      "GBP",
			"pay_frequency": "monthly",
			"created_at":    "2024-02-20T09:00:00Z",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateGPContract(context.Background(), CreateGPContractParams{
		WorkerEmail:  "jane.smith@example.com",
		WorkerName:   "Jane Smith",
		Country:      "GB",
		StartDate:    "2024-03-01",
		JobTitle:     "Payroll Manager",
		Department:   "Finance",
		Salary:       80000.0,
		Currency:     "GBP",
		PayFrequency: "monthly",
	})

	require.NoError(t, err)
	assert.Equal(t, "gp-789", result.ID)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "jane.smith@example.com", result.WorkerEmail)
	assert.Equal(t, "Jane Smith", result.WorkerName)
	assert.Equal(t, "Payroll Manager", result.JobTitle)
	assert.Equal(t, "Finance", result.Department)
	assert.Equal(t, 80000.0, result.Salary)
	assert.Equal(t, "GBP", result.Currency)
	assert.Equal(t, "monthly", result.PayFrequency)
}

func TestCreateGPContract_ValidationError(t *testing.T) {
	server := mockServer(t, "POST", "/rest/v2/gp/contracts", 400, map[string]string{"error": "invalid country code"})
	defer server.Close()

	client := testClient(server)
	_, err := client.CreateGPContract(context.Background(), CreateGPContractParams{
		WorkerEmail:  "test@example.com",
		WorkerName:   "Test User",
		Country:      "INVALID",
		StartDate:    "2024-03-01",
		JobTitle:     "Manager",
		Salary:       50000.0,
		Currency:     "USD",
		PayFrequency: "monthly",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestUpdateGPWorker(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/gp/workers/w-789", func(t *testing.T, body map[string]any) {
		assert.Equal(t, "Jane", body["first_name"])
		assert.Equal(t, "Smith-Jones", body["last_name"])
		assert.Equal(t, "+44 20 7123 4567", body["phone"])
		assert.Equal(t, "123 Main St, London", body["address"])
		assert.Equal(t, "GB123456789", body["tax_id"])
	}, 200, map[string]any{
		"data": map[string]any{
			"id":            "w-789",
			"email":         "jane.smith@example.com",
			"first_name":    "Jane",
			"last_name":     "Smith-Jones",
			"country":       "GB",
			"date_of_birth": "1990-05-15",
			"phone":         "+44 20 7123 4567",
			"address":       "123 Main St, London",
			"tax_id":        "GB123456789",
			"status":        "active",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGPWorker(context.Background(), "w-789", UpdateGPWorkerParams{
		FirstName: "Jane",
		LastName:  "Smith-Jones",
		Phone:     "+44 20 7123 4567",
		Address:   "123 Main St, London",
		TaxID:     "GB123456789",
	})

	require.NoError(t, err)
	assert.Equal(t, "w-789", result.ID)
	assert.Equal(t, "Jane", result.FirstName)
	assert.Equal(t, "Smith-Jones", result.LastName)
	assert.Equal(t, "+44 20 7123 4567", result.Phone)
	assert.Equal(t, "123 Main St, London", result.Address)
	assert.Equal(t, "GB123456789", result.TaxID)
	assert.Equal(t, "active", result.Status)
}

func TestUpdateGPWorker_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/gp/workers/invalid", 404, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateGPWorker(context.Background(), "invalid", UpdateGPWorkerParams{
		Phone: "+1 555 1234",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestUpdateGPCompensation(t *testing.T) {
	server := mockServerWithBody(t, "PATCH", "/rest/v2/gp/workers/w-789/compensation", func(t *testing.T, body map[string]any) {
		assert.Equal(t, 90000.0, body["salary"])
		assert.Equal(t, "GBP", body["currency"])
		assert.Equal(t, "monthly", body["pay_frequency"])
		assert.Equal(t, "2024-04-01", body["effective_date"])
		allowances, ok := body["allowances"].([]any)
		require.True(t, ok)
		require.Len(t, allowances, 2)
		allow1 := allowances[0].(map[string]any)
		assert.Equal(t, "Housing Allowance", allow1["name"])
		assert.Equal(t, 1000.0, allow1["amount"])
		assert.Equal(t, "GBP", allow1["currency"])
		allow2 := allowances[1].(map[string]any)
		assert.Equal(t, "Transport Allowance", allow2["name"])
		assert.Equal(t, 200.0, allow2["amount"])
		assert.Equal(t, "GBP", allow2["currency"])
	}, 200, map[string]any{
		"data": map[string]any{
			"worker_id":      "w-789",
			"salary":         90000.0,
			"currency":       "GBP",
			"pay_frequency":  "monthly",
			"effective_date": "2024-04-01",
			"allowances": []map[string]any{
				{
					"name":     "Housing Allowance",
					"amount":   1000.0,
					"currency": "GBP",
				},
				{
					"name":     "Transport Allowance",
					"amount":   200.0,
					"currency": "GBP",
				},
			},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.UpdateGPCompensation(context.Background(), "w-789", UpdateGPCompensationParams{
		Salary:        90000.0,
		Currency:      "GBP",
		PayFrequency:  "monthly",
		EffectiveDate: "2024-04-01",
		Allowances: []GPAllowance{
			{
				Name:     "Housing Allowance",
				Amount:   1000.0,
				Currency: "GBP",
			},
			{
				Name:     "Transport Allowance",
				Amount:   200.0,
				Currency: "GBP",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "w-789", result.WorkerID)
	assert.Equal(t, 90000.0, result.Salary)
	assert.Equal(t, "GBP", result.Currency)
	assert.Equal(t, "monthly", result.PayFrequency)
	assert.Equal(t, "2024-04-01", result.EffectiveDate)
	require.Len(t, result.Allowances, 2)
	assert.Equal(t, "Housing Allowance", result.Allowances[0].Name)
	assert.Equal(t, 1000.0, result.Allowances[0].Amount)
	assert.Equal(t, "Transport Allowance", result.Allowances[1].Name)
	assert.Equal(t, 200.0, result.Allowances[1].Amount)
}

func TestUpdateGPCompensation_NotFound(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/gp/workers/invalid/compensation", 404, map[string]string{"error": "worker not found"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateGPCompensation(context.Background(), "invalid", UpdateGPCompensationParams{
		Salary:        95000.0,
		EffectiveDate: "2024-05-01",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestUpdateGPCompensation_InvalidEffectiveDate(t *testing.T) {
	server := mockServer(t, "PATCH", "/rest/v2/gp/workers/w-789/compensation", 400, map[string]string{"error": "effective_date is required"})
	defer server.Close()

	client := testClient(server)
	_, err := client.UpdateGPCompensation(context.Background(), "w-789", UpdateGPCompensationParams{
		Salary:        95000.0,
		EffectiveDate: "",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, 400, apiErr.StatusCode)
}
