// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_branch_test

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
	"github.com/ipolyzos/git-write-mcp/plugins/create_branch"
)

func TestExecute_ValidBranchCreation(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   eng,
		Approver: policy.AlwaysApprove{},
		Git:      ex,
		Audit:    audit.NewLogger(nil),
	}
	h := create_branch.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, create_branch.Params{
		BranchName: "feature/valid",
		DryRun:     false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("unexpected result: %+v", res)
	}
	out, err := exec.Command("git", "-C", root, "show-ref", "--verify", "--quiet", "refs/heads/feature/valid").CombinedOutput()
	if err != nil {
		t.Fatalf("branch missing: %v\n%s", err, out)
	}
}

func TestExecute_InvalidBranchName(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   eng,
		Approver: policy.AlwaysApprove{},
		Git:      ex,
		Audit:    audit.NewLogger(nil),
	}
	h := create_branch.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, create_branch.Params{
		BranchName: "../evil",
		DryRun:     false,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) && !strings.Contains(err.Error(), "branch") {
		t.Fatalf("got %v", err)
	}
}

func TestExecute_InvalidSourceRef(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   eng,
		Approver: policy.AlwaysApprove{},
		Git:      ex,
		Audit:    audit.NewLogger(nil),
	}
	h := create_branch.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, create_branch.Params{
		BranchName: "feature/from-missing",
		FromRef:    "refs/heads/does-not-exist-999",
		DryRun:     false,
	})
	if err == nil {
		t.Fatal("expected git error")
	}
	if res.OK {
		t.Fatal("expected failure result")
	}
}

func TestExecute_ApprovalDenied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   eng,
		Approver: policy.DenyAll{},
		Git:      ex,
		Audit:    audit.NewLogger(nil),
	}
	h := create_branch.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, create_branch.Params{
		BranchName: "feature/denied",
		DryRun:     false,
	})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   eng,
		Approver: policy.AlwaysApprove{},
		Git:      ex,
		Audit:    audit.NewLogger(nil),
	}
	h := create_branch.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, create_branch.Params{
		BranchName: "feature/dry",
		DryRun:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || !res.DryRun {
		t.Fatalf("unexpected: %+v", res)
	}
	out, err := exec.Command("git", "-C", root, "rev-parse", "--verify", "refs/heads/feature/dry").CombinedOutput()
	if err == nil {
		t.Fatalf("branch should not exist after dry-run: %s", out)
	}
}

func TestExecute_GitExecutionFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root,
		Policy:   eng,
		Approver: policy.AlwaysApprove{},
		Git:      ex,
		Audit:    audit.NewLogger(nil),
	}
	h := create_branch.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, create_branch.Params{
		BranchName: "feature/gitfail",
		FromRef:    "not-a-valid-git-ref-xyz",
		DryRun:     false,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.OK {
		t.Fatal("expected ok=false")
	}
}

func initGitRepoWithCommit(t *testing.T) string {
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
	if out, err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
