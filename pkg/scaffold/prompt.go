package scaffold

import "fmt"

// PromptOptions extends Options with prompt-specific fields.
type PromptOptions struct {
	Options
	Content string
}

const promptYAML = `apiVersion: ar.dev/v1alpha1
kind: Prompt
metadata:
  name: {{ .Name }}
  version: "{{ .Version }}"
spec:
  description: "{{ .Description }}"
  content: |
    {{ .Content }}
`

// Prompt scaffolds a new prompt resource file.
func Prompt(opts PromptOptions) error {
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}
	if opts.Description == "" {
		opts.Description = fmt.Sprintf("Prompt %s", opts.Name)
	}
	if opts.Content == "" {
		opts.Content = "You are a helpful assistant."
	}

	dir := opts.Dir()
	if err := ensureDir(dir); err != nil {
		return err
	}

	return writeTemplate(dir, "prompt.yaml", promptYAML, opts)
}
