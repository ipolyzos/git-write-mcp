# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `internal/runtime.Pipeline`: validate → repo boundary → centralized `policy.Approver` → dry-run preview / gated `git.RunApproved` → structured `audit.Record` → deterministic `runtime.Result`.
- `internal/audit`: structured records (`timestamp`, `actor`, `action`, `target`, `mode`, `outcome`, optional `detail`) with `Sink` / `Logger`.
- `internal/pathcheck` for canonical containment checks under a repository root.
- `internal/git.Executor`: subprocess-only git (`exec.CommandContext`), argv allowlist, default destructive/history-rewrite blocks, repo workdir boundary, typed `git.KindError`, and `policy.Approver` / dry-run decisions for mutating argv classes.
- `internal/policy`: branch/tag/ref validation, `policy.Engine` path checks, `policy.Approver` / `policy.Decision` extension points, `policy.ErrApprovalRequired` when writes lack approval.

- Makefile targets for build, test, vet, format check, lint, and bundled `ci`.
- Bootstrap layout: `internal/server`, `internal/runtime`, `internal/git`, `internal/policy`, `internal/audit`, `internal/plugin`, and thin plugins under `plugins/` (`create_branch`, `stage_files`, `commit_changes`, `tag_revision`) registering tool names only (no mutation logic yet).
- Minimal stdio MCP server using `github.com/modelcontextprotocol/go-sdk`: `NewMCPServer` / `RunStdio` register registered plugin names as MCP tools with stub `not implemented` responses.
