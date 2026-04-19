# Repository layout

| Path | Role |
|------|------|
| `cmd/git-write-mcp` | Process entrypoint (`-version`, stdio MCP when env is set) |
| `internal/server` | MCP bootstrap, **`NewMCPServer`**, **`RunStdio`**, **`BuildServeConfig`** |
| `internal/runtime` | **`Pipeline`**, **`MutationRequest`** / **`Result`**, dry-run previews |
| `internal/plugin` | Plugin registry, **`DocumentCapabilities`**, **`CapabilitiesJSON`** |
| `internal/git` | Allowlisted **`Executor`** (direct `exec`, no shell) |
| `internal/policy` | Validation, **`ValidateRepoState`**, merge policy, approvers |
| `internal/audit` | **`Record`**, **`Sink`**, **`Logger`**, **`Append`** |
| `internal/contracts` | Tests that align capabilities with policy/mutation actions |
| `internal/integration/github` | GitHub REST foundation; see [github-integration.md](github-integration.md) |
| `plugins/*` | One package per MCP tool |

Tests and CI live under **`.github/workflows/`** and **`Makefile`** targets (`build`, `test`, `vet`, `check-fmt`, `lint`, `ci`).
