# Workflow

This repository follows a **trunk-style** integration model aligned with how **`create_pull_request`** and **`merge_pull_request`** default their bases (`develop`).

---

## Branch model

1. **`main`** — release-ready line; **no direct feature commits**.
2. **`develop`** — **integration branch**; feature work merges here via PR.
3. **`feature/...`** — short-lived branches cut from **`develop`**.
4. **`main`** receives **`develop`** through a **release PR** when shipping a version.

---

## Feature development (into `develop`)

1. Sync: `git checkout develop && git pull`.
2. Branch: `git checkout -b feature/your-change`.
3. Prefer **strict TDD** for behavior changes: failing test → minimal fix → green.
4. Push; open a **PR into `develop`**.
5. Keep **CI green**; address review and automation feedback.
6. Merge via **GitHub PR** (squash or merge per policy)—**not** by pushing directly to **`develop`**.

Optional: use **git-write-mcp** tools from an agent (`create_branch`, `stage_files`, `commit_changes`, `push_branch`, `create_pull_request`, etc.) with approvals and dry-run where appropriate.

---

## Release train (`develop` → `main`)

1. On **`develop`**, update **[CHANGELOG.md](../CHANGELOG.md)** under **`[Unreleased]`** / version section.
2. Run **`make ci`** (or equivalent: format check, vet, race+shuffle tests, lint, build) locally; fix failures before opening the release PR.
3. Open a PR **`develop` → `main`** (release PR).
4. After merge to **`main`**, tag and publish from **`main`** only. The MCP tool **`create_release_from_main`** enforces a **local checkout** on **`GIT_WRITE_MCP_RELEASE_SOURCE_BRANCH`** (default **`main`**) before **`POST /repos/{owner}/{repo}/releases`**.

See [GitHub integration](github-integration.md) for token scope and Actions permissions when automating releases.
