package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
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
	dataOnly  bool
	raw       bool
	agent     bool
	pretty    bool
}

// New creates a new Formatter
func New(out, errOut io.Writer, format Format, colorMode string) *Formatter {
	f := &Formatter{
		out:       out,
		errOut:    errOut,
		format:    format,
		colorMode: colorMode,
		pretty:    true,
	}
	f.profile = f.detectColorProfile()
	return f
}

// SetAgentMode controls agent-optimized behavior (compact JSON, structured errors, etc.).
func (f *Formatter) SetAgentMode(enabled bool) {
	f.agent = enabled
}

// IsAgentMode returns true if agent mode is enabled on the formatter.
func (f *Formatter) IsAgentMode() bool {
	return f.agent
}

// SetPrettyJSON controls whether JSON output is pretty-printed.
func (f *Formatter) SetPrettyJSON(enabled bool) {
	f.pretty = enabled
}

// SetQuery sets an optional JQ-style query for JSON output.
func (f *Formatter) SetQuery(query string) {
	f.query = strings.TrimSpace(query)
}

// SetDataOnly controls whether JSON output should return only the data/items array when present.
func (f *Formatter) SetDataOnly(enabled bool) {
	f.dataOnly = enabled
}

// SetRaw controls whether JSON output should skip the data envelope.
func (f *Formatter) SetRaw(enabled bool) {
	f.raw = enabled
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
	if f.pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(data)
}

// PrintText outputs plain text
func (f *Formatter) PrintText(text string) {
	// In JSON mode, keep stdout clean for machine parsing.
	out := f.out
	if f.IsJSON() {
		out = f.errOut
	}
	if _, err := fmt.Fprintln(out, text); err != nil {
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
	// In JSON mode, keep stdout clean for machine parsing.
	out := f.out
	if f.IsJSON() {
		out = f.errOut
	}
	if _, err := fmt.Fprintln(out, msg); err != nil {
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
		data := jsonData
		queryTarget := jsonData
		raw := f.raw
		if f.dataOnly {
			if extracted, ok := extractData(jsonData); ok {
				data = extracted
				queryTarget = extracted
			}
		} else if !raw {
			data = ensureEnvelope(jsonData)
		}
		if f.query != "" {
			result, err := filter.Apply(queryTarget, f.query)
			if err != nil {
				return err
			}
			return f.PrintJSON(result)
		}
		return f.PrintJSON(data)
	}
	textFn()
	return nil
}

// OutputFiltered writes data with optional JQ filtering from context.
func (f *Formatter) OutputFiltered(ctx context.Context, textFn func(), jsonData any) error {
	if f.IsJSON() {
		origPretty := f.pretty
		if ctx != nil {
			f.pretty = PrettyJSON(ctx)
		}
		defer func() { f.pretty = origPretty }()

		query := GetQuery(ctx)
		if query == "" {
			query = f.query
		}
		dataOnly := f.dataOnly
		if ctx != nil && GetDataOnly(ctx) {
			dataOnly = true
		}
		raw := f.raw
		if ctx != nil && GetRaw(ctx) {
			raw = true
		}

		// JSON Lines output: stream one JSON value per line (compact).
		if ctx != nil && JSONL(ctx) {
			f.pretty = false

			target := jsonData
			if extracted, ok := extractData(jsonData); ok {
				target = extracted
			}

			v := reflect.ValueOf(target)
			for v.Kind() == reflect.Pointer {
				if v.IsNil() {
					break
				}
				v = v.Elem()
			}
			if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
				// Not a list; fall back to normal JSON output (still compact due to pretty=false).
				goto normalJSON
			}

			enc := json.NewEncoder(f.out)
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Interface()
				out := any(item)
				if query != "" {
					result, err := filter.Apply(item, query)
					if err != nil {
						return err
					}
					out = result
				}
				if err := enc.Encode(out); err != nil {
					return err
				}
			}
			return nil
		}

	normalJSON:
		data := jsonData
		queryTarget := jsonData
		if dataOnly {
			if extracted, ok := extractData(jsonData); ok {
				data = extracted
				queryTarget = extracted
			}
		} else if !raw {
			data = ensureEnvelope(jsonData)
		}
		if query != "" {
			result, err := filter.Apply(queryTarget, query)
			if err != nil {
				return err
			}
			return f.PrintJSON(result)
		}

		// Agent mode: normalize success output unless the user is requesting a raw/custom format.
		if ctx != nil && IsAgent(ctx) && query == "" && !dataOnly && !raw {
			return f.PrintJSON(map[string]any{
				"ok":     true,
				"result": data,
			})
		}

		return f.PrintJSON(data)
	}
	textFn()
	return nil
}

func extractData(data any) (any, bool) {
	if data == nil {
		return nil, false
	}
	val := reflect.ValueOf(data)
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil, false
		}
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Map:
		for _, key := range []string{"data", "items"} {
			mv := val.MapIndex(reflect.ValueOf(key))
			if mv.IsValid() {
				return mv.Interface(), true
			}
		}
	case reflect.Struct:
		for _, name := range []string{"Data", "Items"} {
			fv := val.FieldByName(name)
			if fv.IsValid() && fv.CanInterface() {
				return fv.Interface(), true
			}
		}
	}
	return data, false
}

func ensureEnvelope(data any) any {
	if data == nil {
		return map[string]any{"data": nil}
	}
	if _, ok := extractData(data); ok {
		return data
	}
	return map[string]any{"data": data}
}
