// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/audit"
	"github.com/ipolyzos/git-write-mcp/internal/git"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

func TestPipeline_DryRunDoesNotInvokeGitBinary(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	var gitCalls int
	p := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   mustEngine(t, root),
		Approver: alwaysApproveDryRun{},
		Git: runnerFunc(func(ctx context.Context, workDir string, argv []string, skipExecution bool) (git.RunResult, error) {
			gitCalls++
			if !skipExecution {
				t.Fatal("expected skipExecution true for dry-run pipeline path")
			}
			return git.RunResult{Skipped: true}, nil
		}),
		Audit: audit.NewLogger(audit.SinkFunc(func(context.Context, audit.Record) error { return nil })),
	}
	ctx := context.Background()
	req := &runtime.MutationRequest{
		Actor:           "test",
		Action:          "create_branch",
		Target:          "feature/dry",
		WorkDir:         root,
		GitArgs:         []string{"branch", "feature/dry"},
		DryRunRequested: true,
	}
	res, err := p.Execute(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if res.DryRunPreview == nil {
		t.Fatal("expected dry-run preview")
	}
	if gitCalls != 1 {
		t.Fatalf("git runner calls = %d, want 1 (gated skip)", gitCalls)
	}
}

func TestPipeline_EmitAuditContainsRequiredFields(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	var got audit.Record
	p := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   mustEngine(t, root),
		Approver: alwaysApproveExec{},
		Git:      ex,
		Audit: audit.NewLogger(audit.SinkFunc(func(_ context.Context, r audit.Record) error {
			got = r
			return nil
		})),
	}
	ctx := context.Background()
	req := &runtime.MutationRequest{
		Actor:   "actor1",
		Action:  "commit_changes",
		Target:  "index",
		WorkDir: root,
		GitArgs: []string{"rev-parse", "--is-inside-work-tree"},
	}
	_, err = p.Execute(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Timestamp.IsZero() {
		t.Fatal("missing timestamp")
	}
	if got.Actor != "actor1" {
		t.Fatalf("actor = %q", got.Actor)
	}
	if got.Action != "commit_changes" {
		t.Fatalf("action = %q", got.Action)
	}
	if got.Target != "index" {
		t.Fatalf("target = %q", got.Target)
	}
	if got.Mode != audit.ModeExecute {
		t.Fatalf("mode = %q", got.Mode)
	}
	if got.Outcome != audit.OutcomeSuccess {
		t.Fatalf("outcome = %q", got.Outcome)
	}
}

func TestPipeline_ApproverNilRejected(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	p := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   mustEngine(t, root),
		Approver: nil,
		Git:      ex,
		Audit:    audit.NewLogger(audit.SinkFunc(func(context.Context, audit.Record) error { return nil })),
	}
	ctx := context.Background()
	req := &runtime.MutationRequest{
		Actor:   "a",
		Action:  "create_branch",
		Target:  "feature/nil-appr",
		WorkDir: root,
		GitArgs: []string{"branch", "feature/nil-appr"},
	}
	_, err = p.Execute(ctx, req)
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestPipeline_NilMutationRequestRejected(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	p := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   mustEngine(t, root),
		Approver: alwaysApproveExec{},
		Git: runnerFunc(func(context.Context, string, []string, bool) (git.RunResult, error) {
			t.Fatal("git should not run")
			return git.RunResult{}, nil
		}),
		Audit: audit.NewLogger(audit.SinkFunc(func(context.Context, audit.Record) error { return nil })),
	}
	ctx := context.Background()
	_, err := p.Execute(ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "nil mutation") {
		t.Fatalf("got %v", err)
	}
}

func TestPipeline_RequestValidateFailsBeforeGit(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	var gitCalls int
	p := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   mustEngine(t, root),
		Approver: alwaysApproveExec{},
		Git: runnerFunc(func(context.Context, string, []string, bool) (git.RunResult, error) {
			gitCalls++
			return git.RunResult{}, nil
		}),
		Audit: audit.NewLogger(audit.SinkFunc(func(context.Context, audit.Record) error { return nil })),
	}
	ctx := context.Background()
	req := &runtime.MutationRequest{
		Actor:   "a",
		Action:  "create_branch",
		Target:  "feature/boom",
		WorkDir: root,
		GitArgs: []string{"branch", "feature/boom"},
		Validate: func() error {
			return errors.New("boom")
		},
	}
	_, err := p.Execute(ctx, req)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("got %v", err)
	}
	if gitCalls != 0 {
		t.Fatalf("git calls = %d", gitCalls)
	}
}

type runnerFunc func(context.Context, string, []string, bool) (git.RunResult, error)

func (f runnerFunc) RunApproved(ctx context.Context, wd string, argv []string, skip bool) (git.RunResult, error) {
	return f(ctx, wd, argv, skip)
}

type alwaysApproveDryRun struct{}

func (alwaysApproveDryRun) Approve(context.Context, policy.InvocationRequest) (policy.Decision, error) {
	return policy.Decision{Approved: true, DryRun: true}, nil
}

type alwaysApproveExec struct{}

func (alwaysApproveExec) Approve(context.Context, policy.InvocationRequest) (policy.Decision, error) {
	return policy.Decision{Approved: true, DryRun: false}, nil
}

func mustEngine(t *testing.T, root string) *policy.Engine {
	t.Helper()
	e, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	for _, args := range [][]string{
		{"git", "-C", dir, "config", "user.email", "t@t"},
		{"git", "-C", dir, "config", "user.name", "t"},
	} {
		if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
			t.Fatal(err)
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
