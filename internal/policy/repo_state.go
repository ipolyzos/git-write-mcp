// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// RepoGitProbe abstracts read-only git inspection for repository state validation.
// Implemented in [github.com/ipolyzos/git-write-mcp/internal/git] without importing
// transport or MCP layers.
type RepoGitProbe interface {
	GitRepositoryOK(ctx context.Context, workDir string) error
	IsDetachedHEAD(ctx context.Context, workDir string) (detached bool, err error)
	RefExists(ctx context.Context, workDir, ref string) (exists bool, err error)
	HasStagedChanges(ctx context.Context, workDir string) (has bool, err error)
	// CurrentBranchName returns the short branch name for HEAD or an error if detached.
	CurrentBranchName(ctx context.Context, workDir string) (string, error)
}

// RepoValidationCode categorizes repo state failures (empty when OK).
type RepoValidationCode string

const (
	// RepoValidationOK means validation succeeded.
	RepoValidationOK           RepoValidationCode = ""
	RepoValidationNilEngine    RepoValidationCode = "nil_engine"
	RepoValidationRootMismatch RepoValidationCode = "repo_root_mismatch"
	RepoValidationMissing      RepoValidationCode = "repo_missing"
	RepoValidationBoundary     RepoValidationCode = "boundary_violation"
	RepoValidationInvalidRef   RepoValidationCode = "invalid_ref"
	RepoValidationRefNotFound  RepoValidationCode = "ref_not_found"
	RepoValidationNotGit       RepoValidationCode = "not_git_repository"
	RepoValidationDetached     RepoValidationCode = "detached_head"
	RepoValidationPrecondition RepoValidationCode = "precondition_failed"
	RepoValidationProbe        RepoValidationCode = "probe_error"
)

// RepoStateCheckInput gathers optional refs and precondition flags per mutation tool.
type RepoStateCheckInput struct {
	RepoRoot string
	WorkDir  string
	Action   string

	FromRef            string
	TargetRef          string
	RejectDetachedHEAD bool
	RequireStagedIndex bool
	// RequireHEADBranch, if non-empty, requires the current branch to match exactly.
	RequireHEADBranch string
}

// RepoStateResult is deterministic structured output from [ValidateRepoState].
type RepoStateResult struct {
	OK     bool               `json:"ok"`
	Code   RepoValidationCode `json:"code,omitempty"`
	Detail string             `json:"detail,omitempty"`
}

// ValidateRepoState checks filesystem, repository boundary, optional ref syntax and
// existence (via probe), detached HEAD policy, and per-action preconditions such as
// staged index for commits. Transport and MCP layers stay out of this function.
func ValidateRepoState(ctx context.Context, eng *Engine, probe RepoGitProbe, in RepoStateCheckInput) RepoStateResult {
	if eng == nil {
		return RepoStateResult{OK: false, Code: RepoValidationNilEngine, Detail: "policy: nil engine"}
	}
	root := filepath.Clean(in.RepoRoot)
	if root != eng.repoRoot {
		return RepoStateResult{
			OK:     false,
			Code:   RepoValidationRootMismatch,
			Detail: "policy: RepoStateCheckInput.RepoRoot must match policy engine root",
		}
	}

	fi, err := os.Stat(root)
	if err != nil || !fi.IsDir() {
		return RepoStateResult{
			OK:     false,
			Code:   RepoValidationMissing,
			Detail: "policy: repository root does not exist or is not a directory",
		}
	}

	if err := eng.MustBeContainedUnderRepo(in.WorkDir); err != nil {
		return RepoStateResult{OK: false, Code: RepoValidationBoundary, Detail: err.Error()}
	}

	if strings.TrimSpace(in.FromRef) != "" {
		if err := ValidateRefName(in.FromRef); err != nil {
			return refSyntaxFailure(err)
		}
	}
	if strings.TrimSpace(in.TargetRef) != "" {
		if err := ValidateRefName(in.TargetRef); err != nil {
			return refSyntaxFailure(err)
		}
	}

	if probe == nil {
		return RepoStateResult{OK: true, Code: RepoValidationOK}
	}

	if err := probe.GitRepositoryOK(ctx, in.WorkDir); err != nil {
		return RepoStateResult{OK: false, Code: RepoValidationNotGit, Detail: err.Error()}
	}

	if strings.TrimSpace(in.FromRef) != "" {
		ok, err := probe.RefExists(ctx, in.WorkDir, in.FromRef)
		if err != nil {
			return RepoStateResult{OK: false, Code: RepoValidationProbe, Detail: err.Error()}
		}
		if !ok {
			return RepoStateResult{
				OK:     false,
				Code:   RepoValidationRefNotFound,
				Detail: "policy: from_ref does not resolve to a commit",
			}
		}
	}
	if strings.TrimSpace(in.TargetRef) != "" {
		ok, err := probe.RefExists(ctx, in.WorkDir, in.TargetRef)
		if err != nil {
			return RepoStateResult{OK: false, Code: RepoValidationProbe, Detail: err.Error()}
		}
		if !ok {
			return RepoStateResult{
				OK:     false,
				Code:   RepoValidationRefNotFound,
				Detail: "policy: target_ref does not resolve to a commit",
			}
		}
	}

	if in.RejectDetachedHEAD {
		detached, err := probe.IsDetachedHEAD(ctx, in.WorkDir)
		if err != nil {
			return RepoStateResult{OK: false, Code: RepoValidationProbe, Detail: err.Error()}
		}
		if detached {
			return RepoStateResult{
				OK:     false,
				Code:   RepoValidationDetached,
				Detail: "policy: detached HEAD not allowed for this mutation",
			}
		}
	}

	if in.RequireStagedIndex {
		has, err := probe.HasStagedChanges(ctx, in.WorkDir)
		if err != nil {
			return RepoStateResult{OK: false, Code: RepoValidationProbe, Detail: err.Error()}
		}
		if !has {
			return RepoStateResult{
				OK:     false,
				Code:   RepoValidationPrecondition,
				Detail: "policy: nothing staged for commit",
			}
		}
	}

	if want := strings.TrimSpace(in.RequireHEADBranch); want != "" {
		if probe == nil {
			return RepoStateResult{
				OK:     false,
				Code:   RepoValidationProbe,
				Detail: "policy: nil probe cannot verify current branch",
			}
		}
		got, err := probe.CurrentBranchName(ctx, in.WorkDir)
		if err != nil {
			return RepoStateResult{OK: false, Code: RepoValidationProbe, Detail: err.Error()}
		}
		if got != want {
			return RepoStateResult{
				OK:     false,
				Code:   RepoValidationPrecondition,
				Detail: "policy: current branch " + strconv.Quote(got) + " does not match requested branch " + strconv.Quote(want),
			}
		}
	}

	return RepoStateResult{OK: true, Code: RepoValidationOK}
}

func refSyntaxFailure(err error) RepoStateResult {
	var ve *ValidationError
	if errors.As(err, &ve) && ve != nil {
		return RepoStateResult{
			OK:     false,
			Code:   RepoValidationInvalidRef,
			Detail: ve.Error(),
		}
	}
	return RepoStateResult{
		OK:     false,
		Code:   RepoValidationInvalidRef,
		Detail: err.Error(),
	}
}

// ErrRepoState indicates [ValidateRepoState] rejected the mutation (compare with [errors.Is]).
var ErrRepoState = errors.New("policy: repository state validation failed")
