# Getting Started with `arc`

`arc` is a kubectl-style CLI for managing agentic artifacts — agents, MCP
servers, skills, and prompts — against an agent registry.

## Prerequisites

You need a running agent registry instance. Two options:

### Option A: Local (Docker)

The quickest path. Install `arctl` and start the daemon:

```bash
# Install arctl
curl -fsSL https://raw.githubusercontent.com/agentregistry-dev/agentregistry/main/scripts/get-arctl | bash

# Start the registry (runs PostgreSQL + registry server via Docker Compose)
arctl daemon start

# Verify it's running
curl -s http://localhost:12121/ping
```

This starts:
- **Registry API/UI** on `http://localhost:12121`
- **PostgreSQL** (pgvector) on port 5432
- **MCP server** on port 31313

Stop when done:

```bash
arctl daemon stop          # stop containers (preserves data)
arctl daemon stop --purge  # stop and delete all data
```

### Option B: Kubernetes (Helm)

```bash
# Quickstart with bundled PostgreSQL
helm install agentregistry oci://ghcr.io/agentregistry-dev/agentregistry/charts/agentregistry \
  --set config.jwtPrivateKey=$(openssl rand -hex 32)

# Port-forward to access locally
kubectl port-forward svc/agentregistry 12121:12121
```

For production with an external database:

```bash
helm install agentregistry oci://ghcr.io/agentregistry-dev/agentregistry/charts/agentregistry \
  --set config.jwtPrivateKey=$(openssl rand -hex 32) \
  --set database.postgres.bundled.enabled=false \
  --set database.postgres.url="postgres://user:pass@host:5432/dbname?sslmode=require"
```

## Install `arc`

```bash
make build          # produces bin/arc
make install        # installs to $GOPATH/bin/arc
```

## Quick Start

The happy path has three stages: **init**, **build**, **apply**.

### 1. Scaffold a new agent

```bash
arc init agent adk python my-agent \
  --model-provider openai \
  --model-name gpt-4o \
  --description "My first agent"
```

This creates a project directory:

```
my-agent/
├── agent.yaml         # declarative resource definition
├── Dockerfile         # container image
├── main.py            # agent source code
└── requirements.txt   # python dependencies
```

### 2. Build the container image

```bash
arc build ./my-agent
```

To tag and push to a remote registry:

```bash
arc build ./my-agent --image ghcr.io/myorg/my-agent:v1.0 --push
```

### 3. Publish to the registry

```bash
arc apply -f my-agent/agent.yaml
```

### 4. Verify

```bash
arc get agents                       # list all agents
arc get agent my-agent               # show details
arc get agent my-agent -o yaml       # full YAML output
```

## Resource Types

| Kind | `arc init` | `arc build` | Description |
|------|-----------|------------|-------------|
| `Agent` | `arc init agent FRAMEWORK LANG NAME` | `arc build ./NAME` | An AI agent with model, skills, and MCP servers |
| `MCPServer` | `arc init mcpserver NAME` | `arc build ./NAME` | An MCP server providing tools/resources |
| `Skill` | `arc init skill NAME` | N/A (no container) | A reusable skill definition (SKILL.md) |
| `Prompt` | `arc init prompt NAME` | N/A (no container) | A versioned prompt template |

## Working with YAML Resources

Every resource follows the same structure:

```yaml
apiVersion: ar.dev/v1alpha1
kind: Agent                  # Agent | MCPServer | Skill | Prompt
metadata:
  name: my-resource
  version: "1.0.0"
spec:
  # kind-specific fields
```

### Multi-resource files

Define an agent and its dependencies in a single file, separated by `---`:

```bash
arc apply -f examples/full-stack.yaml
```

See [examples/full-stack.yaml](../examples/full-stack.yaml) for a complete
example with an MCP server, skill, prompt, and agent.

### Individual resource files

```bash
arc apply -f examples/mcpserver.yaml
arc apply -f examples/skill.yaml
arc apply -f examples/prompt.yaml
arc apply -f examples/agent.yaml
```

Or apply multiple files at once:

