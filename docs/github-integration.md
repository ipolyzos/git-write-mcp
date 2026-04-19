# GitHub integration (`internal/integration/github`)

Foundation types for GitHub REST automation (pull requests, checks, merges, releases).
Operations on `Integration` are stubs returning `ErrNotImplemented` until implemented;
`Client.DoAuthorizedGET` is available for authenticated reads (see package source).

## Configuration

| Variable | Required | Meaning |
|----------|----------|---------|
| `GITHUB_TOKEN` | Yes for API calls | Bearer token (**PAT**, **fine-grained PAT**, or in Actions the `GITHUB_TOKEN` GitHub passes to the job). Never log this value or echo it in error strings. |
| `GITHUB_REPOSITORY_OWNER` | Yes* | Repository owner (organization or user). |
| `GITHUB_REPOSITORY_NAME` | Yes* | Repository name. |
| `GITHUB_REPOSITORY` | Alternative | `owner/name` form (for example `octocat/Hello-World`). GitHub Actions sets this automatically; used when owner/name vars are unset. |

\* `LoadConfigFromEnv` requires `Config.Validate`: owner, name, and token must all be
present after resolution.

`Config.RedactedSummary` is safe for structured logs (`token=<set>` instead of raw material).

## GitHub Actions considerations

- **Secrets:** Prefer `secrets.GITHUB_TOKEN` (job token) or a stored PAT for cross-repo work. Scope tokens to the smallest permission set.
- **Permissions:** Default job token often has `contents: read`. Future workflows that create PRs, merge, or publish releases need explicit permissions, for example:
  - `pull-requests: write` — open/update PRs
  - `contents: write` — merge default branch or create releases (as applicable)
  - `id-token: write` — only if you add OIDC for cloud deploys (not required by this package)
- **Fork PRs:** The automatic `GITHUB_TOKEN` from a fork workflow has restricted rights; automation that mutates the upstream repo usually runs from `pull_request_target` or a privileged environment—design workflows deliberately (least privilege).
- **Pin actions:** Repository policy already pins third-party Actions in `.github/workflows/ci.yml`; keep that pattern for any new automation jobs.

Current CI (`.github/workflows/ci.yml`) stays **`permissions: contents: read`** — sufficient for build/test and aligned with not granting write API until automation jobs exist.

## End-to-end branch and release flow (with this MCP)

Typical automation aligned with the repository **branch model** (see [Workflow](workflow.md)):

1. **Integration on `develop`:** feature work merges via **`create_pull_request`** (base usually **`develop`**) and **`merge_pull_request`** after **`get_pull_request_status`** shows green checks and policy allows the merge.
2. **Promotion to `main`:** humans or automation open a **release PR** **`develop` → `main`** (outside this server’s defaults—allowed bases are configured explicitly).
3. **Cut a GitHub release:** after checkout **on `main`** (or the branch named by **`GIT_WRITE_MCP_RELEASE_SOURCE_BRANCH`**), run **`create_release_from_main`** with **`tag_name`**, **`release_name`**, and notes. The tool refuses if the local branch is not the configured release branch—this enforces **main-only** (or single-source) release provenance together with **`POST .../releases`** **`target_commitish`**.

Never point **`create_release_from_main`** at arbitrary refs from a non-release checkout; **merge to `main` first**, then release from **`main`**.

## Testing

Inject `HTTPDoer` (fakes recording requests) into `NewClient`; no network or real tokens
are required for unit tests.
