// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
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
	Approver policy.Approver
	Git      GitRunner
	Audit    *audit.Logger
}

// Execute runs the mutation pipeline steps in order:
// 1. request validation
// 2. repository boundary (policy engine)
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

	if p.Policy != nil {
		if err := p.Policy.MustBeContainedUnderRepo(req.WorkDir); err != nil {
			rec.Mode = audit.ModeBlocked
			rec.Outcome = audit.OutcomeFailure
			rec.Detail = err.Error()
			p.emit(ctx, rec)
			return Result{}, err
		}
	}

	if p.Approver == nil {
		rec.Mode = audit.ModeBlocked
		rec.Outcome = audit.OutcomeFailure
		rec.Detail = policy.ErrApprovalRequired.Error()
		p.emit(ctx, rec)
		return Result{}, policy.ErrApprovalRequired
	}

	inv := policy.InvocationRequest{
		WorkDir:         req.WorkDir,
		Args:            req.GitArgs,
		Actor:           req.Actor,
		Action:          req.Action,
		Target:          req.Target,
		DryRunRequested: req.DryRunRequested,
	}
	d, err := p.Approver.Approve(ctx, inv)
	if err != nil {
		rec.Mode = audit.ModeBlocked
		rec.Outcome = audit.OutcomeFailure
		rec.Detail = err.Error()
		p.emit(ctx, rec)
		return Result{}, err
	}
	if !d.Approved {
		rec.Mode = audit.ModeBlocked
		rec.Outcome = audit.OutcomeFailure
		rec.Detail = policy.ErrApprovalRequired.Error()
		p.emit(ctx, rec)
		return Result{}, policy.ErrApprovalRequired
	}

	skipExec := d.DryRun || req.DryRunRequested
	preview := &DryRunPreview{
		GitArgv: append([]string(nil), req.GitArgs...),
		WorkDir: req.WorkDir,
	}

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
