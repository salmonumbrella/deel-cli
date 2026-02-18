package cmd

import (
	"fmt"
	"unicode/utf8"

	"github.com/spf13/pflag"

	"github.com/salmonumbrella/deel-cli/internal/api"
)

// truncateRunes truncates s to at most max runes, appending "..." if truncated.
// It is safe for multi-byte UTF-8 strings.
func truncateRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "..."
}

// flagAlias registers a hidden alias for an existing flag.
// Both flags share the same underlying Value, so setting either one sets both.
func flagAlias(fs *pflag.FlagSet, name, alias string) { //nolint:unparam // general-purpose helper
	f := fs.Lookup(name)
	if f == nil {
		panic(fmt.Sprintf("flagAlias: flag %q not found", name))
	}
	a := *f // shallow copy -- shares the Value interface
	a.Name = alias
	a.Shorthand = ""
	a.Usage = ""
	a.Hidden = true
	fs.AddFlag(&a)
}

// PersonLight is a compact person summary for agent workflows.
// Full: ~20 lines (with nested employments). Light: ~6 lines.
type PersonLight struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	JobTitle string `json:"job_title"`
	Status   string `json:"status"`
	Country  string `json:"country"`
}

func toLightPerson(p api.Person) PersonLight {
	return PersonLight{
		ID:       p.HRISProfileID,
		Name:     truncateRunes(p.Name, 60),
		Email:    truncateRunes(p.Email, 80),
		JobTitle: truncateRunes(p.JobTitle, 80),
		Status:   p.Status,
		Country:  p.Country,
	}
}

func toLightPeople(people []api.Person) []PersonLight {
	out := make([]PersonLight, len(people))
	for i, p := range people {
		out[i] = toLightPerson(p)
	}
	return out
}

// ContractLight is a compact contract summary for agent workflows.
// Full: ~18 lines (with nested worker object). Light: ~7 lines.
type ContractLight struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	WorkerName string `json:"worker_name"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date,omitempty"`
}

func toLightContract(c api.Contract) ContractLight {
	return ContractLight{
		ID:         c.ID,
		Title:      truncateRunes(c.Title, 80),
		Status:     c.Status,
		Type:       c.Type,
		WorkerName: truncateRunes(c.WorkerName, 60),
		StartDate:  c.StartDate,
		EndDate:    c.EndDate,
	}
}

func toLightContracts(contracts []api.Contract) []ContractLight {
	out := make([]ContractLight, len(contracts))
	for i, c := range contracts {
		out[i] = toLightContract(c)
	}
	return out
}
