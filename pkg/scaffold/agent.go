package scaffold

import "fmt"

// AgentOptions extends Options with agent-specific fields.
type AgentOptions struct {
	Options
	Framework     string // adk (default)
	Language      string // python (default)
	ModelProvider string // gemini, openai, anthropic
	ModelName     string
	Image         string
}

const agentYAML = `apiVersion: ar.dev/v1alpha1
kind: Agent
metadata:
  name: {{ .Name }}
  version: "{{ .Version }}"
spec:
  image: {{ .Image }}
  language: {{ .Language }}
  framework: {{ .Framework }}
  modelProvider: {{ .ModelProvider }}
  modelName: {{ .ModelName }}
  description: "{{ .Description }}"
  status: active
  mcpServers: []
  skills: []
  prompts: []
`

const agentDockerfile = `FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8080

CMD ["python", "main.py"]
`

const agentRequirements = `google-adk>=1.0.0
`

const agentMain = `"""{{ .Name }} - an agent powered by {{ .ModelProvider }}/{{ .ModelName }}."""

# TODO: Implement your agent logic here.
# This is a minimal scaffold to get started.


def main():
    print("Hello from {{ .Name }}!")


if __name__ == "__main__":
    main()
`

// Agent scaffolds a new agent project.
func Agent(opts AgentOptions) error {
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}
	if opts.Framework == "" {
		opts.Framework = "adk"
	}
	if opts.Language == "" {
		opts.Language = "python"
	}
	if opts.ModelProvider == "" {
		opts.ModelProvider = "Gemini"
	}
	if opts.ModelName == "" {
		opts.ModelName = "gemini-2.0-flash"
	}
	if opts.Image == "" {
		opts.Image = fmt.Sprintf("localhost:5001/%s:latest", opts.Name)
	}
	if opts.Description == "" {
		opts.Description = fmt.Sprintf("Agent %s", opts.Name)
	}

	dir := opts.Dir()
	if err := ensureDir(dir); err != nil {
		return err
	}

	files := []struct {
		name string
		tmpl string
	}{
		{"agent.yaml", agentYAML},
		{"Dockerfile", agentDockerfile},
		{"requirements.txt", agentRequirements},
		{"main.py", agentMain},
	}

	for _, f := range files {
		if err := writeTemplate(dir, f.name, f.tmpl, opts); err != nil {
			return err
		}
	}

	return nil
}
