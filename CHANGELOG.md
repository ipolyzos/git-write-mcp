# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`policy.ValidateRepoState`** with **`RepoStateCheckInput`**, **`RepoStateResult`**, **`RepoValidationCode`**, and **`policy.ErrRepoState`**: deterministic structured checks for policy engine root alignment, repository path existence, workdir containment, optional ref syntax and existence, detached-HEAD policy, and per-action preconditions (e.g. staged index for commits). **`policy.RepoGitProbe`** abstracts read-only git inspection; **`git.NewRepoProbe`** implements it via narrow allowlisted argv.
- **`runtime.Pipeline.RepoProbe`** and **`runtime.MutationRequest`** fields **`FromRef`**, **`TargetRef`**, **`RejectDetachedHEAD`**, **`RequireStagedIndex`** wire repo-state validation after request validation and before approval. Server **`BuildServeConfig`** sets **`RepoProbe: git.NewRepoProbe(ex)`**.
- **`policy.EvaluateApproval`** / **`ApprovalCheckInput`** / **`ApprovalDecision`**: centralized mutation approval with typed deny reasons (`missing_context`, `unsupported_action`, `no_approver`, `rejected`, `approver_error`); **`runtime.Pipeline`** uses this path before git runs.

### Changed

- **`runtime.Pipeline.Execute`**: repository boundary enforcement is folded into **`policy.ValidateRepoState`** (replacing the standalone **`MustBeContainedUnderRepo`** call). Mutations that fail repo-state checks return **`policy.ErrRepoState`** before approval.

## [0.1.0] - 2026-04-18

First release of the git-write-mcp stdio MCP server with four mutation tools,
policy-backed validation, approvals, dry-run, and structured auditing.

### Added

- Integration tests in `internal/server` (`mutation_tools_integration_test.go`): all four mutation tools run against a shared pipeline with audit capture; dry-run and approval-denied audit modes; repo-boundary rejection for `stage_files`.
- MCP tool **`tag_revision`**: annotated tags via `git tag -a -m` only (no force/delete argv), validates tag name, optional target ref, and annotation message; approval, dry-run, audit via `plugins/tag_revision` and `runtime.Pipeline`.
- MCP tool **`commit_changes`**: validates commit messages (`policy.ValidateCommitMessage`), commits staged changes with allowlisted `git commit -m` only (no amend), approval, dry-run, audit via `plugins/commit_changes` and `runtime.Pipeline`.
- MCP tool **`stage_files`**: validates and resolves paths under the repo root (traversal-safe), stages via allowlisted `git add`, requires approval, supports dry-run and audit via `plugins/stage_files` and `runtime.Pipeline`.
- MCP tool **`create_branch`**: validates branch name and optional source ref, requires approval via `policy.Approver`, supports dry-run, emits audit records, runs through `runtime.Pipeline`, returns deterministic structured JSON via `plugins/create_branch`.
- **`GIT_WRITE_MCP_REPO_ROOT`** (required for stdio serve), **`GIT_WRITE_MCP_AUTO_APPROVE`**, **`GIT_WRITE_MCP_ACTOR`** server configuration (`internal/server`).
- `internal/runtime.Pipeline`: validate → repo boundary → centralized `policy.Approver` → dry-run preview / gated `git.RunApproved` → structured `audit.Record` → deterministic `runtime.Result`.
- `internal/audit`: structured records (`timestamp`, `actor`, `action`, `target`, `mode`, `outcome`, optional `detail`) with `Sink` / `Logger`.
- `internal/pathcheck` for canonical containment checks under a repository root.
- `internal/git.Executor`: subprocess-only git (`exec.CommandContext`), argv allowlist, default destructive/history-rewrite blocks, repo workdir boundary, typed `git.KindError`, and `policy.Approver` / dry-run decisions for mutating argv classes.
- `internal/policy`: branch/tag/ref/commit-message validation, `policy.Engine` path checks, `policy.Approver` / `policy.Decision` extension points, `policy.ErrApprovalRequired` when writes lack approval.

- Makefile targets for build, test, vet, format check, lint, and bundled `ci`.
- Bootstrap layout: `internal/server`, `internal/runtime`, `internal/git`, `internal/policy`, `internal/audit`, `internal/plugin`, and plugins under `plugins/` (`create_branch`, `stage_files`, `commit_changes`, `tag_revision`).
- Minimal stdio MCP server using `github.com/modelcontextprotocol/go-sdk`: `NewMCPServer` / `RunStdio` register tools; mutations are active when `GIT_WRITE_MCP_REPO_ROOT` resolves and `ServeConfig` supplies the pipeline (otherwise unmanaged tools reply `not implemented`).

### Changed

- **Breaking (internal API):** `runtime.Pipeline.Execute` now takes `*runtime.MutationRequest` so validation can populate `GitArgs` / `Target` before approval and git (used by `stage_files`).
- GitHub Actions `go test` runs with `-race` (aligned with `make test` / `make ci`).

### Fixed

- `stage_files`: path validation runs inside the mutation pipeline so invalid or out-of-repo paths emit a blocked audit record (previously returned before `Pipeline.Execute`).

[Unreleased]: https://github.com/ipolyzos/git-write-mcp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ipolyzos/git-write-mcp/releases/tag/v0.1.0
