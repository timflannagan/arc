// Package printer handles output formatting for ar commands.
// Supports table, YAML, and JSON output modes.
package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format represents an output format.
type Format string

const (
	FormatTable Format = "table"
	FormatYAML  Format = "yaml"
	FormatJSON  Format = "json"
)

// ParseFormat parses a format string, defaulting to table.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "yaml":
		return FormatYAML
	case "json":
		return FormatJSON
	default:
		return FormatTable
	}
}

// PrintTable writes tabular output with aligned columns.
func PrintTable(w io.Writer, columns []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	// Header
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, strings.ToUpper(col))
	}
	fmt.Fprintln(tw)

	// Rows
	for _, row := range rows {
		for i, val := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, val)
		}
		fmt.Fprintln(tw)
	}

	tw.Flush()
}

// PrintYAML writes data as YAML to the writer.
func PrintYAML(w io.Writer, data any) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return err
	}
	return encoder.Close()
}

// PrintJSON writes data as indented JSON to the writer.
func PrintJSON(w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
