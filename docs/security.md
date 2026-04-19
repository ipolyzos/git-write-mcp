# Security principles

- **Approvals** required for mutating MCP actions (enforced in **`runtime.Pipeline`** via **`policy.EvaluateApproval`**).
- **Dry-run** supported for mutating actions where the tool exposes it (`dry_run` in JSON).
- **Structured audit trail** for every mutation (**`audit.Record`** via **`audit.Append`**).
- **Repository boundary** enforcement (paths and **`GIT_WRITE_MCP_REPO_ROOT`**).
- **Strict validation** of refs, tags, branch names, commit messages, and paths.
- **No arbitrary shell execution**; **`internal/git.Executor`** uses an explicit **argv allowlist**.
- **No privilege escalation**; outputs are deterministic structured JSON only.
- **Destructive / history-rewriting** operations **blocked by default** (including unsafe `git` argv classes).

Operational guidance:

- Prefer **`GIT_WRITE_MCP_AUTO_APPROVE`** **off** in shared or production-facing setups; pair human review with agent runs.
- Scope **`GITHUB_TOKEN`** to least privilege (fine-grained PAT or Actions `GITHUB_TOKEN` with explicit `permissions:`).
- Never echo raw tokens into logs or `ToolResult` strings; GitHub config exposes **`RedactedSummary`** for safe diagnostics.
