// Package batch provides utilities for batch operations on JSON/NDJSON input.
package batch

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Constants for batch processing limits.
const (
	MaxInputSize = 10 * 1024 * 1024 // 10MB
	MaxItemCount = 10000
)

// Item represents a single item in a batch.
type Item struct {
	raw json.RawMessage
}

// Unmarshal decodes the item into the given value.
func (i Item) Unmarshal(v any) error {
	return json.Unmarshal(i.raw, v)
}

// Raw returns the raw JSON bytes.
func (i Item) Raw() json.RawMessage {
	return i.raw
}

// Result represents the result of processing a single item.
type Result struct {
	Index   int
	Success bool
	Data    any
	Error   error
}

// Summary provides statistics about batch processing.
type Summary struct {
	Total     int
	Succeeded int
	Failed    int
}

// ReadItems reads items from a file (use "-" for stdin).
func ReadItems(filename string) ([]Item, error) {
	var reader io.Reader

	if filename == "-" {
		reader = io.LimitReader(os.Stdin, MaxInputSize+1)
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				return
			}
		}()

		// Check file size
		stat, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}
		if stat.Size() > MaxInputSize {
			return nil, fmt.Errorf("file too large: %d bytes (max %d)", stat.Size(), MaxInputSize)
		}

		reader = f
	}

	return parseJSON(reader)
}

// parseJSON parses JSON array or NDJSON from a reader.
func parseJSON(reader io.Reader) ([]Item, error) {
	// Read all data with size limit
	data, err := io.ReadAll(io.LimitReader(reader, MaxInputSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	if len(data) > MaxInputSize {
		return nil, fmt.Errorf("input too large: exceeds %d bytes", MaxInputSize)
	}

	// Trim whitespace
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}

	// Try parsing as JSON array first
	if data[0] == '[' {
		return parseJSONArray(data)
	}

	// Try parsing as NDJSON
	return parseNDJSON(data)
}

func parseJSONArray(data []byte) ([]Item, error) {
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}

	if len(rawItems) > MaxItemCount {
		return nil, fmt.Errorf("too many items: %d (max %d)", len(rawItems), MaxItemCount)
	}

	items := make([]Item, 0, len(rawItems))
	for _, raw := range rawItems {
		// Validate each item is an object
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) == 0 || trimmed[0] != '{' {
			return nil, fmt.Errorf("expected object, got: %s", string(trimmed[:min(20, len(trimmed))]))
		}
		items = append(items, Item{raw: raw})
	}

	return items, nil
}

func parseNDJSON(data []byte) ([]Item, error) {
	var items []Item
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		if len(items) >= MaxItemCount {
			return nil, fmt.Errorf("too many items: exceeds %d", MaxItemCount)
		}

		// Validate it's a JSON object
		if line[0] != '{' {
			return nil, fmt.Errorf("expected object, got: %s", string(line[:min(20, len(line))]))
		}

		// Validate it's valid JSON
		var obj json.RawMessage
		if err := json.Unmarshal(line, &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON on line %d: %w", len(items)+1, err)
		}

		items = append(items, Item{raw: json.RawMessage(append([]byte(nil), line...))})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan input: %w", err)
	}

	return items, nil
}