```bash
arc apply -f examples/mcpserver.yaml -f examples/skill.yaml -f examples/agent.yaml
```

## Full Agent Lifecycle

Here's the end-to-end workflow for building an agent with dependencies:

```bash
# 1. Scaffold the MCP server
arc init mcpserver my-tools --transport stdio

# 2. Scaffold the agent
arc init agent adk python my-agent --model-provider openai --model-name gpt-4o

# 3. Edit my-agent/agent.yaml to reference the MCP server:
#    mcpServers:
#      - type: registry
#        name: my-tools
#        registryServerName: my-tools
#        registryServerVersion: "0.1.0"

# 4. Build both images
arc build ./my-tools
arc build ./my-agent

# 5. Publish the MCP server first (it's a dependency)
arc apply -f my-tools/mcpserver.yaml

# 6. Publish the agent
arc apply -f my-agent/agent.yaml

# 7. Verify everything is registered
arc get mcpservers
arc get agents
arc get agent my-agent -o yaml
```

## Configuration

`arc` uses a kubeconfig-style config file at `~/.arc/config` to manage
multiple registry instances. This is how you switch between a local
development registry and a remote (e.g. Kubernetes-hosted) one.

### Default config

On first run, `arc` creates a default config pointing at localhost:

```yaml
current-context: local
clusters:
  - name: local
    server: http://localhost:12121
contexts:
  - name: local
    cluster: local
```

View it with:

```bash
arc config view
```

### Adding a remote registry

Say you have a registry deployed on Kubernetes at
`https://registry.internal.example.com`:

```bash
# Register the cluster
arc config set-cluster kube --server https://registry.internal.example.com

# Create a context with auth
arc config set-context kube --cluster kube --token "${MY_TOKEN}"
```

Your config now has both:

```yaml
current-context: local
clusters:
  - name: local
    server: http://localhost:12121
  - name: kube
    server: https://registry.internal.example.com
contexts:
  - name: local
    cluster: local
  - name: kube
    cluster: kube
    token: <your-token>
```

### Switching between local and remote

```bash
arc config use-context kube     # point at remote Kubernetes registry
arc get agents                  # lists agents on remote

arc config use-context local    # back to localhost
arc get agents                  # lists agents locally
```

This also affects `apply`, `pull`, `import`, `export`, and `delete` —
all registry operations go to whichever context is active.

### Override per-command

Skip context switching for one-off commands:

```bash
arc get agents --server https://registry.internal.example.com --token "${MY_TOKEN}"
```

### Environment variables

```bash
export ARC_SERVER=https://registry.internal.example.com
export ARC_TOKEN=my-token
arc get agents    # uses env vars, ignores config context
```

Environment variables take precedence over the config file.

## Output Formats

All `get` commands support `-o` for output format:

```bash
arc get agents              # table (default)
arc get agents -o yaml      # YAML
arc get agents -o json      # JSON
```

## Pulling Resources

Fetch a resource from the registry as a local YAML file (inverse of `apply`):

```bash
arc pull agent my-agent                    # writes ./my-agent/agent.yaml
arc pull mcpserver acme/fetch              # writes ./fetch/mcpserver.yaml
arc pull agent my-agent --version 1.0.0    # specific version
arc pull agent my-agent -d ./staging/      # custom output directory
```

## Import / Export

Bulk operations for migrating or seeding registries:

```bash
# Export everything to a file
arc export -f catalog.yaml

# Export specific types
arc export agents mcpservers -f partial.yaml

# Export to stdout (pipe-friendly)
arc export agents | arc import -f /dev/stdin --server https://other-registry.example.com

# Import into a registry
arc import -f catalog.yaml
```

## Deleting Resources

```bash
arc delete agent my-agent --version 1.0.0
arc delete mcpserver my-server --version 0.1.0
```

## Shell Completion

```bash
# Bash
arc completion bash > /etc/bash_completion.d/arc

# Zsh (add to ~/.zshrc)
source <(arc completion zsh)

# Or install permanently
arc completion zsh > "${fpath[1]}/_arc"
```
