package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContract_UnmarshalJSON_PopulatesWorkerAlias(t *testing.T) {
	raw := []byte(`{
  "id": "c1",
  "title": "Contractor Agreement",
  "type": "pay_as_you_go",
  "status": "active",
  "start_date": "2024-01-01",
  "termination_date": "2024-03-01",
  "client": {
    "legal_entity": {
      "id": "le-1",
      "name": "Acme, Inc."
    }
  },
  "worker": {
    "full_name": "Zora Example",
    "email": "zora@example.com",
    "country": "US"
  },
  "compensation_details": {
    "currency_code": "USD",
    "amount": "100.50"
  }
}`)

	var c Contract
	require.NoError(t, json.Unmarshal(raw, &c))

	assert.Equal(t, "c1", c.ID)
	assert.Equal(t, "Contractor Agreement", c.Title)
	assert.Equal(t, "pay_as_you_go", c.Type)
	assert.Equal(t, "active", c.Status)
	assert.Equal(t, "2024-01-01", c.StartDate)
	assert.Equal(t, "2024-03-01", c.EndDate)

	// Backwards compatible flat fields.
	assert.Equal(t, "Zora Example", c.WorkerName)
	assert.Equal(t, "zora@example.com", c.WorkerEmail)
	assert.Equal(t, "US", c.Country)

	// New ergonomic alias for scripting.
	assert.Equal(t, ContractWorker{
		Name:    "Zora Example",
		Email:   "zora@example.com",
		Country: "US",
	}, c.Worker)

	// Ensure we actually emit the alias in JSON output.
	out, err := json.Marshal(c)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(out, &m))
	assert.Equal(t, "Zora Example", m["worker_name"])
	require.Contains(t, m, "worker")
	worker, ok := m["worker"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Zora Example", worker["name"])
	assert.Equal(t, "zora@example.com", worker["email"])
	assert.Equal(t, "US", worker["country"])
}

func TestListContracts(t *testing.T) {
	response := map[string]any{
		"data": []map[string]any{
			{
				"id":     "c1",
				"title":  "Contract 1",
				"status": "active",
				"client": map[string]any{
					"legal_entity": map[string]any{
						"id":   "le-1",
						"name": "Taiwan Entity",
					},
				},
			},
			{"id": "c2", "title": "Contract 2", "status": "pending"},
		},
		"page": map[string]any{"next": "cursor123", "total": 2},
	}
	server := mockServer(t, "GET", "/rest/v2/contracts", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.ListContracts(context.Background(), ContractsListParams{})

	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "c1", result.Data[0].ID)
	assert.Equal(t, "le-1", result.Data[0].EntityID)
	assert.Equal(t, "Taiwan Entity", result.Data[0].Entity)
	assert.Equal(t, "cursor123", result.Page.Next)
}

func TestGetContract(t *testing.T) {
	response := map[string]any{
		"data": map[string]any{
			"id":     "c1",
			"title":  "Test Contract",
			"status": "active",
		},
	}
	server := mockServer(t, "GET", "/rest/v2/contracts/c1", http.StatusOK, response)
	defer server.Close()

	client := testClient(server)
	result, err := client.GetContract(context.Background(), "c1")

	require.NoError(t, err)
	assert.Equal(t, "c1", result.ID)
	assert.Equal(t, "Test Contract", result.Title)
}

func TestCreateContract(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/contracts", func(t *testing.T, body map[string]any) {
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")
		assert.Equal(t, "New Contract", data["title"])
		assert.Equal(t, "fixed_rate", data["type"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":     "c-new",
			"title":  "New Contract",
			"type":   "fixed_rate",
			"status": "draft",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateContract(context.Background(), CreateContractParams{
		Title:       "New Contract",
		Type:        "fixed_rate",
		WorkerEmail: "worker@example.com",
		Currency:    "USD",
		Rate:        5000,
		Country:     "US",
	})

	require.NoError(t, err)
	assert.Equal(t, "c-new", result.ID)
}

func TestSignContract(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/contracts/c1/signatures", func(t *testing.T, body map[string]any) {
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")
		assert.Equal(t, "John Smith", data["client_signature"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":     "c1",
			"status": "signed",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.SignContract(context.Background(), "c1", "John Smith")

	require.NoError(t, err)
	assert.Equal(t, "signed", result.Status)
}

func TestTerminateContract(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/contracts/c1/terminations", func(t *testing.T, body map[string]any) {
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")
		assert.Equal(t, true, data["terminate_now"])
		assert.Equal(t, "TERMINATION", data["termination_type"])
		assert.Equal(t, "reason-123", data["termination_reason_id"])
	}, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":     "term-1",
			"status": "pending",
		},
	})
	defer server.Close()

	client := testClient(server)
	err := client.TerminateContract(context.Background(), "c1", TerminateContractParams{
		TerminateNow:        true,
		TerminationType:     "TERMINATION",
		TerminationReasonID: "reason-123",
	})

	require.NoError(t, err)
}

