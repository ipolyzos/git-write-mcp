// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/ipolyzos/git-write-mcp/internal/git"

// MutationRequest is input to [Pipeline.Execute] for one mutating tool operation.
// Execute receives a pointer so [MutationRequest.Validate] may set GitArgs and Target
// after validating inputs (see plugins/stage_files).
type MutationRequest struct {
	Actor   string
	Action  string
	Target  string
	WorkDir string
	GitArgs []string

	// DryRunRequested asks for a preview only (no git mutation); combined with [policy.Decision.DryRun].
	DryRunRequested bool

	// Validate is optional request-level validation (tool-specific inputs).
	Validate func() error
}

// Result is the deterministic structured output from [Pipeline.Execute].
type Result struct {
	OK            bool           `json:"ok"`
	DryRunPreview *DryRunPreview `json:"dry_run_preview,omitempty"`
	GitResult     git.RunResult  `json:"git_result,omitempty"`
}
