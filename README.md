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
| `create_branch`, `stage_files`, `commit_changes`, `tag_revision` | `git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`, `git_branch_list` |
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

CI and **`make test`** / **`make ci`** run **`go test -race ./...`** (data race detector).

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
| `internal/policy` | Ref/branch/tag/commit-message validation, repo path containment, Approver extension |
| `internal/audit` | Structured audit `Record`, `Sink`, `Logger` |
| `plugins/*` | Tool plugins (`create_branch`, `stage_files`, `commit_changes`, `tag_revision`) |

---

## Configuration (stdio server)

The process must know which repository to operate on and how approvals are decided:

| Environment variable | Meaning |
|----------------------|---------|
| `GIT_WRITE_MCP_REPO_ROOT` | **Required** (except `-version`). Absolute or relative path to the Git repository root. |
| `GIT_WRITE_MCP_AUTO_APPROVE` | If `1`, `true`, `yes`, or `on`, mutations use `policy.AlwaysApprove`. Otherwise `policy.DenyAll` (default safe). |
| `GIT_WRITE_MCP_ACTOR` | Optional label stored on audit records (default: `mcp`). |

Example:

```bash
export GIT_WRITE_MCP_REPO_ROOT=/path/to/repo
export GIT_WRITE_MCP_AUTO_APPROVE=1
./bin/git-write-mcp
```

---

## Tool: `create_branch`

Creates a branch with `git branch <branch_name> [<from_ref>]`. Validation uses `policy.ValidateBranchName` and `policy.ValidateRefName`. Execution goes through the shared `runtime.Pipeline` (approval, dry-run preview, gated git, audit).

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `branch_name` | yes | New branch name. |
| `from_ref` | no | Start point (ref or commit-ish). Omit to branch from current `HEAD`. |
| `dry_run` | no | If `true`, runs validation + preview only (no branch created). |

**Structured output (`ToolResult`):** `ok`, `mode` (`execute` \| `dry-run` \| `blocked`), `branch_name`, `from_ref`, `dry_run_requested`, `dry_run`, `git_stdout`, `git_stderr`, `error`.

Example success (execute):

```json
{
  "ok": true,
  "branch_name": "feature/foo",
  "from_ref": "",
  "dry_run_requested": false,
  "mode": "execute",
  "git_stdout": "",
  "git_stderr": ""
}
```

Example dry-run:

```json
{
  "ok": true,
  "branch_name": "feature/foo",
  "dry_run_requested": true,
  "mode": "dry-run",
  "dry_run": true
}
```

---

## Tool: `stage_files`

Stages paths with `git add -- <path>...`. Each path is trimmed, resolved relative to the repository root or validated if absolute, then checked with `policy.Engine.MustBeContainedUnderRepo` (rejects traversal and paths outside `GIT_WRITE_MCP_REPO_ROOT`). Duplicate paths after normalization are de-duplicated; argv order is sorted for determinism. Execution uses `runtime.Pipeline` (approval, dry-run preview, gated git, audit).

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `paths` | yes | Non-empty array of paths (repo-relative or absolute paths inside the repo). |
| `dry_run` | no | If `true`, validation + preview only (nothing staged). |

**Structured output (`ToolResult`):** `ok`, `mode` (`execute` \| `dry-run` \| `blocked`), `paths` (validated relative paths in sorted order), `dry_run_requested`, `dry_run`, `git_stdout`, `git_stderr`, `error`.

Example success (execute):

```json
{
  "ok": true,
  "paths": ["src/file.go"],
  "dry_run_requested": false,
  "mode": "execute",
  "git_stdout": "",
  "git_stderr": ""
}
```

---

## Tool: `commit_changes`

Creates a commit from **currently staged** changes only with `git commit -m <message>`. There is **no** `--amend` or other history-rewrite argv (also blocked by `internal/git` destructive checks). Message validation uses `policy.ValidateCommitMessage`. Execution goes through `runtime.Pipeline`.

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `message` | yes | Commit message (subject may span lines; validated for length and control characters). |
| `dry_run` | no | If `true`, validation + preview only (no commit created). |

**Structured output (`ToolResult`):** `ok`, `mode` (`execute` \| `dry-run` \| `blocked`), `message`, `dry_run_requested`, `dry_run`, `git_stdout`, `git_stderr`, `error`.

Example success (execute):

```json
{
  "ok": true,
  "message": "Add export helper",
  "dry_run_requested": false,
  "mode": "execute",
  "git_stdout": "",
  "git_stderr": ""
}
```

---

## Tool: `tag_revision`

Creates an **annotated** tag with `git tag -a <tag_name> -m <message> [<target_ref>]`. Omit `target_ref` to tag `HEAD`. Tag names use `policy.ValidateTagName`; annotation body uses `policy.ValidateCommitMessage`; optional `target_ref` uses `policy.ValidateRefName`. argv does **not** include `-f`/`--force` or `-d` (executor also treats forced tag operations as destructive). Runs through `runtime.Pipeline`.

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `tag_name` | yes | Tag identifier (e.g. `v1.0.0`). |
| `target_ref` | no | Commit-ish or ref to tag; omit for `HEAD`. |
| `message` | yes | Annotation message (same rules as commit messages). |
| `dry_run` | no | If `true`, validation + preview only (no tag created). |

**Structured output (`ToolResult`):** `ok`, `mode` (`execute` \| `dry-run` \| `blocked`), `tag_name`, `target_ref`, `message`, `dry_run_requested`, `dry_run`, `git_stdout`, `git_stderr`, `error`.

Example success (execute):

```json
{
  "ok": true,
  "tag_name": "v1.0.0",
  "target_ref": "",
  "message": "Release",
  "dry_run_requested": false,
  "mode": "execute",
  "git_stdout": "",
  "git_stderr": ""
}
```

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
