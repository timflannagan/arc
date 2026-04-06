# `arc`

`arc` is a kubectl-style CLI for publishing and inspecting registry resources
against an AgentRegistry instance.

It works with four first-class resource types:

- `Agent`
- `MCPServer`
- `Skill`
- `Prompt`

Resources are authored as YAML, mapped into the registry REST API, and can be
retrieved back into YAML for inspection or reuse.

## Status

This is a lightweight client for the registry API. It is intentionally thinner
than the main `arctl` CLI and focuses on declarative resource workflows.

## Build

Build locally:

```bash
make build
```

Install to `$GOPATH/bin`:

```bash
make install
```

Show available Make targets:

```bash
make help
```

## Registry Prerequisite

You need a running registry before `arc` can publish or fetch resources.

For the fastest local path, start the upstream daemon with `arctl`:

```bash
curl -fsSL https://raw.githubusercontent.com/agentregistry-dev/agentregistry/main/scripts/get-arctl | bash
arctl daemon start
curl -s http://localhost:12121/v0/ping
```

More setup options, including Helm and running from source, are documented in
[GETTING_STARTED.md](docs/GETTING_STARTED.md).

## Quick Start

Scaffold an agent:

```bash
arc init agent adk python my-agent \
  --model-provider openai \
  --model-name gpt-4o
```

Build its container image:

```bash
arc build ./my-agent
```

Publish it to the registry:

```bash
arc apply -f my-agent/agent.yaml
```

Verify it:

```bash
arc api ping
arc api version
arc get agents
arc get agent my-agent -o yaml
```

## Common Commands

```bash
# Apply one or more resource files
arc apply -f examples/mcpserver.yaml -f examples/agent.yaml

# List resources
arc get agents
arc get mcpservers

# Pull a resource back to local YAML
arc pull agent my-agent

# Delete a specific version
arc delete agent my-agent --version 1.0.0

# Inspect non-resource API endpoints
arc api version
arc api jwks
arc api get /v0/providers
```

## Project Layout

```text
cmd/arc/      CLI entrypoint
pkg/cmd/      Cobra commands
pkg/client/   HTTP client for the registry API
pkg/resource/ Kind-specific YAML <-> REST mappings
pkg/scheme/   YAML decoding and validation
pkg/scaffold/ Project scaffolding for new resources
examples/     Sample resource manifests
docs/         Longer-form documentation
```

## Docs

- Getting started: [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)
- Example manifests: [examples](examples)
