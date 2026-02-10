package api

import (
	"encoding/json"
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
