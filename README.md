# git-write-mcp

stdio Model Context Protocol (MCP) server that exposes **controlled, write-capable Git operations** for tools such as Cursor on **Linux** and **macOS**. It shells out only to the **`git` CLI** using an explicit **command allowlist**—never arbitrary shell execution.

**Upstream:** https://github.com/ipolyzos/git-write-mcp  
**Contact:** git@ipolyzos.com  

This software is licensed under **Apache-2.0**. Copyright **2026** **Ioannis Polyzos**. See `LICENSE`.

---

## Purpose

Provide MCP tools that can **mutate** a Git repository (branch, index, commits, tags) with **explicit approvals**, **dry-run** support, **structured auditing**, **repository boundary enforcement**, and **strict validation** of branches, refs, tags, paths, and commit messages. Deterministic structured responses are required for all operations.

---

## Scope boundaries

| In scope (this MCP) | Out of scope (use a read Git MCP or the host) |
|---------------------|-----------------------------------------------|
| `create_branch`, `stage_files`, `commit_changes`, `tag_revision` (when implemented) | `git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`, `git_branch_list` |
| Allowlisted `git` invocations; no freestyle shell | Ad-hoc shell, pipe chains, or user-defined commands |
| Policy + audit around each mutation | History rewriting / destructive defaults (must stay off by default) |

Read-oriented queries stay in a dedicated read MCP so this server can focus on **writes** and **policy**.

---

## Branch model

1. **`main`** — minimal, release-ready line; **no direct feature work**.
2. **`develop`** — **only** integration branch; all features merge here via PR.
3. **Feature branches** — short-lived branches cut from **`develop`** (`feature/...`).
4. **`main`** receives **`develop`** only via **release PR** when cutting a release.

---

## Pull request workflow

1. Branch from **`develop`**: `git checkout develop && git pull && git checkout -b feature/your-change`.
2. Prefer **strict TDD** for behavior changes: failing test → minimal implementation → green.
3. Push frequently; open a PR into **`develop`**.
4. Keep CI green; iterate until reviewers and checks pass.
5. Merge to **`develop`** via PR (squash or merge per repo policy—not via direct pushes to **`develop`**).

---

## Release workflow

1. On **`develop`**, update **`CHANGELOG.md`** under `[Unreleased]` / version section as appropriate.
2. Ensure tests, lint, and formatting pass locally and in GitHub Actions.
3. Open a PR **`develop` → `main`** (release PR).
4. After merge to **`main`**, tag and publish releases **from `main`** only.

---

## TDD requirements

For code changes:

1. Define scope.
2. Add **failing tests** first (compile-fail or assertion fail).
3. Implement the **minimum** to pass.
4. Run **`make ci`** or the same steps as CI (`check-fmt`, `vet`, `test`, `lint`, `build`).
5. Update docs when behavior or contracts change.

---

## Security principles

- **Approvals** required for mutating MCP actions (enforced in the mutation pipeline).
- **Dry-run** supported for mutating actions.
- **Structured audit trail** for every mutation.
- **Repository boundary** enforcement (paths and repo root).
- **Strict validation** of refs, tags, branch names, commit messages, and paths.
- **No arbitrary shell execution**; **git command allowlist** only.
- **No privilege escalation**; deterministic structured outputs only.
- **Destructive / history-rewriting** operations **blocked by default**.

---

## Repository layout (bootstrap)

| Path | Role |
|------|------|
| `cmd/git-write-mcp` | Process entrypoint |
| `internal/server` | MCP server wiring (bootstrap registers plugins) |
| `internal/runtime` | Shared mutation `Pipeline`, `MutationRequest` / `Result`, dry-run previews |
| `internal/plugin` | Plugin registry and interface |
| `internal/git` | Allowlisted `Executor` (direct `exec`, no shell) |
| `internal/policy` | Ref/branch/tag validation, repo path containment, Approver extension |
| `internal/audit` | Structured audit `Record`, `Sink`, `Logger` |
| `plugins/*` | Thin tool plugins (names only until tools are implemented) |

---

## Build and test

```bash
make build
make test
make ci
```

---

## Changelog

See `CHANGELOG.md` (Keep a Changelog).
