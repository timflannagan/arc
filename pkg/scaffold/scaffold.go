// Package scaffold generates project files for new resources.
// Each resource type has a scaffold function that writes a directory
// containing the resource YAML and any supporting files (Dockerfile, source).
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Options holds common scaffolding parameters.
type Options struct {
	Name        string
	Version     string
	Description string
	OutputDir   string // defaults to ./<Name>
}

// Dir returns the resolved output directory.
func (o *Options) Dir() string {
	if o.OutputDir != "" {
		return o.OutputDir
	}
	return o.Name
}

// writeFile creates a file inside the scaffold directory.
func writeFile(dir, name, content string) error {
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// writeTemplate renders a Go template and writes the result to a file.
func writeTemplate(dir, name, tmpl string, data any) error {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", name, err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template %s: %w", name, err)
	}

	return writeFile(dir, name, buf.String())
}

// ensureDir creates the scaffold directory if it doesn't exist.
// Returns an error if the directory already exists and is non-empty.
func ensureDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err == nil && len(entries) > 0 {
		return fmt.Errorf("directory %s already exists and is not empty", dir)
	}
	return os.MkdirAll(dir, 0o755)
}
