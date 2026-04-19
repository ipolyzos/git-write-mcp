// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package merge_pull_request_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/audit"
	"github.com/ipolyzos/git-write-mcp/internal/git"
	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
	"github.com/ipolyzos/git-write-mcp/plugins/merge_pull_request"
)

func goodStatus() *github.PullRequestStatus {
	return &github.PullRequestStatus{
		Number: 5, State: "open", Merged: false, Draft: false,
		MergeableState: "clean", BaseRef: "develop", HeadRef: "f", HeadSHA: "abc",
		ChecksSummary:  github.ChecksSummary{State: "success", TotalCount: 1},
		ReviewsSummary: github.ReviewsSummary{Approved: 1, ChangesRequested: 0},
	}
}

func TestExecute_ValidMergeWhenPolicySatisfied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	t.Setenv(policy.EnvMergeAllowedMethods, "squash")
	t.Setenv(policy.EnvMergeRequireGreenChecks, "1")
	t.Setenv(policy.EnvMergeMinApprovals, "1")
	t.Setenv(policy.EnvMergeBlockOnChangesRequested, "1")
	t.Setenv(policy.EnvMergeRequireCleanMergeability, "1")

	root := initGitRepoWithCommit(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})

	gh := &fakeMCP{
		status:   goodStatus(),
		mergeRes: &github.MergePullRequestResult{Merged: true, SHA: "def", Message: "ok"},
	}
	h := merge_pull_request.Handler{
		Pipeline: pipe, GitHub: gh,
		GitHubOwner: "o", GitHubRepo: "r", Actor: "test",
	}
	res, err := h.Execute(ctx, root, merge_pull_request.Params{
		PullRequestNumber: 5,
		MergeMethod:       "squash",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "execute" || res.MergedSHA != "def" || gh.mergeCalls != 1 {
		t.Fatalf("%+v calls=%d", res, gh.mergeCalls)
	}
	if gh.mergeLast.MergeMethod != "squash" {
		t.Fatal(gh.mergeLast)
	}
}

func TestExecute_DeniedWhenChecksRed(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")

	root := initGitRepoWithCommit(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	st := goodStatus()
	st.ChecksSummary.State = "failure"

	gh := &fakeMCP{status: st}
	h := merge_pull_request.Handler{Pipeline: pipe, GitHub: gh, GitHubOwner: "o", GitHubRepo: "r", Actor: "x"}
	_, err := h.Execute(ctx, root, merge_pull_request.Params{PullRequestNumber: 5})
	if err == nil || !errors.Is(err, policy.ErrMergeChecksNotSuccessful) {
		t.Fatalf("got %v", err)
	}
	if gh.mergeCalls != 0 {
		t.Fatal("merge must not run")
	}
}

func TestExecute_DeniedWhenReviewsInsufficient(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	t.Setenv(policy.EnvMergeMinApprovals, "2")

	root := initGitRepoWithCommit(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	st := goodStatus()
	st.ReviewsSummary.Approved = 1

	gh := &fakeMCP{status: st}
	h := merge_pull_request.Handler{Pipeline: pipe, GitHub: gh, GitHubOwner: "o", GitHubRepo: "r", Actor: "x"}
	_, err := h.Execute(ctx, root, merge_pull_request.Params{PullRequestNumber: 5})
	if err == nil || !errors.Is(err, policy.ErrMergeInsufficientApprovals) {
		t.Fatalf("got %v", err)
	}
	if gh.mergeCalls != 0 {
		t.Fatal("merge must not run")
	}
}

func TestExecute_ApprovalDenied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")

	root := initGitRepoWithCommit(t)
	pipe := testPipeline(t, root, policy.DenyAll{})
	gh := &fakeMCP{status: goodStatus()}
	h := merge_pull_request.Handler{Pipeline: pipe, GitHub: gh, GitHubOwner: "o", GitHubRepo: "r", Actor: "x"}
	_, err := h.Execute(ctx, root, merge_pull_request.Params{PullRequestNumber: 5})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("%v", err)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")

	root := initGitRepoWithCommit(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	gh := &fakeMCP{status: goodStatus()}
	h := merge_pull_request.Handler{Pipeline: pipe, GitHub: gh, GitHubOwner: "o", GitHubRepo: "r", Actor: "x"}
	res, err := h.Execute(ctx, root, merge_pull_request.Params{
		PullRequestNumber: 5,
		DryRun:            true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "dry-run" || res.DryRunPreview == nil {
		t.Fatalf("%+v", res)
	}
	if gh.mergeCalls != 0 {
		t.Fatal("dry-run must not call merge")
	}
}

func TestExecute_GitHubAPIFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")

	root := initGitRepoWithCommit(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	gh := &fakeMCP{
		status:   goodStatus(),
		mergeErr: errors.New("GitHub refused merge"),
	}
	h := merge_pull_request.Handler{Pipeline: pipe, GitHub: gh, GitHubOwner: "o", GitHubRepo: "r", Actor: "x"}
	res, err := h.Execute(ctx, root, merge_pull_request.Params{PullRequestNumber: 5})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.OK {
		t.Fatal("ok=false")
	}
	if gh.mergeCalls != 1 {
		t.Fatal("merge should have been attempted")
	}
}

func testPipeline(t *testing.T, repoRoot string, appr policy.Approver) *runtime.Pipeline {
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
		Audit:     audit.NewLogger(nil),
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

type fakeMCP struct {
	status     *github.PullRequestStatus
	mergeRes   *github.MergePullRequestResult
	mergeErr   error
	mergeCalls int
	mergeLast  github.MergePullRequestInput
}

func (f *fakeMCP) CreatePullRequest(context.Context, github.CreatePullRequestInput) (*github.PullRequest, error) {
	return nil, github.ErrNotImplemented
}

func (f *fakeMCP) GetPullRequestStatus(context.Context, string, string, int) (*github.PullRequestStatus, error) {
	return f.status, nil
}

func (f *fakeMCP) MergePullRequest(_ context.Context, _, _ string, _ int, in github.MergePullRequestInput) (*github.MergePullRequestResult, error) {
	f.mergeCalls++
	f.mergeLast = in
	if f.mergeErr != nil {
		return nil, f.mergeErr
	}
	if f.mergeRes != nil {
		return f.mergeRes, nil
	}
	return &github.MergePullRequestResult{Merged: true}, nil
}

func (f *fakeMCP) CreateRelease(context.Context, github.CreateReleaseInput) (*github.Release, error) {
	return nil, github.ErrNotImplemented
}
