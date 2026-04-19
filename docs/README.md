# git-write-mcp documentation

stdio **Model Context Protocol (MCP)** server for **controlled write-capable Git** (and optional GitHub REST) from agents such as Cursor. Operations use **allowlisted `git` argv only**—never arbitrary shell.

**Upstream:** https://github.com/ipolyzos/git-write-mcp  

---

## Start here

| Doc | Audience |
|-----|----------|
| [Overview](overview.md) | What this server does, scope boundaries |
| [Safety model & pipeline](safety-model.md) | Validation → approval → dry-run / execute → audit |
| [Configuration](configuration.md) | All environment variables |
| [Workflow](workflow.md) | Branch model (`develop` / `main`), PRs, releases |
| [Security](security.md) | Principles and safe defaults |
| [Development](development.md) | Build, test, CI parity, TDD expectations |
| [MCP clients & Cursor setup](mcp-clients.md) | Wiring stdio MCP, approvals, sandbox repos |
| [Repository layout](repository-layout.md) | Package map (`internal/*`, `plugins/*`) |
| [MCP tool reference](reference/mcp-tools.md) | Per-tool JSON inputs, `ToolResult` fields, examples |
| [GitHub integration](github-integration.md) | Token, repository env, Actions, E2E flow |
| [Changelog](../CHANGELOG.md) | Version history (kept at repository root) |

---

## Community & legal (repository root)

GitHub links **License**, **Security policy**, and **Code of conduct** from the repo overview when these root files exist:

| Topic | Root file |
|-------|-----------|
| License | [`LICENSE`](../LICENSE) |
| Security (disclosure) | [`SECURITY.md`](../SECURITY.md) |
| Code of Conduct | [`CODE_OF_CONDUCT.md`](../CODE_OF_CONDUCT.md) |

Product-facing security **behavior** of this MCP server (approvals, tokens, boundaries): [Security principles](security.md).

---

## Capabilities at a glance

- **Write (git):** `create_branch`, `stage_files`, `unstage_files`, `commit_changes`, `push_branch`, `tag_revision`
- **Write (GitHub):** `create_pull_request`, `merge_pull_request`, `create_release_from_main` (require `GITHUB_TOKEN` + repo at **server start**)
- **Read (GitHub):** `get_pull_request_status` (same preconditions)
- **Metadata:** `get_capabilities` (no mutation; lists `write_tools`, `read_tools`, `behaviors`, `excluded_read_tools`)

**Not here:** `git status` / `diff` / `log` / `show` / `blame` / `branch` list—use a read-only Git MCP or the host.

---

## License & conduct

**Apache-2.0** — see [`LICENSE`](../LICENSE).  
Copyright **2026** **Ioannis Polyzos**. **Contact:** git@ipolyzos.com  

See also [`CODE_OF_CONDUCT.md`](../CODE_OF_CONDUCT.md) and [`SECURITY.md`](../SECURITY.md).
