// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/ipolyzos/git-write-mcp/internal/pathcheck"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

// Executor runs allowlisted git subprocesses without a shell (direct exec only).
type Executor struct {
	repoRoot string
	gitPath  string
	approver policy.Approver
}

// ExecutorOptions configures [NewExecutor].
type ExecutorOptions struct {
	RepoRoot string
	// GitPath is the git binary path; empty uses exec.LookPath("git").
	GitPath string
	// Approver is required for mutating argv classes (see readsWithoutApproval). Nil rejects writes.
	Approver policy.Approver
}

// RunResult is the outcome of [Executor.Run] or [Executor.RunApproved].
type RunResult struct {
	Stdout  []byte
	Stderr  []byte
	Skipped bool // true when dry-run approval skipped subprocess execution
}

// NewExecutor validates repo root and resolves the git binary path.
func NewExecutor(opts ExecutorOptions) (*Executor, error) {
	if opts.RepoRoot == "" {
		return nil, &KindError{Kind: KindInvalidRepoRoot, Msg: "repo root is empty"}
	}
	if !filepath.IsAbs(opts.RepoRoot) {
		return nil, &KindError{Kind: KindInvalidRepoRoot, Msg: "repo root must be absolute"}
	}
	root := filepath.Clean(opts.RepoRoot)
	gitPath := opts.GitPath
	if gitPath == "" {
		var err error
		gitPath, err = exec.LookPath("git")
		if err != nil {
			return nil, &KindError{Kind: KindInvalidRepoRoot, Msg: fmt.Sprintf("resolve git binary: %v", err)}
		}
	}
	return &Executor{repoRoot: root, gitPath: gitPath, approver: opts.Approver}, nil
}

// Run executes git argv (without the literal "git" token) under workDir using separate argv slots (no shell).
func (e *Executor) Run(ctx context.Context, workDir string, argv []string) (RunResult, error) {
	if e == nil {
		return RunResult{}, &KindError{Kind: KindInvalidRepoRoot, Msg: "nil executor"}
	}
	if err := pathcheck.MustBeContainedUnderRoot(e.repoRoot, workDir); err != nil {
		return RunResult{}, &KindError{Kind: KindOutsideRepo, Msg: fmt.Sprintf("workdir outside repo: %v", err)}
	}
	if isDestructive(argv) {
		return RunResult{}, &KindError{Kind: KindDestructiveBlocked, Msg: fmt.Sprintf("destructive invocation blocked: %v", argv)}
	}
	if !allowlisted(argv) {
		return RunResult{}, &KindError{Kind: KindNotAllowed, Msg: fmt.Sprintf("subcommand not allowlisted: %v", argv)}
	}
	if !readsWithoutApproval(argv) {
		if e.approver == nil {
			return RunResult{}, policy.ErrApprovalRequired
		}
		d, err := e.approver.Approve(ctx, policy.InvocationRequest{WorkDir: workDir, Args: argv})
		if err != nil {
			return RunResult{}, err
		}
		if !d.Approved {
			return RunResult{}, policy.ErrApprovalRequired
		}
		if d.DryRun {
			return e.runGated(ctx, workDir, argv, true)
		}
	}
	return e.runGated(ctx, workDir, argv, false)
}

// RunApproved runs safety gates then either skips execution (dry-run / preview) or invokes git.
// Approval must be enforced by the caller (e.g. [runtime.Pipeline]); [Executor.Run] is the convenience path that includes approval.
func (e *Executor) RunApproved(ctx context.Context, workDir string, argv []string, skipExecution bool) (RunResult, error) {
	return e.runGated(ctx, workDir, argv, skipExecution)
}

func (e *Executor) runGated(ctx context.Context, workDir string, argv []string, skipExecution bool) (RunResult, error) {
	if e == nil {
		return RunResult{}, &KindError{Kind: KindInvalidRepoRoot, Msg: "nil executor"}
	}
	if err := pathcheck.MustBeContainedUnderRoot(e.repoRoot, workDir); err != nil {
		return RunResult{}, &KindError{Kind: KindOutsideRepo, Msg: fmt.Sprintf("workdir outside repo: %v", err)}
	}
	if isDestructive(argv) {
		return RunResult{}, &KindError{Kind: KindDestructiveBlocked, Msg: fmt.Sprintf("destructive invocation blocked: %v", argv)}
	}
	if !allowlisted(argv) {
		return RunResult{}, &KindError{Kind: KindNotAllowed, Msg: fmt.Sprintf("subcommand not allowlisted: %v", argv)}
	}
	if skipExecution {
		return RunResult{Skipped: true}, nil
	}

	cmd := exec.CommandContext(ctx, e.gitPath, argv...)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	res := RunResult{Stdout: stdout, Stderr: stderr.Bytes()}
	if err != nil {
		return res, fmt.Errorf("git %v: %w\nstderr: %s", argv, err, stderr.Bytes())
	}
	return res, nil
}

func readsWithoutApproval(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	switch argv[0] {
	case "version", "rev-parse":
		return true
	default:
		return false
	}
}
