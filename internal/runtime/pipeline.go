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
// 6. audit record ([audit.Append])
// 7. structured result
func (p *Pipeline) Execute(ctx context.Context, req *MutationRequest) (Result, error) {
	if p == nil {
		return Result{}, errors.New("runtime: nil pipeline")
	}
	if req == nil {
		return Result{}, errors.New("runtime: nil mutation request")
	}

	ts := time.Now().UTC()

	if p.Git == nil {
		meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeBlocked, audit.OutcomeFailure, "nil git runner")
		res := Result{Audit: meta}
		if aerr != nil {
			return res, fmt.Errorf("audit: %w", aerr)
		}
		return res, errors.New("runtime: nil git runner")
	}

	if req.Validate != nil {
		if err := req.Validate(); err != nil {
			meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeBlocked, audit.OutcomeFailure, err.Error())
			res := Result{Audit: meta}
			if aerr != nil {
				return res, fmt.Errorf("audit: %w", aerr)
			}
			return res, err
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
			meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeBlocked, audit.OutcomeFailure, rs.Detail)
			res := Result{Audit: meta}
			if aerr != nil {
				return res, fmt.Errorf("audit: %w", aerr)
			}
			return res, fmt.Errorf("%w: %v", policy.ErrRepoState, rs.Detail)
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
		meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeBlocked, audit.OutcomeFailure, dec.DenyDetail)
		res := Result{Audit: meta}
		if aerr != nil {
			return res, fmt.Errorf("audit: %w", aerr)
		}
		return res, policy.ErrApprovalFromDecision(dec)
	}

	skipExec := dec.DryRun || req.DryRunRequested
	previewBody := BuildDryRunPreview(req)
	preview := &previewBody

	if skipExec {
		gr, err := p.Git.RunApproved(ctx, req.WorkDir, req.GitArgs, true)
		if err != nil {
			meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeBlocked, audit.OutcomeFailure, err.Error())
			res := Result{DryRunPreview: preview, GitResult: gr, Audit: meta}
			if aerr != nil {
				return res, fmt.Errorf("audit: %w", aerr)
			}
			return res, err
		}
		meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeDryRun, audit.OutcomeSuccess, "")
		res := Result{OK: true, DryRunPreview: preview, GitResult: gr, Audit: meta}
		if aerr != nil {
			return res, fmt.Errorf("audit: %w", aerr)
		}
		return res, nil
	}

	gr, err := p.Git.RunApproved(ctx, req.WorkDir, req.GitArgs, false)
	if err != nil {
		meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeExecute, audit.OutcomeFailure, err.Error())
		res := Result{GitResult: gr, Audit: meta}
		if aerr != nil {
			return res, fmt.Errorf("audit: %w", aerr)
		}
		return res, err
	}
	meta, aerr := p.appendAudit(ctx, req, ts, audit.ModeExecute, audit.OutcomeSuccess, "")
	res := Result{OK: true, GitResult: gr, Audit: meta}
	if aerr != nil {
		return res, fmt.Errorf("audit: %w", aerr)
	}
	return res, nil
}

func (p *Pipeline) appendAudit(ctx context.Context, req *MutationRequest, ts time.Time, mode, outcome, detail string) (audit.AppendMetadata, error) {
	return audit.Append(ctx, p.Audit, audit.AppendInput{
		Timestamp:     ts,
		Actor:         req.Actor,
		Action:        req.Action,
		Target:        req.Target,
		Mode:          mode,
		Outcome:       outcome,
		Detail:        detail,
		CorrelationID: req.CorrelationID,
	})
}
