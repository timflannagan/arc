# arc - Declarative CLI for the Agent Registry

kubectl-style CLI that manages agents, MCP servers, skills, and prompts via YAML.

## Build

```bash
make build    # outputs bin/arc
make install  # installs to $GOPATH/bin
```

## Architecture

- **Resource registry pattern** (`pkg/resource/`): Each resource type (Agent, MCPServer, Skill, Prompt) implements `ResourceType` and self-registers via `init()`. Adding a new type = one file.
- **Scheme** (`pkg/scheme/`): Parses YAML with `apiVersion`/`kind` dispatch, supports multi-document files.
- **Client** (`pkg/client/`): Thin HTTP client for the registry v0 API. Intentionally decoupled from the agentregistry module.
- **Config** (`pkg/config/`): kubeconfig-style config at `~/.arc/config`. `set-context` is the single command for adding registries (manages clusters internally).
- **Scaffold** (`pkg/scaffold/`): Project scaffolding templates for each resource type. Used by `arc init`.
- **Commands** (`pkg/cmd/`): Cobra command tree. Local ops: `init`, `build`. Registry ops: `apply`, `get`, `pull`, `delete`, `import`, `export`. Provider ops: `provider list`, `provider add`, `provider delete`, `provider update`, `provider setup`. Config: `config`.
- **Printer** (`pkg/printer/`): Output formatting (table, YAML, JSON) via `-o` flag.

## API Version

Resources use `apiVersion: ar.dev/v1alpha1`. This is defined in `pkg/scheme/scheme.go`.

## Environment Variables

- `ARC_SERVER` - Registry server URL (overrides config)
- `ARC_TOKEN` - Auth token (overrides config)
