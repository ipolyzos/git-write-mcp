# git-write-mcp

stdio Model Context Protocol (MCP) server that exposes **controlled, write-capable Git operations** for tools such as Cursor on **Linux** and **macOS**. It shells out only to the **`git` CLI** using an explicit **command allowlist**—never arbitrary shell execution.

**Upstream:** https://github.com/ipolyzos/git-write-mcp  
**Contact:** git@ipolyzos.com  

This software is licensed under **Apache-2.0**. Copyright **2026** **Ioannis Polyzos**. See `LICENSE`.

---

## Purpose

Provide MCP tools that can **mutate** a Git repository (branch, index, commits, tags) with **explicit approvals**, **dry-run** support, **structured auditing**, **repository boundary enforcement**, and **strict validation** of branches, refs, tags, paths, and commit messages. Optional **read-only GitHub integration** exposes pull request readiness (`get_pull_request_status`) for babysitting merges. Deterministic structured responses are required for all operations.

---

## Scope boundaries

| In scope (this MCP) | Out of scope (use a read Git MCP or the host) |
|---------------------|-----------------------------------------------|
| `create_branch`, `create_pull_request`, **`merge_pull_request`**, `push_branch`, `stage_files`, `unstage_files`, `commit_changes`, `tag_revision`; read-only **`get_pull_request_status`** (GitHub REST, requires token + repo env) | `git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`, `git_branch_list` |
| Allowlisted `git` invocations; no freestyle shell | Ad-hoc shell, pipe chains, or user-defined commands |
| Policy + audit around each mutation | History rewriting / destructive defaults (must stay off by default) |

Repository file history and blame stay in a dedicated read Git MCP; **`get_pull_request_status`** only queries GitHub’s PR/check/review APIs for the configured repository.

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
| `internal/integration/github` | GitHub REST foundation (`Config`, `Client`, `Integration`); see `docs/github-integration.md` |
| `plugins/*` | Tool plugins (`create_branch`, `create_pull_request`, `get_pull_request_status`, `merge_pull_request`, `push_branch`, `stage_files`, `unstage_files`, `commit_changes`, `tag_revision`) |

---

## Configuration (stdio server)

The process must know which repository to operate on and how approvals are decided:

| Environment variable | Meaning |
|----------------------|---------|
| `GIT_WRITE_MCP_REPO_ROOT` | **Required** (except `-version`). Absolute or relative path to the Git repository root. |
| `GIT_WRITE_MCP_AUTO_APPROVE` | If `1`, `true`, `yes`, or `on`, mutations use `policy.AlwaysApprove`. Otherwise `policy.DenyAll` (default safe). |
| `GIT_WRITE_MCP_ACTOR` | Optional label stored on audit records (default: `mcp`). |
| `GIT_WRITE_MCP_ALLOWED_PUSH_REMOTES` | Comma-separated remote names allowed for **`push_branch`** (default: `origin`). Example: `origin,upstream`. |
| `GIT_WRITE_MCP_ALLOWED_PR_TARGET_BRANCHES` | Comma-separated base branch names allowed for **`create_pull_request`** and the **PR base ref** checked before **`merge_pull_request`** (default: **`develop`**). |
| `GIT_WRITE_MCP_MERGE_ALLOWED_METHODS` | Comma-separated GitHub **`merge_method`** values allowed for **`merge_pull_request`**: `merge`, `squash`, `rebase` (default: **`squash`** only). |
| `GIT_WRITE_MCP_MERGE_REQUIRE_GREEN_CHECKS` | Default **`1`**: combined commit **`checks_summary.state`** must be **`success`** before merge. Set **`0`** to disable (unsafe). |
| `GIT_WRITE_MCP_MERGE_ALLOW_PENDING_CHECKS` | Default **`0`**: **`pending`** checks block merge. Set **`1`** to allow merging while checks are still pending (dangerous). |
| `GIT_WRITE_MCP_MERGE_MIN_APPROVALS` | Minimum count of **`APPROVED`** reviews from the PR reviews API (default **`1`**). |
| `GIT_WRITE_MCP_MERGE_BLOCK_ON_CHANGES_REQUESTED` | Default **`1`**: merge blocked if any review has **`CHANGES_REQUESTED`**. Set **`0`** to ignore. |
| `GIT_WRITE_MCP_MERGE_REQUIRE_CLEAN_MERGEABILITY` | Default **`1`**: GitHub **`mergeable_state`** must be **`clean`**. |
| `GITHUB_TOKEN` | Bearer token for GitHub REST (required for **`create_pull_request`**, **`get_pull_request_status`**, and **`merge_pull_request`** when the GitHub client initializes at server start). |
| `GITHUB_REPOSITORY` | Repository as **`owner/name`** (e.g. **`octocat/Hello-World`**). Alternative: set **`GITHUB_REPOSITORY_OWNER`** and **`GITHUB_REPOSITORY_NAME`**. |

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

