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

## Testing

Inject `HTTPDoer` (fakes recording requests) into `NewClient`; no network or real tokens
are required for unit tests.
