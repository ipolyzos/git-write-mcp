// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ipolyzos/git-write-mcp/internal/audit"
	"github.com/ipolyzos/git-write-mcp/internal/git"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

// GitRunner executes gated git argv after pipeline approval (typically [*git.Executor]).
type GitRunner interface {
	RunApproved(ctx context.Context, workDir string, argv []string, skipExecution bool) (git.RunResult, error)
}

// Pipeline coordinates validation, approval, dry-run preview, gated git execution, and audit emission.
type Pipeline struct {
	RepoRoot string
	Policy   *policy.Engine
	// RepoProbe supplies read-only git inspection for [policy.ValidateRepoState]. Nil skips git probes.
	RepoProbe policy.RepoGitProbe
	Approver  policy.Approver
	Git       GitRunner
	Audit     *audit.Logger
}

// Execute runs the mutation pipeline steps in order:
// 1. request validation
// 2. repository state ([policy.ValidateRepoState])
// 3. centralized approval
// 4. dry-run preview + gated skip (no mutation)
// 5. git execution
// 6. audit record
// 7. structured result
func (p *Pipeline) Execute(ctx context.Context, req *MutationRequest) (Result, error) {
	if p == nil {
		return Result{}, errors.New("runtime: nil pipeline")
	}
	if req == nil {
		return Result{}, errors.New("runtime: nil mutation request")
	}

	ts := time.Now().UTC()
	rec := audit.Record{
		Timestamp: ts,
		Actor:     req.Actor,
		Action:    req.Action,
		Target:    req.Target,
	}

	if p.Git == nil {
		rec.Mode = audit.ModeBlocked
		rec.Outcome = audit.OutcomeFailure
		rec.Detail = "nil git runner"
		p.emit(ctx, rec)
		return Result{}, errors.New("runtime: nil git runner")
	}

	if req.Validate != nil {
		if err := req.Validate(); err != nil {
			rec.Mode = audit.ModeBlocked
			rec.Outcome = audit.OutcomeFailure
			rec.Detail = err.Error()
			p.emit(ctx, rec)
			return Result{}, err
		}
	}

	repoRoot := strings.TrimSpace(p.RepoRoot)
	if p.Policy != nil && repoRoot == "" {
		repoRoot = p.Policy.RepoRoot()
	}
	if p.Policy != nil {
		rs := policy.ValidateRepoState(ctx, p.Policy, p.RepoProbe, policy.RepoStateCheckInput{
			RepoRoot:           repoRoot,
			WorkDir:            req.WorkDir,
			Action:             req.Action,
			FromRef:            req.FromRef,
			TargetRef:          req.TargetRef,
			RejectDetachedHEAD: req.RejectDetachedHEAD,
			RequireStagedIndex: req.RequireStagedIndex,
		})
		if !rs.OK {
			rec.Mode = audit.ModeBlocked
			rec.Outcome = audit.OutcomeFailure
			rec.Detail = rs.Detail
			p.emit(ctx, rec)
			return Result{}, fmt.Errorf("%w: %v", policy.ErrRepoState, rs.Detail)
		}
	}

	dec := policy.EvaluateApproval(ctx, p.Approver, policy.ApprovalCheckInput{
		Actor:           req.Actor,
		Action:          req.Action,
		Target:          req.Target,
		WorkDir:         req.WorkDir,
		GitArgs:         req.GitArgs,
		DryRunRequested: req.DryRunRequested,
	})
	if !dec.Allowed {
		rec.Mode = audit.ModeBlocked
		rec.Outcome = audit.OutcomeFailure
		rec.Detail = dec.DenyDetail
		p.emit(ctx, rec)
		return Result{}, policy.ErrApprovalFromDecision(dec)
	}

	skipExec := dec.DryRun || req.DryRunRequested
	previewBody := BuildDryRunPreview(req)
	preview := &previewBody

	if skipExec {
		gr, err := p.Git.RunApproved(ctx, req.WorkDir, req.GitArgs, true)
		if err != nil {
			rec.Mode = audit.ModeBlocked
			rec.Outcome = audit.OutcomeFailure
			rec.Detail = err.Error()
			p.emit(ctx, rec)
			return Result{DryRunPreview: preview, GitResult: gr}, err
		}
		rec.Mode = audit.ModeDryRun
		rec.Outcome = audit.OutcomeSuccess
		p.emit(ctx, rec)
		return Result{OK: true, DryRunPreview: preview, GitResult: gr}, nil
	}

	rec.Mode = audit.ModeExecute
	gr, err := p.Git.RunApproved(ctx, req.WorkDir, req.GitArgs, false)
	if err != nil {
		rec.Outcome = audit.OutcomeFailure
		rec.Detail = err.Error()
		p.emit(ctx, rec)
		return Result{GitResult: gr}, err
	}
	rec.Outcome = audit.OutcomeSuccess
	p.emit(ctx, rec)
	return Result{OK: true, GitResult: gr}, nil
}

func (p *Pipeline) emit(ctx context.Context, rec audit.Record) {
	if p == nil || p.Audit == nil {
		return
	}
	_ = p.Audit.Emit(ctx, rec)
}
