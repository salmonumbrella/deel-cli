package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

func TestToLightPerson(t *testing.T) {
	person := api.Person{
		ID:            "123",
		HRISProfileID: "hris-456",
		FirstName:     "John",
		LastName:      "Doe",
		Name:          "John Doe",
		Email:         "john@example.com",
		JobTitle:      "Engineer",
		Status:        "active",
		Country:       "US",
		StartDate:     "2024-01-15",
		HiringType:    "contractor",
		Employments: []api.Employment{
			{ID: "e1", Name: "Corp", ContractStatus: "active"},
		},
	}

	light := toLightPerson(person)
	assert.Equal(t, "hris-456", light.ID)
	assert.Equal(t, "John Doe", light.Name)
	assert.Equal(t, "john@example.com", light.Email)
	assert.Equal(t, "Engineer", light.JobTitle)
	assert.Equal(t, "active", light.Status)
	assert.Equal(t, "US", light.Country)
}

func TestToLightPeople(t *testing.T) {
	people := []api.Person{
		{HRISProfileID: "1", Name: "Alice", Email: "alice@example.com", Status: "active"},
		{HRISProfileID: "2", Name: "Bob", Email: "bob@example.com", Status: "inactive"},
	}

	light := toLightPeople(people)
	assert.Len(t, light, 2)
	assert.Equal(t, "1", light[0].ID)
	assert.Equal(t, "2", light[1].ID)
}

func TestToLightContract(t *testing.T) {
	contract := api.Contract{
		ID:                 "c-123",
		Title:              "Engineering Contract",
		Type:               "fixed_rate",
		Status:             "active",
		WorkerName:         "John Doe",
		WorkerEmail:        "john@example.com",
		Entity:             "Acme Corp",
		EntityID:           "ent-1",
		StartDate:          "2024-01-15",
		EndDate:            "2025-01-15",
		Currency:           "USD",
		CompensationAmount: 5000.00,
		Country:            "US",
	}

	light := toLightContract(contract)
	assert.Equal(t, "c-123", light.ID)
	assert.Equal(t, "Engineering Contract", light.Title)
	assert.Equal(t, "active", light.Status)
	assert.Equal(t, "fixed_rate", light.Type)
	assert.Equal(t, "John Doe", light.WorkerName)
	assert.Equal(t, "2024-01-15", light.StartDate)
	assert.Equal(t, "2025-01-15", light.EndDate)
}

func TestToLightContracts(t *testing.T) {
	contracts := []api.Contract{
		{ID: "1", Title: "Contract A", Status: "active"},
		{ID: "2", Title: "Contract B", Status: "cancelled"},
	}

	light := toLightContracts(contracts)
	assert.Len(t, light, 2)
	assert.Equal(t, "1", light[0].ID)
	assert.Equal(t, "2", light[1].ID)
}

func TestTruncateRunes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello..."},
		{"empty", "", 5, ""},
		{"zero max", "hello", 0, "hello"},
		{"negative max", "hello", -1, "hello"},
		{"multibyte", "cafe\u0301 latte", 5, "cafe\u0301..."},
		{"cjk", "\u4e16\u754c\u4f60\u597d\u5417", 3, "\u4e16\u754c\u4f60..."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, truncateRunes(tc.in, tc.max))
		})
	}
}

func TestToLightPerson_Truncation(t *testing.T) {
	longName := strings.Repeat("A", 100)
	person := api.Person{
		HRISProfileID: "1",
		Name:          longName,
		Email:         strings.Repeat("x", 100) + "@example.com",
		JobTitle:      strings.Repeat("B", 100),
		Status:        "active",
		Country:       "US",
	}
	light := toLightPerson(person)
	assert.Len(t, []rune(strings.TrimSuffix(light.Name, "...")), 60)
	assert.Contains(t, light.Name, "...")
	assert.Contains(t, light.Email, "...")
	assert.Contains(t, light.JobTitle, "...")
}

func TestToLightContract_Truncation(t *testing.T) {
	contract := api.Contract{
		ID:         "c-1",
		Title:      strings.Repeat("T", 100),
		Status:     "active",
		Type:       "fixed_rate",
		WorkerName: strings.Repeat("W", 100),
		StartDate:  "2024-01-01",
	}
	light := toLightContract(contract)
	assert.Contains(t, light.Title, "...")
	assert.Contains(t, light.WorkerName, "...")
}

func TestFlagAlias(t *testing.T) {
	// Verify flagAlias creates a hidden alias that shares the same value.
	cmd := peopleListCmd
	f := cmd.Flags().Lookup("light")
	assert.NotNil(t, f, "people list should have --light flag")

	li := cmd.Flags().Lookup("li")
	assert.NotNil(t, li, "people list should have --li alias")
	assert.True(t, li.Hidden, "--li should be hidden")
}
