// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package push_branch_test

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
	"github.com/ipolyzos/git-write-mcp/plugins/push_branch"
)

type failGitRunner struct{}

func (failGitRunner) RunApproved(ctx context.Context, workDir string, argv []string, skipExecution bool) (git.RunResult, error) {
	_ = ctx
	_ = workDir
	_ = argv
	_ = skipExecution
	return git.RunResult{}, errors.New("git: simulated push failure")
}

func TestExecute_ValidPushToAllowedRemote(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPushRemotes, "origin")
	bare := initBareRemote(t)
	root := initWorkerRepoWithBranch(t, bare, "feature/push-ok")

	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := push_branch.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, push_branch.Params{
		BranchName: "feature/push-ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "execute" {
		t.Fatalf("unexpected: %+v", res)
	}
	out, err := exec.Command("git", "--git-dir", bare, "show-ref", "--verify", "refs/heads/feature/push-ok").CombinedOutput()
	if err != nil {
		t.Fatalf("remote missing branch: %v\n%s", err, out)
	}
}

func TestExecute_DeniedPushToDisallowedRemote(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPushRemotes, "upstream")
	bare := initBareRemote(t)
	root := initWorkerRepoWithBranch(t, bare, "feature/x")

	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := push_branch.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, push_branch.Params{
		BranchName: "feature/x",
		Remote:     "origin",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "remote" {
		t.Fatalf("got %v", err)
	}
}

func TestExecute_InvalidBranchName(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPushRemotes, "origin")
	bare := initBareRemote(t)
	root := initWorkerRepoWithBranch(t, bare, "feature/ok")

	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := push_branch.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, push_branch.Params{
		BranchName: "..",
		Remote:     "origin",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecute_ApprovalDenied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPushRemotes, "origin")
	bare := initBareRemote(t)
	root := initWorkerRepoWithBranch(t, bare, "feature/deny")

	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.DenyAll{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := push_branch.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, push_branch.Params{
		BranchName: "feature/deny",
		Remote:     "origin",
	})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPushRemotes, "origin")
	bare := initBareRemote(t)
	root := initWorkerRepoWithBranch(t, bare, "feature/dry")

	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := push_branch.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, push_branch.Params{
		BranchName: "feature/dry",
		Remote:     "origin",
		DryRun:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "dry-run" {
		t.Fatalf("unexpected: %+v", res)
	}
	if res.DryRunPreview == nil || res.DryRunPreview.Action != "push_branch" {
		t.Fatalf("preview: %+v", res.DryRunPreview)
	}
	out, err := exec.Command("git", "--git-dir", bare, "show-ref", "--verify", "refs/heads/feature/dry").CombinedOutput()
	if err == nil {
		t.Fatalf("dry-run must not create remote branch, got ref: %s", out)
	}
}

func TestExecute_GitExecutionFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPushRemotes, "origin")
	bare := initBareRemote(t)
	root := initWorkerRepoWithBranch(t, bare, "feature/fail")

	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       failGitRunner{},
		Audit:     audit.NewLogger(nil),
	}
	h := push_branch.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, push_branch.Params{
		BranchName: "feature/fail",
		Remote:     "origin",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.OK {
		t.Fatal("expected ok=false")
	}
}

func initBareRemote(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "remote.git")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "init", "--bare", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func initWorkerRepoWithBranch(t *testing.T, bareRemote, branch string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "work")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "config", "user.email", "t@t"},
		{"git", "-C", dir, "config", "user.name", "t"},
	} {
		if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
			t.Fatal(err)
		}
	}
	if out, err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "Add repository baseline").CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}
	if out, err := exec.Command("git", "-C", dir, "remote", "add", "origin", bareRemote).CombinedOutput(); err != nil {
		t.Fatalf("remote add: %v\n%s", err, out)
	}
	if out, err := exec.Command("git", "-C", dir, "checkout", "-b", branch).CombinedOutput(); err != nil {
		t.Fatalf("checkout -b: %v\n%s", err, out)
	}
	if out, err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "Add feature branch tip").CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	cur, err := exec.Command("git", "-C", abs, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(cur)) != branch {
		t.Fatalf("precondition branch: want %q got %q", branch, cur)
	}
	return abs
}