## Tool: `create_pull_request`

Opens a pull request on GitHub via the REST API (`POST /repos/{owner}/{repo}/pulls`). The server constructs a minimal GitHub client when **`GITHUB_TOKEN`** and repository env vars resolve at startup; otherwise the tool returns a clear configuration error on execute. Source and target branch names use the same conservative validation as other tools; **`target_branch`** defaults to **`develop`** and must appear in **`GIT_WRITE_MCP_ALLOWED_PR_TARGET_BRANCHES`** (see Configuration). Execution uses **`runtime.Pipeline`** (approval, dry-run preview describing the API call, audit).

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `source_branch` | yes | Head branch name (must already exist on the remote for a successful PR). |
| `target_branch` | no | Base branch (default **`develop`**). |
| `title` | yes | PR title. |
| `body` | no | PR body. |
| `draft` | no | If `true`, create a draft PR. |
| `dry_run` | no | If `true`, validation + preview only (no GitHub API call). |

**Structured output (`ToolResult`):** `ok`, `mode` (`execute` \| `dry-run` \| `blocked`), `source_branch`, `target_branch`, `title`, `pull_request_number`, `html_url`, `draft`, `dry_run_requested`, `dry_run`, `dry_run_preview`, `error`.

Example success (execute):

```json
{
  "ok": true,
  "source_branch": "feature/foo",
  "target_branch": "develop",
  "title": "Add feature",
  "pull_request_number": 42,
  "html_url": "https://github.com/org/repo/pull/42",
  "dry_run_requested": false,
  "mode": "execute"
}
```

---

## Tool: `get_pull_request_status`

Read-only inspection of an existing pull request in the configured GitHub repository. Calls **`GET /repos/{owner}/{repo}/pulls/{number}`**, **`GET .../commits/{head_sha}/status`**, and **`GET .../pulls/{number}/reviews`**. Does **not** use `runtime.Pipeline`, approvals, or git mutation. Registered only when **`GIT_WRITE_MCP_REPO_ROOT`** is set **and** the GitHub client loads (**`GITHUB_TOKEN`** + repository env).

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `pull_request_number` | yes | Numeric PR number in the configured repo. |

**Structured output (`ToolResult`):** `ok`, `error`, `pull_request` (object with `state`, `merged`, `draft`, `mergeable`, `mergeable_state`, `title`, `html_url`, `head_ref`, `base_ref`, `head_sha`, `checks_summary`, `reviews_summary`).

Example success:

```json
{
  "ok": true,
  "pull_request": {
    "number": 24,
    "title": "Feature",
    "state": "open",
    "merged": false,
    "mergeable_state": "clean",
    "head_ref": "feature/x",
    "base_ref": "develop",
    "head_sha": "abc123",
    "checks_summary": {
      "state": "success",
      "total_count": 2,
      "contexts": [{"context": "ci", "state": "success"}]
    },
    "reviews_summary": {
      "approved": 1,
      "changes_requested": 0,
      "commented": 0,
      "pending": 0,
      "dismissed": 0
    }
  }
}
```

---

## Tool: `merge_pull_request`

Merges an open pull request via **`PUT /repos/{owner}/{repo}/pulls/{number}/merge`** only after:

