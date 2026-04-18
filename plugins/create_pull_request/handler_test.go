// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_pull_request_test

import (
	"context"
	"encoding/json"
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
	"github.com/ipolyzos/git-write-mcp/plugins/create_pull_request"
)

func TestExecute_ValidPRCreation(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")

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

	gh := &fakeGH{
		out: &github.PullRequest{
			Number:  42,
			HTMLURL: "https://github.com/acme/repo/pull/42",
			State:   "open",
			Title:   "Feature work",
			HeadSHA: "abc",
		},
	}
	h := create_pull_request.Handler{
		Pipeline:    pipe,
		GitHub:      gh,
		GitHubOwner: "acme",
		GitHubRepo:  "demo",
		Actor:       "test",
	}

	res, err := h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "feature/from-handler",
		TargetBranch: "develop",
		Title:        "Feature work",
		Body:         "Describe change.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "execute" || res.PullRequestNumber != 42 {
		t.Fatalf("%+v", res)
	}
	if res.HTMLURL != gh.out.HTMLURL || res.Title != "Feature work" {
		t.Fatalf("%+v", res)
	}
	if len(gh.calls) != 1 || gh.calls[0].Head != "feature/from-handler" ||
		gh.calls[0].Base != "develop" || gh.calls[0].Title != "Feature work" {
		t.Fatalf("calls: %+v", gh.calls)
	}
}

func TestExecute_InvalidSourceBranch(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root, Policy: eng, RepoProbe: git.NewRepoProbe(ex),
		Git: ex, Approver: policy.AlwaysApprove{}, Audit: audit.NewLogger(nil),
	}
	h := create_pull_request.Handler{Pipeline: pipe, GitHub: &fakeGH{}, Actor: "x"}
	_, err = h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "..",
		TargetBranch: "develop",
		Title:        "Bad",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "source_branch" {
		t.Fatalf("want source_branch validation, got %v", err)
	}
}

func TestExecute_InvalidTargetBranch(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	root := t.TempDir()
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root, Policy: eng, RepoProbe: git.NewRepoProbe(ex),
		Git: ex, Approver: policy.AlwaysApprove{}, Audit: audit.NewLogger(nil),
	}
	h := create_pull_request.Handler{Pipeline: pipe, GitHub: &fakeGH{}, Actor: "x"}
	_, err = h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "feature/ok",
		TargetBranch: "unauthorized-target",
		Title:        "x",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecute_ApprovalDenied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
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
		RepoRoot: root, Policy: eng, RepoProbe: git.NewRepoProbe(ex),
		Git: ex, Approver: policy.DenyAll{}, Audit: audit.NewLogger(nil),
	}
	h := create_pull_request.Handler{Pipeline: pipe, GitHub: &fakeGH{}, Actor: "x"}
	_, err = h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "feature/a",
		Title:        "t",
	})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("%v", err)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
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
		RepoRoot: root, Policy: eng, RepoProbe: git.NewRepoProbe(ex),
		Git: ex, Approver: policy.AlwaysApprove{}, Audit: audit.NewLogger(nil),
	}
	gh := &fakeGH{out: &github.PullRequest{Number: 1}}
	h := create_pull_request.Handler{
		Pipeline: pipe, GitHub: gh, Actor: "x",
		GitHubOwner: "acme", GitHubRepo: "demo",
	}
	res, err := h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "feature/dry",
		Title:        "Dry title",
		DryRun:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "dry-run" || res.DryRunPreview == nil {
		t.Fatalf("%+v", res)
	}
	if res.DryRunPreview.Action != "create_pull_request" {
		t.Fatal(res.DryRunPreview.Action)
	}
	if len(gh.calls) != 0 {
		t.Fatal("dry-run must not call GitHub API")
	}
}

func TestExecute_GitHubAPIFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
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
		RepoRoot: root, Policy: eng, RepoProbe: git.NewRepoProbe(ex),
		Git: ex, Approver: policy.AlwaysApprove{}, Audit: audit.NewLogger(nil),
	}
	h := create_pull_request.Handler{
		Pipeline: pipe,
		GitHub:   &fakeGH{apiErr: errors.New("GitHub API is down")},
		Actor:    "x",
	}
	res, err := h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "feature/api",
		Title:        "T",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.OK {
		t.Fatal("expect ok=false")
	}
}

func TestExecute_NilGitHubIntegration(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	root := t.TempDir()
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(root)
	if err != nil {
		t.Fatal(err)
	}
	pipe := &runtime.Pipeline{
		RepoRoot: root, Policy: eng, RepoProbe: git.NewRepoProbe(ex),
		Git: ex, Approver: policy.AlwaysApprove{}, Audit: audit.NewLogger(nil),
	}
	h := create_pull_request.Handler{Pipeline: pipe, GitHub: nil, Actor: "x"}
	_, err = h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "feature/x",
		Title:        "T",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecute_DefaultTargetBranchDevelop(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
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
		RepoRoot: root, Policy: eng, RepoProbe: git.NewRepoProbe(ex),
		Git: ex, Approver: policy.AlwaysApprove{}, Audit: audit.NewLogger(nil),
	}
	gh := &fakeGH{out: &github.PullRequest{Number: 7, Title: "T"}}
	h := create_pull_request.Handler{Pipeline: pipe, GitHub: gh, Actor: "x"}
	_, err = h.Execute(ctx, root, create_pull_request.Params{
		SourceBranch: "feature/default-target",
		Title:        "T",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gh.calls) != 1 || gh.calls[0].Base != "develop" {
		t.Fatalf("calls: %+v", gh.calls)
	}
}

func TestToolResult_JSONDeterministic(t *testing.T) {
	t.Helper()
	a := create_pull_request.ToolResult{
		OK:                true,
		SourceBranch:      "feature/a",
		TargetBranch:      "develop",
		Title:             "T",
		PullRequestNumber: 1,
		HTMLURL:           "https://example.com/pull/1",
		Mode:              "execute",
	}
	b1, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if string(b1) != string(b2) {
		t.Fatal("non-deterministic json")
	}
}

type fakeGH struct {
	out    *github.PullRequest
	apiErr error
	calls  []github.CreatePullRequestInput
}

func (f *fakeGH) CreatePullRequest(ctx context.Context, in github.CreatePullRequestInput) (*github.PullRequest, error) {
	_ = ctx
	if f.apiErr != nil {
		return nil, f.apiErr
	}
	f.calls = append(f.calls, in)
	if f.out != nil {
		return f.out, nil
	}
	return &github.PullRequest{Number: 1}, nil
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
