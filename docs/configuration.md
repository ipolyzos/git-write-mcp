# Configuration (stdio server)

The process must know **which repository** to operate on and **how approvals** are decided. All variables are read from the process environment (e.g. your Cursor **`mcp.json`** `env` block, a systemd unit, or a shell wrapper).

---

## Core

| Environment variable | Meaning |
|----------------------|---------|
| `GIT_WRITE_MCP_REPO_ROOT` | **Required** (except **`-version`**). Absolute or relative path to the **Git repository root** the server may mutate. |
| `GIT_WRITE_MCP_AUTO_APPROVE` | If `1`, `true`, `yes`, or `on`, mutations use **`policy.AlwaysApprove`**. Otherwise **`policy.DenyAll`** (default **safe**). |
| `GIT_WRITE_MCP_ACTOR` | Optional label on audit records (default: `mcp`). |

---

## Push

| Environment variable | Meaning |
|----------------------|---------|
| `GIT_WRITE_MCP_ALLOWED_PUSH_REMOTES` | Comma-separated remotes allowed for **`push_branch`** (default: `origin`). Example: `origin,upstream`. |

---

## Pull requests & merge

| Environment variable | Meaning |
|----------------------|---------|
| `GIT_WRITE_MCP_ALLOWED_PR_TARGET_BRANCHES` | Bases allowed for **`create_pull_request`** and the **PR `base_ref`** checked by **`merge_pull_request`** (default: **`develop`**). |
| `GIT_WRITE_MCP_MERGE_ALLOWED_METHODS` | `merge`, `squash`, `rebase` allowlist for **`merge_pull_request`** (default: **`squash`** only). |
| `GIT_WRITE_MCP_MERGE_REQUIRE_GREEN_CHECKS` | Default **`1`**: combined status for head SHA must be **`success`**. Set **`0`** to disable (unsafe). |
| `GIT_WRITE_MCP_MERGE_ALLOW_PENDING_CHECKS` | Default **`0`**: **`pending`** blocks merge. **`1`** allows merging while pending (dangerous). |
| `GIT_WRITE_MCP_MERGE_MIN_APPROVALS` | Minimum **`APPROVED`** reviews from GitHub reviews API (default **`1`**). |
| `GIT_WRITE_MCP_MERGE_BLOCK_ON_CHANGES_REQUESTED` | Default **`1`**: merge blocked if any **`CHANGES_REQUESTED`**. **`0`** ignores. |
| `GIT_WRITE_MCP_MERGE_REQUIRE_CLEAN_MERGEABILITY` | Default **`1`**: **`mergeable_state`** must be **`clean`**. |

---

## Releases

| Environment variable | Meaning |
|----------------------|---------|
| `GIT_WRITE_MCP_RELEASE_SOURCE_BRANCH` | Local **`HEAD`** must equal this branch before **`create_release_from_main`** executes (default **`main`**). **`target_commitish`** uses this ref on GitHub. |

---

## GitHub REST

| Environment variable | Meaning |
|----------------------|---------|
| `GITHUB_TOKEN` | Bearer token (PAT or Actions token). Required for **`create_pull_request`**, **`get_pull_request_status`**, **`merge_pull_request`**, **`create_release_from_main`** when the client initializes at **server start**. |
| `GITHUB_REPOSITORY` | `owner/name`. **Or** set **`GITHUB_REPOSITORY_OWNER`** + **`GITHUB_REPOSITORY_NAME`**. |

If the GitHub client **does not** load (missing token/repo), GitHub-backed tools either return a **configuration error** on execute or register as generic **stub** `"not implemented"` placeholders—see [MCP clients & Cursor setup](mcp-clients.md).

---

## Example launch

```bash
export GIT_WRITE_MCP_REPO_ROOT=/path/to/repo
export GIT_WRITE_MCP_AUTO_APPROVE=1
./bin/git-write-mcp
```

Combine with **`GITHUB_*`** when you need PR or release tools.
