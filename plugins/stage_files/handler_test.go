// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package stage_files_test

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
	"github.com/ipolyzos/git-write-mcp/plugins/stage_files"
)

func TestExecute_ValidStaging(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	writeFile(t, root, "tracked.txt", "hello")

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
	h := stage_files.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, stage_files.Params{
		Paths:  []string{"tracked.txt"},
		DryRun: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "execute" {
		t.Fatalf("unexpected: %+v", res)
	}
	out, err := exec.Command("git", "-C", root, "diff", "--cached", "--name-only").CombinedOutput()
	if err != nil {
		t.Fatalf("git diff: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "tracked.txt" {
		t.Fatalf("want tracked.txt staged, got %q", string(out))
	}
}

func TestExecute_EmptyPaths(t *testing.T) {
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
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := stage_files.Handler{Pipeline: pipe, Actor: "test"}
	for _, paths := range [][]string{nil, {}} {
		_, err := h.Execute(ctx, root, stage_files.Params{Paths: paths})
		if err == nil {
			t.Fatal("expected error for empty paths")
		}
		var ve *policy.ValidationError
		if !errors.As(err, &ve) || ve.Field != "paths" {
			t.Fatalf("want paths validation error, got %v", err)
		}
	}
}

func TestExecute_InvalidPath(t *testing.T) {
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
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := stage_files.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, stage_files.Params{Paths: []string{"ok.txt", ""}})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "paths" {
		t.Fatalf("want validation error for empty element, got %v", err)
	}
}

func TestExecute_TraversalAttempt(t *testing.T) {
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
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := stage_files.Handler{Pipeline: pipe, Actor: "test"}
	outside := filepath.Join(filepath.Dir(root), "outside-secret.txt")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(outside) })

	// cwd = repo root; ../file escapes repository boundary
	_, err = h.Execute(ctx, root, stage_files.Params{Paths: []string{"../" + filepath.Base(outside)}})
	if err == nil {
		t.Fatal("expected containment error")
	}
}

func TestExecute_ApprovalDenied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	writeFile(t, root, "x.txt", "y")
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
	h := stage_files.Handler{Pipeline: pipe, Actor: "test"}
	_, err = h.Execute(ctx, root, stage_files.Params{Paths: []string{"x.txt"}})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepoWithCommit(t)
	writeFile(t, root, "dry.txt", "z")
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
	h := stage_files.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, stage_files.Params{Paths: []string{"dry.txt"}, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "dry-run" {
		t.Fatalf("unexpected: %+v", res)
	}
	if res.DryRunPreview == nil || res.DryRunPreview.Action != "stage_files" ||
		len(res.DryRunPreview.GitArgv) == 0 || res.DryRunPreview.EffectSummary == "" ||
		len(res.DryRunPreview.GitCommands) != 1 || len(res.DryRunPreview.Targets) != 1 {
		t.Fatalf("dry_run_preview: %+v", res.DryRunPreview)
	}
	out, err := exec.Command("git", "-C", root, "diff", "--cached", "--name-only").CombinedOutput()
	if err != nil {
		t.Fatalf("git diff: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("expected nothing staged after dry-run, got %q", string(out))
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
		RepoRoot:  root,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  policy.AlwaysApprove{},
		Git:       ex,
		Audit:     audit.NewLogger(nil),
	}
	h := stage_files.Handler{Pipeline: pipe, Actor: "test"}
	res, err := h.Execute(ctx, root, stage_files.Params{Paths: []string{"no-such-file-in-repo.txt"}})
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
