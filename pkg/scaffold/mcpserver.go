package scaffold

import "fmt"

// MCPServerOptions extends Options with MCP server-specific fields.
type MCPServerOptions struct {
	Options
	Framework string // fastmcp-python (default)
	Transport string // stdio, streamable-http
	Image     string
}

const mcpserverYAML = `apiVersion: ar.dev/v1alpha1
kind: MCPServer
metadata:
  name: {{ .Name }}
  version: "{{ .Version }}"
spec:
  title: {{ .Name }}
  description: "{{ .Description }}"
  status: active
  packages:
    - registryType: oci
      identifier: {{ .Image }}
      version: "{{ .Version }}"
      runTimeHint: python
      transport:
        type: {{ .Transport }}
`

const mcpDockerfile = `FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "server.py"]
`

const mcpRequirements = `fastmcp>=2.0.0
`

const mcpServer = `"""{{ .Name }} - MCP server."""

from fastmcp import FastMCP

mcp = FastMCP("{{ .Name }}")


@mcp.tool()
def hello(name: str) -> str:
    """Say hello to someone."""
    return f"Hello, {name}!"


if __name__ == "__main__":
    mcp.run(transport="{{ .Transport }}")
`

// MCPServer scaffolds a new MCP server project.
func MCPServer(opts MCPServerOptions) error {
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}
	if opts.Framework == "" {
		opts.Framework = "fastmcp-python"
	}
	if opts.Transport == "" {
		opts.Transport = "stdio"
	}
	if opts.Image == "" {
		opts.Image = fmt.Sprintf("localhost:5001/%s:latest", opts.Name)
	}
	if opts.Description == "" {
		opts.Description = fmt.Sprintf("MCP server %s", opts.Name)
	}

	dir := opts.Dir()
	if err := ensureDir(dir); err != nil {
		return err
	}

	files := []struct {
		name string
		tmpl string
	}{
		{"mcpserver.yaml", mcpserverYAML},
		{"Dockerfile", mcpDockerfile},
		{"requirements.txt", mcpRequirements},
		{"server.py", mcpServer},
	}

	for _, f := range files {
		if err := writeTemplate(dir, f.name, f.tmpl, opts); err != nil {
			return err
		}
	}

	return nil
}
