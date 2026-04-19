# Safety model & pipeline

All **write** MCP tools share the same behavioral core: **`runtime.Pipeline`** runs:

1. **Request validation** (tool-specific `Validate`)
2. **Repository state checks** via **`policy.ValidateRepoState`** where wired (detached HEAD rules, staged index for commits, current branch for push, etc.)
3. **Centralized approval** — **`policy.EvaluateApproval`** (`internal/policy/approval_check.go`)
4. **Dry-run preview** or **execute** (git subprocess or GitHub API callback)
5. **Audit append** via **`audit.Append`**

Structured responses include `ok`, `mode` (`execute` | `dry-run` | `blocked`), tool fields, optional `dry_run_preview`, and `error`.

---

## Write tools matrix

| Write tool | Approval | Dry-run | Audit | Notes |
|------------|----------|---------|-------|-------|
| `stage_files`, `unstage_files` | yes | yes | yes | Paths under repo root; **unstage** uses allowlisted `git reset HEAD --` only |
| `commit_changes` | yes | yes | yes | Requires **staged** index (`RequireStagedIndex`) |
| `create_branch`, `push_branch`, `tag_revision` | yes | yes | yes | Push must match checked-out branch; remote allowlist |
| `create_pull_request`, `merge_pull_request` | yes | yes | yes | GitHub REST; merge policy gates (checks, reviews, base branch allowlist, method) |
| `create_release_from_main` | yes | yes | yes | Local **`HEAD`** must match **`GIT_WRITE_MCP_RELEASE_SOURCE_BRANCH`** (default **`main`**); **`POST .../releases`** |

**Read tools:** **`get_capabilities`** — no mutation; deterministic JSON. **`get_pull_request_status`** — GitHub reads only; **no** mutation pipeline.

---

## Contract parity

Extend the **approval allowlist** and **`plugin.DocumentCapabilities().WriteTools`** together — **`internal/contracts`** tests fail if capability metadata and policy drift apart.

See also: [MCP tool reference](reference/mcp-tools.md) for per-tool JSON.
