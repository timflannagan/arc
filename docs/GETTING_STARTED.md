# Getting Started with `ar`

`ar` is a kubectl-style CLI for managing agentic artifacts — agents, MCP
servers, skills, and prompts — against an agent registry.

## Install

```bash
make build          # produces bin/ar
make install        # installs to $GOPATH/bin/ar
```

## Quick Start

The happy path has three stages: **init**, **build**, **apply**.

### 1. Scaffold a new agent

```bash
ar init agent adk python my-agent \
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
ar build ./my-agent
```

To tag and push to a remote registry:

```bash
ar build ./my-agent --image ghcr.io/myorg/my-agent:v1.0 --push
```

### 3. Publish to the registry

```bash
ar apply -f my-agent/agent.yaml
```

### 4. Verify

```bash
ar get agents                       # list all agents
ar get agent my-agent               # show details
ar get agent my-agent -o yaml       # full YAML output
```

## Resource Types

| Kind | `ar init` | `ar build` | Description |
|------|-----------|------------|-------------|
| `Agent` | `ar init agent FRAMEWORK LANG NAME` | `ar build ./NAME` | An AI agent with model, skills, and MCP servers |
| `MCPServer` | `ar init mcpserver NAME` | `ar build ./NAME` | An MCP server providing tools/resources |
| `Skill` | `ar init skill NAME` | N/A (no container) | A reusable skill definition (SKILL.md) |
| `Prompt` | `ar init prompt NAME` | N/A (no container) | A versioned prompt template |

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
ar apply -f examples/full-stack.yaml
```

See [examples/full-stack.yaml](../examples/full-stack.yaml) for a complete
example with an MCP server, skill, prompt, and agent.

### Individual resource files

```bash
ar apply -f examples/mcpserver.yaml
ar apply -f examples/skill.yaml
ar apply -f examples/prompt.yaml
ar apply -f examples/agent.yaml
```

Or apply multiple files at once:

```bash
ar apply -f examples/mcpserver.yaml -f examples/skill.yaml -f examples/agent.yaml
```

## Full Agent Lifecycle

Here's the end-to-end workflow for building an agent with dependencies:

```bash
# 1. Scaffold the MCP server
ar init mcpserver my-tools --transport stdio

# 2. Scaffold the agent
ar init agent adk python my-agent --model-provider openai --model-name gpt-4o

# 3. Edit my-agent/agent.yaml to reference the MCP server:
#    mcpServers:
#      - type: registry
#        name: my-tools
#        registryServerName: my-tools
#        registryServerVersion: "0.1.0"

# 4. Build both images
ar build ./my-tools
ar build ./my-agent

# 5. Publish the MCP server first (it's a dependency)
ar apply -f my-tools/mcpserver.yaml

# 6. Publish the agent
ar apply -f my-agent/agent.yaml

# 7. Verify everything is registered
ar get mcpservers
ar get agents
ar get agent my-agent -o yaml
```

## Configuration

`ar` uses a kubeconfig-style config file at `~/.ar/config`.

### View current config

```bash
ar config view
```

Default output:

```yaml
current-context: local
clusters:
  - name: local
    server: http://localhost:12121
contexts:
  - name: local
    cluster: local
```

### Add a production registry

```bash
ar config set-cluster prod --server https://registry.example.com
ar config set-context prod --cluster prod --token "${MY_TOKEN}"
ar config use-context prod
```

### Switch between registries

```bash
ar config use-context local       # local dev
ar config use-context prod        # production
```

### Override per-command

```bash
ar get agents --server https://registry.example.com --token "${MY_TOKEN}"
```

### Environment variables

```bash
export AR_SERVER=https://registry.example.com
export AR_TOKEN=my-token
ar get agents    # uses env vars
```

## Output Formats

All `get` commands support `-o` for output format:

```bash
ar get agents              # table (default)
ar get agents -o yaml      # YAML
ar get agents -o json      # JSON
```

## Deleting Resources

```bash
ar delete agent my-agent --version 1.0.0
ar delete mcpserver my-server --version 0.1.0
```

## Shell Completion

```bash
# Bash
ar completion bash > /etc/bash_completion.d/ar

# Zsh (add to ~/.zshrc)
source <(ar completion zsh)

# Or install permanently
ar completion zsh > "${fpath[1]}/_ar"
```
