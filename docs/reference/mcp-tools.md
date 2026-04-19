# MCP tool reference — JSON inputs and responses

This page documents every `ToolResult` shape and the `get_capabilities` contract. The live tool list is always the `get_capabilities` response; `internal/contracts` tests keep it aligned with the approval allowlist.

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

## Tool: `create_release_from_main`

Creates a **GitHub release** via **`POST /repos/{owner}/{repo}/releases`** with **`target_commitish`** set to the configured release branch (default **`main`**). The tool **rejects** execution unless the **current local branch** matches **`GIT_WRITE_MCP_RELEASE_SOURCE_BRANCH`** (provenance gate—checkout **`main`** before cutting the release from your workstation).

Requires **`runtime.Pipeline`** (approval, dry-run preview, audit) and a working GitHub client at server start (**`GITHUB_TOKEN`** + repository env). At least one of **`release_notes`** or **`changelog_section`** must be non-empty; if both are set, **`release_notes`** wins for the release body.

**Inputs (JSON):**

| Field | Required | Description |
|-------|----------|-------------|
| `tag_name` | yes | Git tag for the release (must pass tag name policy). |
| `release_name` | yes | GitHub release title / name. |
| `release_notes` | no* | Body text for the release. |
| `changelog_section` | no* | Alternative body source (used when `release_notes` is empty). |
| `dry_run` | no | If `true`, validation + preview only (no GitHub API call). |

\* One of **`release_notes`** or **`changelog_section`** is required.

**Structured output (`ToolResult`):** `ok`, `error`, `tag_name`, `release_name`, `release_id`, `html_url`, `release_source_ref`, `dry_run_requested`, `mode`, `dry_run`, `dry_run_preview`, `git_stdout`, `git_stderr`.

Example success (execute):

```json
{
  "ok": true,
  "tag_name": "v1.2.0",
  "release_name": "v1.2.0",
  "release_id": 12345678,
  "html_url": "https://github.com/org/repo/releases/tag/v1.2.0",
  "release_source_ref": "main",
  "dry_run_requested": false,
  "mode": "execute"
}
```

Example dry-run:

```json
{
  "ok": true,
  "tag_name": "v1.2.0",
  "release_name": "v1.2.0",
  "release_source_ref": "main",
  "dry_run_requested": true,
  "mode": "dry-run",
  "dry_run": true,
  "dry_run_preview": {
    "action": "create_release_from_main",
    "targets": ["v1.2.0"],
    "secondary_refs": ["main"],
    "git_argv": ["create_release_from_main", "v1.2.0", "v1.2.0"],
    "git_commands": ["GitHub API POST /repos/org/repo/releases tag=\"v1.2.0\" name=\"v1.2.0\" target_commitish=main"],
    "work_dir": "/abs/path/to/repo",
    "effect_summary": "GitHub API POST /repos/org/repo/releases tag=\"v1.2.0\" name=\"v1.2.0\" target_commitish=main"
  }
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
