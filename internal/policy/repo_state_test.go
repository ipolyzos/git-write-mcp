// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/git"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func TestValidateRepoState_ValidRepositoryState(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	eng := mustEngine(t, root)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	probe := git.NewRepoProbe(ex)

	in := policy.RepoStateCheckInput{
		RepoRoot: root,
		WorkDir:  root,
		Action:   "create_branch",
	}
	got := policy.ValidateRepoState(ctx, eng, probe, in)
	if !got.OK || got.Code != policy.RepoValidationOK {
		t.Fatalf("want ok, got %+v", got)
	}
}

func TestValidateRepoState_RepoOutsideBoundary(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	eng := mustEngine(t, root)
	in := policy.RepoStateCheckInput{
		RepoRoot: root,
		WorkDir:  outside,
		Action:   "stage_files",
	}
	got := policy.ValidateRepoState(ctx, eng, noopProbe{}, in)
	if got.OK || got.Code != policy.RepoValidationBoundary {
		t.Fatalf("want boundary failure, got %+v", got)
	}
}

func TestValidateRepoState_MissingRepo(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := filepath.Join(t.TempDir(), "does-not-exist-999")
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	in := policy.RepoStateCheckInput{
		RepoRoot: root,
		WorkDir:  root,
		Action:   "create_branch",
	}
	got := policy.ValidateRepoState(ctx, eng, noopProbe{}, in)
	if got.OK || got.Code != policy.RepoValidationMissing {
		t.Fatalf("want missing, got %+v", got)
	}
}

func TestValidateRepoState_DetachedHEAD(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	if out, err := exec.Command("git", "-C", root, "checkout", "--detach", "HEAD").CombinedOutput(); err != nil {
		t.Fatalf("detach: %v\n%s", err, out)
	}
	eng := mustEngine(t, root)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	probe := git.NewRepoProbe(ex)

	in := policy.RepoStateCheckInput{
		RepoRoot:           root,
		WorkDir:            root,
		Action:             "create_branch",
		RejectDetachedHEAD: true,
	}
	got := policy.ValidateRepoState(ctx, eng, probe, in)
	if got.OK || got.Code != policy.RepoValidationDetached {
		t.Fatalf("want detached failure, got %+v", got)
	}
}

func TestValidateRepoState_TargetRefNotFound(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	eng := mustEngine(t, root)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	probe := git.NewRepoProbe(ex)

	in := policy.RepoStateCheckInput{
		RepoRoot:  root,
		WorkDir:   root,
		Action:    "tag_revision",
		TargetRef: "refs/heads/does-not-exist-branch-xyz",
	}
	got := policy.ValidateRepoState(ctx, eng, probe, in)
	if got.OK || got.Code != policy.RepoValidationRefNotFound {
		t.Fatalf("want ref_not_found, got %+v", got)
	}
}

func TestValidateRepoState_InvalidTargetRefSyntax(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	eng := mustEngine(t, root)
	in := policy.RepoStateCheckInput{
		RepoRoot:  root,
		WorkDir:   root,
		Action:    "tag_revision",
		TargetRef: "bad ref syntax",
	}
	got := policy.ValidateRepoState(ctx, eng, noopProbe{}, in)
	if got.OK || got.Code != policy.RepoValidationInvalidRef {
		t.Fatalf("want invalid ref, got %+v", got)
	}
}

func TestValidateRepoState_InvalidActionPreconditions(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	eng := mustEngine(t, root)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	probe := git.NewRepoProbe(ex)

	in := policy.RepoStateCheckInput{
		RepoRoot:           root,
		WorkDir:            root,
		Action:             "commit_changes",
		RequireStagedIndex: true,
	}
	got := policy.ValidateRepoState(ctx, eng, probe, in)
	if got.OK || got.Code != policy.RepoValidationPrecondition {
		t.Fatalf("want precondition failure, got %+v", got)
	}
}

func TestValidateRepoState_RequireHEADBranch_Match(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	def, err := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	cur := strings.TrimSpace(string(def))
	eng := mustEngine(t, root)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	probe := git.NewRepoProbe(ex)
	in := policy.RepoStateCheckInput{
		RepoRoot:          root,
		WorkDir:           root,
		Action:            "push_branch",
		RequireHEADBranch: cur,
	}
	got := policy.ValidateRepoState(ctx, eng, probe, in)
	if !got.OK {
		t.Fatalf("%+v", got)
	}
}

func TestValidateRepoState_RequireHEADBranch_Mismatch(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	def, err := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	prev := strings.TrimSpace(string(def))
	if out, err := exec.Command("git", "-C", root, "checkout", "-b", "feature/other-branch").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	eng := mustEngine(t, root)
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	probe := git.NewRepoProbe(ex)
	in := policy.RepoStateCheckInput{
		RepoRoot:          root,
		WorkDir:           root,
		Action:            "push_branch",
		RequireHEADBranch: prev,
	}
	got := policy.ValidateRepoState(ctx, eng, probe, in)
	if got.OK || got.Code != policy.RepoValidationPrecondition {
		t.Fatalf("%+v", got)
	}
}

func TestValidateRepoState_DeterministicValidationOutput(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initGitRepo(t)
	eng := mustEngine(t, root)
	in := policy.RepoStateCheckInput{
		RepoRoot: root,
		WorkDir:  filepath.Join(root, "..", "..", filepath.Base(root)),
		Action:   "stage_files",
	}
	a := policy.ValidateRepoState(ctx, eng, noopProbe{}, in)
	b := policy.ValidateRepoState(ctx, eng, noopProbe{}, in)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("non-deterministic: %+v vs %+v", a, b)
	}
	j1, _ := json.Marshal(a)
	j2, _ := json.Marshal(b)
	if string(j1) != string(j2) {
		t.Fatal("json differs")
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, cmd := range [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "config", "user.email", "t@t"},
		{"git", "-C", dir, "config", "user.name", "t"},
		{"git", "-C", dir, "commit", "--allow-empty", "-m", "Add repository baseline"},
	} {
		if out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", cmd, err, out)
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func mustEngine(t *testing.T, root string) *policy.Engine {
	t.Helper()
	e, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

type noopProbe struct{}

func (noopProbe) GitRepositoryOK(context.Context, string) error { return nil }
func (noopProbe) IsDetachedHEAD(context.Context, string) (bool, error) {
	return false, nil
}
func (noopProbe) RefExists(context.Context, string, string) (bool, error) { return true, nil }
func (noopProbe) HasStagedChanges(context.Context, string) (bool, error) {
	return false, nil
}

func (noopProbe) CurrentBranchName(context.Context, string) (string, error) {
	return "", nil
}
