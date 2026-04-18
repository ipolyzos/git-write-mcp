// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/ipolyzos/git-write-mcp/internal/audit"
	"github.com/ipolyzos/git-write-mcp/internal/git"
)

// MutationRequest is input to [Pipeline.Execute] for one mutating tool operation.
// Execute receives a pointer so [MutationRequest.Validate] may set GitArgs and Target
// after validating inputs (see plugins/stage_files).
type MutationRequest struct {
	Actor   string
	Action  string
	Target  string
	WorkDir string
	GitArgs []string

	// FromRef and TargetRef are optional refs for [policy.ValidateRepoState] (syntax + existence checks).
	FromRef   string
	TargetRef string

	// RejectDetachedHEAD rejects mutations when HEAD is detached (tools that assume a branch).
	RejectDetachedHEAD bool
	// RequireStagedIndex rejects commit mutations when the index has no staged changes.
	RequireStagedIndex bool
	// RequireHEADBranch rejects push_branch when the current branch does not match.
	RequireHEADBranch string
	// PushRemote is the validated remote name for push_branch previews and argv.
	PushRemote string

	// APIExecute runs a non-git side effect after approval (e.g. GitHub REST).
	// When non-nil, dry-run skips git; execute calls this instead of git.
	APIExecute func(context.Context) error

	// PRSourceBranch, PRTargetBranch, PRTitle, PRDraft support create_pull_request previews.
	PRSourceBranch string
	PRTargetBranch string
	PRTitle        string
	PRDraft        bool
	GitHubOwner    string
	GitHubRepo     string

	// DryRunRequested asks for a preview only (no git mutation); combined with [policy.Decision.DryRun].
	DryRunRequested bool

	// CorrelationID optional trace key stored on audit records ([audit.Append]).
	CorrelationID string

	// Validate is optional request-level validation (tool-specific inputs).
	Validate func() error
}

// Result is the deterministic structured output from [Pipeline.Execute].
type Result struct {
	OK            bool                 `json:"ok"`
	DryRunPreview *DryRunPreview       `json:"dry_run_preview,omitempty"`
	GitResult     git.RunResult        `json:"git_result,omitempty"`
	Audit         audit.AppendMetadata `json:"audit,omitempty"`
}
