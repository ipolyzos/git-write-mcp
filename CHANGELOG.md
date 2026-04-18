# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Makefile targets for build, test, vet, format check, lint, and bundled `ci`.
- Bootstrap layout: `internal/server`, `internal/runtime`, `internal/git`, `internal/policy`, `internal/audit`, `internal/plugin`, and thin plugins under `plugins/` (`create_branch`, `stage_files`, `commit_changes`, `tag_revision`) registering tool names only (no mutation logic yet).