1. **Repository policy:** the PR’s **`base_ref`** is in **`GIT_WRITE_MCP_ALLOWED_PR_TARGET_BRANCHES`** (same list as **`create_pull_request`**).
2. **Merge method:** **`merge_method`** is allowlisted (**`GIT_WRITE_MCP_MERGE_ALLOWED_METHODS`**); omitted or empty defaults to **`squash`**.
3. **Checks:** combined status for the PR head SHA must be **`success`** unless **`GIT_WRITE_MCP_MERGE_ALLOW_PENDING_CHECKS`** allows **`pending`** (see Configuration).
4. **Reviews:** at least **`GIT_WRITE_MCP_MERGE_MIN_APPROVALS`** approving reviews; **`GIT_WRITE_MCP_MERGE_BLOCK_ON_CHANGES_REQUESTED`** blocks when **`changes_requested`** counts are greater than zero.
5. **Mergeability:** **`mergeable_state`** must be **`clean`** when **`GIT_WRITE_MCP_MERGE_REQUIRE_CLEAN_MERGEABILITY`** is on (default).

Uses **`runtime.Pipeline`** (approval, dry-run preview of the PUT, audit). Draft or already-merged PRs are rejected before calling GitHub. Registered only when **`GIT_WRITE_MCP_REPO_ROOT`** is set **and** the GitHub client loads (**`GITHUB_TOKEN`** + repo env).

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `pull_request_number` | yes | Numeric PR number in the configured repo. |
| `merge_method` | no | `merge`, `squash`, or `rebase` (default **`squash`** after policy normalization). |
| `dry_run` | no | If `true`, fetch PR status + policy validation + preview only (no merge API call). |

**Structured output (`ToolResult`):** `ok`, `mode`, `pull_request_number`, `base_ref`, `merge_method`, `merged_sha`, `merge_message`, `dry_run_requested`, `dry_run_preview`, `error`.

Example success (execute):

```json
{
  "ok": true,
  "pull_request_number": 42,
  "base_ref": "develop",
  "merge_method": "squash",
  "merged_sha": "abc123def456",
  "merge_message": "Pull Request successfully merged",
  "dry_run_requested": false,
  "mode": "execute"
}
```

---

## Tool: `push_branch`

Pushes the **current** branch to an **allowlisted** remote using only `git push <remote> <branch>`: **no** `--force`, **no** refspecs (`:`), and **no** extra arguments. `policy.ValidateBranchName` applies to `branch_name`; the current `HEAD` must match that branch. Default `remote` is `origin` if omitted. Allowlist comes from `GIT_WRITE_MCP_ALLOWED_PUSH_REMOTES` (see Configuration). Uses `runtime.Pipeline` (approval, dry-run preview, gated git, audit).

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `branch_name` | yes | Branch to push; must be the checked-out branch. |
| `remote` | no | Remote name (default `origin`); must appear in the allowlist. |
| `dry_run` | no | If `true`, validation + preview only (nothing pushed). |

**Structured output (`ToolResult`):** `ok`, `mode` (`execute` \| `dry-run` \| `blocked`), `branch_name`, `remote`, `dry_run_requested`, `dry_run`, `dry_run_preview`, `git_stdout`, `git_stderr`, `error`.

Example success (execute):

```json
{
  "ok": true,
  "branch_name": "feature/foo",
  "remote": "origin",
  "dry_run_requested": false,
  "mode": "execute",
  "git_stdout": "",
  "git_stderr": ""
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

## Tool: `unstage_files`

Removes paths from the index with `git reset HEAD -- <path>...` (working tree unchanged). Path validation matches `stage_files`: trimmed, resolved under the repository root, `policy.Engine.MustBeContainedUnderRepo` (rejects traversal and paths outside `GIT_WRITE_MCP_REPO_ROOT`), de-duplicated, sorted argv order. Uses `runtime.Pipeline` (approval, dry-run preview, gated git, audit).

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `paths` | yes | Non-empty array of paths (repo-relative or absolute paths inside the repo). |
| `dry_run` | no | If `true`, validation + preview only (index unchanged). |

**Structured output (`ToolResult`):** `ok`, `mode` (`execute` \| `dry-run` \| `blocked`), `paths` (validated relative paths, sorted), `dry_run_requested`, `dry_run`, `dry_run_preview` (when dry-run), `git_stdout`, `git_stderr`, `error`.

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
