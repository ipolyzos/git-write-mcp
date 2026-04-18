// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"path/filepath"

	"github.com/ipolyzos/git-write-mcp/internal/pathcheck"
)

// Engine provides repository-scoped validation helpers (paths relative to a single repo root).
type Engine struct {
	repoRoot string
}

// NewEngine returns a policy engine bound to an absolute repository root.
func NewEngine(repoRoot string) (*Engine, error) {
	if repoRoot == "" {
		return nil, &ValidationError{Field: "repo_root", Msg: "must be non-empty"}
	}
	if !filepath.IsAbs(repoRoot) {
		return nil, &ValidationError{Field: "repo_root", Msg: "must be an absolute path"}
	}
	return &Engine{repoRoot: filepath.Clean(repoRoot)}, nil
}

// MustBeContainedUnderRepo delegates to path containment using the engine root.
func (e *Engine) MustBeContainedUnderRepo(candidate string) error {
	if e == nil {
		return &ValidationError{Field: "engine", Msg: "nil engine"}
	}
	return pathcheck.MustBeContainedUnderRoot(e.repoRoot, candidate)
}
