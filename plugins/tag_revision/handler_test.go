// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package tag_revision_test

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
	"github.com/ipolyzos/git-write-mcp/plugins/tag_revision"
)

func TestExecute_ValidAnnotatedTag(t *testing.T) {
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
	h := tag_revision.Handler{Pipeline: pipe, Actor: "test"}

	res, err := h.Execute(ctx, root, tag_revision.Params{
		TagName: "v1.2.3-test",
		Message: "Annotated release marker",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "execute" {
		t.Fatalf("unexpected: %+v", res)
	}

	out, err := exec.Command("git", "-C", root, "rev-parse", "refs/tags/v1.2.3-test^{}").CombinedOutput()
	if err != nil {
		t.Fatalf("missing annotated tag: %v\n%s", err, out)
	}
	ft, err := exec.Command("git", "-C", root, "cat-file", "-t", "refs/tags/v1.2.3-test").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(ft)) != "tag" {
		t.Fatalf("want annotated tag object, got %q", string(ft))
	}
}

func TestExecute_InvalidTagName(t *testing.T) {
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
	h := tag_revision.Handler{Pipeline: pipe, Actor: "test"}

	_, err = h.Execute(ctx, root, tag_revision.Params{
		TagName: "bad tag",
		Message: "Valid body",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "tag name" {
		t.Fatalf("want tag name validation error, got %v", err)
	}
}

func TestExecute_InvalidTargetRef(t *testing.T) {
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
	h := tag_revision.Handler{Pipeline: pipe, Actor: "test"}

	_, err = h.Execute(ctx, root, tag_revision.Params{
		TagName:   "v1.0.0",
		TargetRef: "refs/heads/not valid ref",
		Message:   "Annotation text",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "ref" {
		t.Fatalf("want ref validation error, got %v", err)
	}
}

func TestExecute_EmptyMessage(t *testing.T) {
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
	h := tag_revision.Handler{Pipeline: pipe, Actor: "test"}

	_, err = h.Execute(ctx, root, tag_revision.Params{
		TagName: "v0.0.1",
		Message: "   ",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "message" {
		t.Fatalf("want message validation error, got %v", err)
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
	h := tag_revision.Handler{Pipeline: pipe, Actor: "test"}

	_, err = h.Execute(ctx, root, tag_revision.Params{
		TagName: "v9.9.9",
		Message: "Denied tag",
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
	h := tag_revision.Handler{Pipeline: pipe, Actor: "test"}

	res, err := h.Execute(ctx, root, tag_revision.Params{
		TagName: "v0-dry-run",
		Message: "Should not exist",
		DryRun:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "dry-run" {
		t.Fatalf("unexpected: %+v", res)
	}
	out, err := exec.Command("git", "-C", root, "tag", "-l", "v0-dry-run").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("tag should not exist after dry-run: %q", string(out))
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
	h := tag_revision.Handler{Pipeline: pipe, Actor: "test"}

	res, err := h.Execute(ctx, root, tag_revision.Params{
		TagName:   "v-missing-target",
		TargetRef: "refs/heads/does-not-exist-branch-999",
		Message:   "Points nowhere",
	})
	if err == nil {
		t.Fatal("expected git error")
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
	if out, err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "Add repository baseline").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
