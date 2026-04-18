// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/git"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func TestNewExecutor_RejectsNonAbsoluteRepoRoot(t *testing.T) {
	t.Helper()
	_, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: "relative/repo"})
	if err == nil {
		t.Fatal("expected error for non-absolute repo root")
	}
}

func TestRun_DisallowsNonAllowlistedSubcommand(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = ex.Run(ctx, root, []string{"log", "-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ke *git.KindError
	if !errors.As(err, &ke) || ke.Kind != git.KindNotAllowed {
		t.Fatalf("got %v, want KindNotAllowed", err)
	}
}

func TestRun_AllowsRevParseInsideRepo(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	res, err := ex.Run(ctx, root, []string{"rev-parse", "--is-inside-work-tree"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped {
		t.Fatal("unexpected dry-run skip")
	}
	if !strings.Contains(string(res.Stdout), "true") {
		t.Fatalf("stdout = %q", res.Stdout)
	}
}

func TestRun_BlocksDestructiveResetHard(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = ex.Run(ctx, root, []string{"reset", "--hard", "HEAD"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ke *git.KindError
	if !errors.As(err, &ke) || ke.Kind != git.KindDestructiveBlocked {
		t.Fatalf("got %v, want KindDestructiveBlocked", err)
	}
}

func TestRun_WriteBlockedWhenApproverNil(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{
		RepoRoot: root,
		Approver: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = ex.Run(ctx, root, []string{"add", "README.md"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("got %v, want ErrApprovalRequired", err)
	}
}

type alwaysApprove struct{}

func (alwaysApprove) Approve(context.Context, policy.InvocationRequest) (policy.Decision, error) {
	return policy.Decision{Approved: true, DryRun: false}, nil
}

func TestRun_AllowsWriteWhenApproverApproves(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	readme := filepath.Join(root, "README.md")
	if err := os.WriteFile(readme, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ex, err := git.NewExecutor(git.ExecutorOptions{
		RepoRoot: root,
		Approver: alwaysApprove{},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	res, err := ex.Run(ctx, root, []string{"add", "README.md"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped {
		t.Fatal("unexpected skip")
	}
}

type approveDryRun struct{}

func (approveDryRun) Approve(context.Context, policy.InvocationRequest) (policy.Decision, error) {
	return policy.Decision{Approved: true, DryRun: true}, nil
}

func TestRun_DryRunSkipsGitInvocation(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{
		RepoRoot: root,
		Approver: approveDryRun{},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	res, err := ex.Run(ctx, root, []string{"add", "README.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped {
		t.Fatal("expected dry-run skip")
	}
	if len(res.Stdout) != 0 || len(res.Stderr) != 0 {
		t.Fatalf("unexpected output stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
}

func TestRun_RejectsWorkDirOutsideRepo(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := git.NewExecutor(git.ExecutorOptions{
		RepoRoot: root,
		Approver: alwaysApprove{},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err = ex.Run(ctx, outside, []string{"rev-parse", "--is-inside-work-tree"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ke *git.KindError
	if !errors.As(err, &ke) || ke.Kind != git.KindOutsideRepo {
		t.Fatalf("got %v, want KindOutsideRepo", err)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", dir, "init")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	cfg := exec.Command("git", "-C", dir, "config", "user.email", "t@t")
	if err := cfg.Run(); err != nil {
		t.Fatal(err)
	}
	cfg2 := exec.Command("git", "-C", dir, "config", "user.name", "t")
	if err := cfg2.Run(); err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
