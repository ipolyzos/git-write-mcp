// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package commit_changes_test

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
	"github.com/ipolyzos/git-write-mcp/plugins/commit_changes"
)

func TestExecute_ValidCommit(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	writeFile(t, root, "note.txt", "hello")

	if out, err := exec.Command("git", "-C", root, "add", "note.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

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
	h := commit_changes.Handler{Pipeline: pipe, Actor: "test"}
	msg := "Add tracked content"
	res, err := h.Execute(ctx, root, commit_changes.Params{Message: msg})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "execute" {
		t.Fatalf("unexpected: %+v", res)
	}
	out, err := exec.Command("git", "-C", root, "log", "-1", "--pretty=%B").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(out))
	want := strings.TrimSpace(msg)
	if got != want {
		t.Fatalf("log message = %q, want %q", got, want)
	}
}

func TestExecute_InvalidOrEmptyMessage(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	writeFile(t, root, "x.txt", "y")
	if out, err := exec.Command("git", "-C", root, "add", "x.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

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
	h := commit_changes.Handler{Pipeline: pipe, Actor: "test"}

	for _, msg := range []string{"", "   ", "\t\n"} {
		_, err := h.Execute(ctx, root, commit_changes.Params{Message: msg})
		if err == nil {
			t.Fatalf("expected error for message %q", msg)
		}
		var ve *policy.ValidationError
		if !errors.As(err, &ve) || ve.Field != "message" {
			t.Fatalf("want validation error for message, got %v", err)
		}
	}

	long := strings.Repeat("x", policy.MaxCommitMessageRunes+1)
	_, err = h.Execute(ctx, root, commit_changes.Params{Message: long})
	if err == nil {
		t.Fatal("expected error for overlong message")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "message" {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestExecute_NoStagedChanges(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	// Untracked file only — nothing staged
	writeFile(t, root, "orphan.txt", "z")

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
	h := commit_changes.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, commit_changes.Params{Message: "Commit unstaged only"})
	if err == nil {
		t.Fatal("expected repo state error")
	}
	if !errors.Is(err, policy.ErrRepoState) {
		t.Fatalf("want ErrRepoState, got %v", err)
	}
	if res.OK {
		t.Fatal("expected ok=false")
	}
}

func TestExecute_ApprovalDenied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	writeFile(t, root, "a.txt", "b")
	if out, err := exec.Command("git", "-C", root, "add", "a.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

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
	h := commit_changes.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, commit_changes.Params{Message: "Denied"})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	writeFile(t, root, "dry.txt", "x")
	if out, err := exec.Command("git", "-C", root, "add", "dry.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	headBefore, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}

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
	h := commit_changes.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, commit_changes.Params{Message: "Should not run", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "dry-run" {
		t.Fatalf("unexpected: %+v", res)
	}
	headAfter, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(headBefore) != string(headAfter) {
		t.Fatal("dry-run must not create a commit")
	}
	out, err := exec.Command("git", "-C", root, "diff", "--cached", "--name-only").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "dry.txt" {
		t.Fatalf("staging should remain: %q", string(out))
	}
}

func TestExecute_GitExecutionFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	hookDir := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		t.Fatal(err)
	}
	preCommit := filepath.Join(hookDir, "pre-commit")
	if err := os.WriteFile(preCommit, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, root, "blocked.txt", "y")
	if out, err := exec.Command("git", "-C", root, "add", "blocked.txt").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

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
	h := commit_changes.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, commit_changes.Params{Message: "Hook blocks this"})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.OK {
		t.Fatal("expected ok=false")
	}
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
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
	if out, err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "Add repository baseline").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
