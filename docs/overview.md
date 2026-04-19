# Overview

**git-write-mcp** exposes MCP tools that can **mutate** a Git repository (branches, index, commits, tags) and, when configured, perform **selected GitHub REST** operations. Every mutation is designed for **automation safety**:

- **Explicit approvals** (configurable; default deny)
- **Dry-run** previews with structured `dry_run_preview` when supported
- **Structured audit** records
- **Repository boundary** enforcement (paths must stay under `GIT_WRITE_MCP_REPO_ROOT`)
- **Strict validation** of branch names, refs, tags, paths, and commit messages
- **Deterministic JSON** responses (`ok`, `mode`, optional `error`, tool-specific fields)

Optional **read-only GitHub integration** exposes **`get_pull_request_status`** so agents can babysit merges (checks, reviews, mergeability) without invoking a write tool.

---

## Scope boundaries

| In scope (this MCP) | Out of scope (use a read Git MCP or the host) |
|---------------------|-----------------------------------------------|
| `create_branch`, `create_pull_request`, **`merge_pull_request`**, **`create_release_from_main`**, `push_branch`, `stage_files`, `unstage_files`, `commit_changes`, `tag_revision`; read-only **`get_pull_request_status`** (GitHub REST; token + repo env at server start) | `git_status`, `git_diff`, `git_log`, `git_show`, `git_blame`, `git_branch_list` |
| Allowlisted `git` invocations; **no freestyle shell** | Ad-hoc shell, pipes, user-defined commands |
| Policy + audit around each mutation | History rewriting / destructive defaults (must stay **off** by default) |

Repository history and blame belong in a **read Git MCP**. **`get_pull_request_status`** only queries GitHub’s PR / status / review APIs for the configured repository.

---

## Relationship to other docs

- Execution order and approval hooks: [Safety model & pipeline](safety-model.md)
- Env vars: [Configuration](configuration.md)
- Branching and releases: [Workflow](workflow.md)