func TestGetContractPDF(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/contracts/c1/pdf", http.StatusOK, map[string]any{
		"data": map[string]any{
			"url": "https://storage.deel.com/contracts/c1.pdf",
		},
	})
	defer server.Close()

	client := testClient(server)
	url, err := client.GetContractPDF(context.Background(), "c1")

	require.NoError(t, err)
	assert.Contains(t, url, "c1.pdf")
}

func TestInviteWorker(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/contracts/c1/invitations", func(t *testing.T, body map[string]any) {
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")
		assert.Equal(t, "worker@example.com", data["email"])
		assert.Equal(t, "en", data["locale"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"invited": true,
		},
	})
	defer server.Close()

	client := testClient(server)
	err := client.InviteWorker(context.Background(), "c1", InviteWorkerParams{
		Email:  "worker@example.com",
		Locale: "en",
	})

	require.NoError(t, err)
}

func TestGetInviteLink(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/contracts/c1/invite-link", http.StatusOK, map[string]any{
		"data": map[string]any{
			"url": "https://app.deel.com/invite/abc123",
		},
	})
	defer server.Close()

	client := testClient(server)
	url, err := client.GetInviteLink(context.Background(), "c1")

	require.NoError(t, err)
	assert.Contains(t, url, "invite")
}

func TestListContractTemplates(t *testing.T) {
	server := mockServer(t, "GET", "/rest/v2/contract-templates", http.StatusOK, map[string]any{
		"data": []map[string]any{
			{"id": "tpl1", "name": "Standard Contractor", "type": "fixed_rate"},
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.ListContractTemplates(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "tpl1", result[0].ID)
}

func TestCreateContractWithExtendedFields(t *testing.T) {
	server := mockServerWithBody(t, "POST", "/rest/v2/contracts", func(t *testing.T, body map[string]any) {
		data, ok := body["data"].(map[string]any)
		require.True(t, ok, "body should have 'data' wrapper")

		assert.Equal(t, "Host Contract", data["title"])
		assert.Equal(t, "payg_tasks", data["type"])

		assert.Equal(t, "tpl-host-ca", data["contract_template_id"])

		client, ok := data["client"].(map[string]any)
		require.True(t, ok, "data should have 'client' object")

		legalEntity, ok := client["legal_entity"].(map[string]any)
		require.True(t, ok, "client should have 'legal_entity' object")
		assert.Equal(t, "le-123", legalEntity["id"])

		team, ok := client["team"].(map[string]any)
		require.True(t, ok, "client should have 'team' object")
		assert.Equal(t, "team-456", team["id"])

		comp, ok := data["compensation_details"].(map[string]any)
		require.True(t, ok, "data should have 'compensation_details' object")
		assert.Equal(t, float64(5), comp["cycle_end"])
		assert.Equal(t, "DAY_OF_MONTH", comp["cycle_end_type"])
		assert.Equal(t, "monthly", comp["frequency"])
	}, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"id":     "c-new",
			"title":  "Host Contract",
			"type":   "pay_as_you_go_time_based",
			"status": "draft",
		},
	})
	defer server.Close()

	client := testClient(server)
	result, err := client.CreateContract(context.Background(), CreateContractParams{
		Title:         "Host Contract",
		Type:          "pay_as_you_go_time_based",
		WorkerEmail:   "worker@example.com",
		Currency:      "CAD",
		Rate:          23.00,
		Country:       "CA",
		TemplateID:    "tpl-host-ca",
		LegalEntityID: "le-123",
		GroupID:       "team-456",
		CycleEnd:      5,
		CycleEndType:  "DAY_OF_MONTH",
		Frequency:     "monthly",
	})

	require.NoError(t, err)
	assert.Equal(t, "c-new", result.ID)
}
