// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/audit"
	"github.com/ipolyzos/git-write-mcp/internal/git"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
	"github.com/ipolyzos/git-write-mcp/plugins/commit_changes"
	"github.com/ipolyzos/git-write-mcp/plugins/create_branch"
	"github.com/ipolyzos/git-write-mcp/plugins/stage_files"
	"github.com/ipolyzos/git-write-mcp/plugins/tag_revision"
)

func TestMutationTools_AllFourEmitExecuteAuditRecords(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := gitInitEmptyCommit(t)

	var mu sync.Mutex
	var recs []audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		mu.Lock()
		defer mu.Unlock()
		recs = append(recs, r)
		return nil
	})

	pipe := testPipeline(t, root, policy.AlwaysApprove{}, sink)

	cb := create_branch.Handler{Pipeline: pipe, Actor: "integration"}
	if _, err := cb.Execute(ctx, root, create_branch.Params{BranchName: "feature/integration-seq"}); err != nil {
		t.Fatal(err)
	}

	writeFile(t, root, "note.txt", "data")
	sf := stage_files.Handler{Pipeline: pipe, Actor: "integration"}
	if _, err := sf.Execute(ctx, root, stage_files.Params{Paths: []string{"note.txt"}}); err != nil {
		t.Fatal(err)
	}

	cc := commit_changes.Handler{Pipeline: pipe, Actor: "integration"}
	if _, err := cc.Execute(ctx, root, commit_changes.Params{Message: "Add note content"}); err != nil {
		t.Fatal(err)
	}

	tr := tag_revision.Handler{Pipeline: pipe, Actor: "integration"}
	if _, err := tr.Execute(ctx, root, tag_revision.Params{
		TagName: "v0-integration-seq",
		Message: "Integration tag",
	}); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(recs) != 4 {
		t.Fatalf("audit records = %d, want 4: %+v", len(recs), recs)
	}
	wantActions := []string{"create_branch", "stage_files", "commit_changes", "tag_revision"}
	for i, a := range wantActions {
		if recs[i].Action != a {
			t.Fatalf("record[%d].Action = %q, want %q", i, recs[i].Action, a)
		}
		if recs[i].Mode != audit.ModeExecute {
			t.Fatalf("record[%d].Mode = %q, want execute", i, recs[i].Mode)
		}
		if recs[i].Outcome != audit.OutcomeSuccess {
			t.Fatalf("record[%d].Outcome = %q", i, recs[i].Outcome)
		}
		if recs[i].Actor != "integration" {
			t.Fatalf("record[%d].Actor = %q", i, recs[i].Actor)
		}
	}
}

func TestMutationTools_DryRunEmitsDryRunAudit(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := gitInitEmptyCommit(t)

	var mu sync.Mutex
	var recs []audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		mu.Lock()
		defer mu.Unlock()
		recs = append(recs, r)
		return nil
	})

	pipe := testPipeline(t, root, policy.AlwaysApprove{}, sink)
	h := create_branch.Handler{Pipeline: pipe, Actor: "dry"}
	if _, err := h.Execute(ctx, root, create_branch.Params{
		BranchName: "feature/dry-audit",
		DryRun:     true,
	}); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(recs) != 1 {
		t.Fatalf("want 1 audit record, got %d %+v", len(recs), recs)
	}
	if recs[0].Mode != audit.ModeDryRun {
		t.Fatalf("mode = %q, want dry-run", recs[0].Mode)
	}
	if recs[0].Outcome != audit.OutcomeSuccess {
		t.Fatalf("outcome = %q", recs[0].Outcome)
	}
}

func TestMutationTools_ApprovalDeniedEmitsBlockedAudit(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := gitInitEmptyCommit(t)

	var mu sync.Mutex
	var recs []audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		mu.Lock()
		defer mu.Unlock()
		recs = append(recs, r)
		return nil
	})

	pipe := testPipeline(t, root, policy.DenyAll{}, sink)
	h := create_branch.Handler{Pipeline: pipe, Actor: "denied"}
	_, err := h.Execute(ctx, root, create_branch.Params{BranchName: "feature/blocked"})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(recs) != 1 {
		t.Fatalf("want 1 audit record, got %d %+v", len(recs), recs)
	}
	if recs[0].Mode != audit.ModeBlocked {
		t.Fatalf("mode = %q", recs[0].Mode)
	}
	if recs[0].Outcome != audit.OutcomeFailure {
		t.Fatalf("outcome = %q", recs[0].Outcome)
	}
}

func TestMutationTools_StageFilesPathOutsideRepoIsBlocked(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := gitInitEmptyCommit(t)

	outside := filepath.Join(filepath.Dir(root), "outside-int.txt")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(outside) })

	var mu sync.Mutex
	var recs []audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		mu.Lock()
		defer mu.Unlock()
		recs = append(recs, r)
		return nil
	})

	pipe := testPipeline(t, root, policy.AlwaysApprove{}, sink)
	h := stage_files.Handler{Pipeline: pipe, Actor: "boundary"}
	_, err := h.Execute(ctx, root, stage_files.Params{Paths: []string{"../" + filepath.Base(outside)}})
	if err == nil {
		t.Fatal("expected error")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(recs) != 1 {
		t.Fatalf("want 1 audit record, got %d %+v", len(recs), recs)
	}
	if recs[0].Mode != audit.ModeBlocked {
		t.Fatalf("mode = %q", recs[0].Mode)
	}
	if recs[0].Outcome != audit.OutcomeFailure {
		t.Fatalf("outcome = %q", recs[0].Outcome)
	}
	if !strings.Contains(recs[0].Detail, "outside") && !strings.Contains(strings.ToLower(recs[0].Detail), "repository") {
		t.Fatalf("unexpected detail: %q", recs[0].Detail)
	}
}

func testPipeline(t *testing.T, repoRoot string, appr policy.Approver, sink audit.Sink) *runtime.Pipeline {
	t.Helper()
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: repoRoot})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	return &runtime.Pipeline{
		RepoRoot:  repoRoot,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  appr,
		Git:       ex,
		Audit:     audit.NewLogger(sink),
	}
}

func gitInitEmptyCommit(t *testing.T) string {
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

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
