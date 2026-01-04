package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/muesli/termenv"

	"github.com/salmonumbrella/deel-cli/internal/dryrun"
	"github.com/salmonumbrella/deel-cli/internal/filter"
)

// Format represents the output format
type Format string

const (
	// FormatText renders human-readable output.
	FormatText Format = "text"
	// FormatJSON renders JSON output.
	FormatJSON Format = "json"
)

// Formatter handles output formatting
type Formatter struct {
	out       io.Writer
	errOut    io.Writer
	format    Format
	colorMode string
	profile   termenv.Profile
	query     string
}

// New creates a new Formatter
func New(out, errOut io.Writer, format Format, colorMode string) *Formatter {
	f := &Formatter{
		out:       out,
		errOut:    errOut,
		format:    format,
		colorMode: colorMode,
	}
	f.profile = f.detectColorProfile()
	return f
}

// SetQuery sets an optional JQ-style query for JSON output.
func (f *Formatter) SetQuery(query string) {
	f.query = strings.TrimSpace(query)
}

func (f *Formatter) detectColorProfile() termenv.Profile {
	switch f.colorMode {
	case "never":
		return termenv.Ascii
	case "always":
		return termenv.TrueColor
	default: // "auto"
		// Check NO_COLOR environment variable
		if os.Getenv("NO_COLOR") != "" {
			return termenv.Ascii
		}
		return termenv.ColorProfile()
	}
}

// IsJSON returns true if output format is JSON
func (f *Formatter) IsJSON() bool {
	return f.format == FormatJSON
}

// PrintJSON outputs data as JSON
func (f *Formatter) PrintJSON(data any) error {
	enc := json.NewEncoder(f.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// PrintText outputs plain text
func (f *Formatter) PrintText(text string) {
	if _, err := fmt.Fprintln(f.out, text); err != nil {
		return
	}
}

// PrintError outputs an error message to stderr
func (f *Formatter) PrintError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if f.profile != termenv.Ascii {
		msg = termenv.String(msg).Foreground(f.profile.Color("1")).String()
	}
	if _, err := fmt.Fprintln(f.errOut, msg); err != nil {
		return
	}
}

// PrintSuccess outputs a success message
func (f *Formatter) PrintSuccess(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if f.profile != termenv.Ascii {
		msg = termenv.String(msg).Foreground(f.profile.Color("2")).String()
	}
	if _, err := fmt.Fprintln(f.out, msg); err != nil {
		return
	}
}

// PrintWarning outputs a warning message
func (f *Formatter) PrintWarning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if f.profile != termenv.Ascii {
		msg = termenv.String(msg).Foreground(f.profile.Color("3")).String()
	}
	if _, err := fmt.Fprintln(f.errOut, msg); err != nil {
		return
	}
}

// PrintDryRun outputs a dry-run preview in the configured format.
func (f *Formatter) PrintDryRun(preview *dryrun.Preview) error {
	if f.IsJSON() {
		return f.PrintJSON(map[string]any{
			"dry_run": true,
			"preview": preview,
		})
	}
	return preview.Write(f.out)
}

// Table represents a text table for output
type Table struct {
	formatter *Formatter
	headers   []string
	rows      [][]string
	widths    []int
}

// NewTable creates a new table
func (f *Formatter) NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		formatter: f,
		headers:   headers,
		widths:    widths,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(values ...string) {
	// Pad with empty strings if needed
	for len(values) < len(t.headers) {
		values = append(values, "")
	}
	// Truncate if too many
	if len(values) > len(t.headers) {
		values = values[:len(t.headers)]
	}
	// Update widths
	for i, v := range values {
		if len(v) > t.widths[i] {
			t.widths[i] = len(v)
		}
	}
	t.rows = append(t.rows, values)
}

// Render outputs the table
func (t *Table) Render() {
	if len(t.rows) == 0 {
		return
	}

	// Print header
	headerLine := t.formatRow(t.headers)
	if t.formatter.profile != termenv.Ascii {
		headerLine = termenv.String(headerLine).Bold().String()
	}
	if _, err := fmt.Fprintln(t.formatter.out, headerLine); err != nil {
		return
	}

	// Print rows
	for _, row := range t.rows {
		if _, err := fmt.Fprintln(t.formatter.out, t.formatRow(row)); err != nil {
			return
		}
	}
}

func (t *Table) formatRow(values []string) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = padRight(v, t.widths[i])
	}
	return strings.Join(parts, "  ")
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// Output writes data in the configured format
func (f *Formatter) Output(textFn func(), jsonData any) error {
	if f.IsJSON() {
		if f.query != "" {
			result, err := filter.Apply(jsonData, f.query)
			if err != nil {
				return err
			}
			return f.PrintJSON(result)
		}
		return f.PrintJSON(jsonData)
	}
	textFn()
	return nil
}

// OutputFiltered writes data with optional JQ filtering from context.
func (f *Formatter) OutputFiltered(ctx context.Context, textFn func(), jsonData any) error {
	if f.IsJSON() {
		query := GetQuery(ctx)
		if query == "" {
			query = f.query
		}
		if query != "" {
			result, err := filter.Apply(jsonData, query)
			if err != nil {
				return err
			}
			return f.PrintJSON(result)
		}
		return f.PrintJSON(jsonData)
	}
	textFn()
	return nil
}
