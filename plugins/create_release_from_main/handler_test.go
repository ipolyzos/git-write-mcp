// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_release_from_main_test

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
	"github.com/ipolyzos/git-write-mcp/plugins/create_release_from_main"
)

func TestExecute_ValidReleaseFromMain(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvReleaseSourceBranch, "")

	root := initRepoOnMain(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	rel := &github.Release{ID: 9, HTMLURL: "https://github.com/o/r/releases/tag/v1", TagName: "v1.0.0", Name: "First"}
	f := &fakeRel{rel: rel}

	h := create_release_from_main.Handler{
		Pipeline: pipe, GitHub: f,
		GitHubOwner: "o", GitHubRepo: "r", Actor: "test",
	}
	res, err := h.Execute(ctx, root, create_release_from_main.Params{
		TagName:      "v1.0.0",
		ReleaseName:  "First",
		ReleaseNotes: "## Notes\n\nHello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "execute" || res.ReleaseID != 9 || f.calls != 1 {
		t.Fatalf("%+v calls=%d", res, f.calls)
	}
	if f.last.TagName != "v1.0.0" || f.last.TargetCommitish != "main" || f.last.Body != "## Notes\n\nHello" {
		t.Fatalf("%+v", f.last)
	}
}

func TestExecute_DeniedWhenNotOnMain(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	t.Setenv(policy.EnvReleaseSourceBranch, "")

	root := initRepoOnDevelop(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	h := create_release_from_main.Handler{
		Pipeline: pipe, GitHub: &fakeRel{},
		GitHubOwner: "o", GitHubRepo: "r", Actor: "x",
	}
	_, err := h.Execute(ctx, root, create_release_from_main.Params{
		TagName: "v1", ReleaseName: "R", ReleaseNotes: "x",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "git_branch" {
		t.Fatalf("%v", err)
	}
}

func TestExecute_MissingReleaseNotes(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initRepoOnMain(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	h := create_release_from_main.Handler{
		Pipeline: pipe, GitHub: &fakeRel{},
		GitHubOwner: "o", GitHubRepo: "r", Actor: "x",
	}
	_, err := h.Execute(ctx, root, create_release_from_main.Params{
		TagName: "v1", ReleaseName: "R",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecute_ApprovalDenied(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initRepoOnMain(t)
	pipe := testPipeline(t, root, policy.DenyAll{})
	h := create_release_from_main.Handler{
		Pipeline: pipe, GitHub: &fakeRel{},
		GitHubOwner: "o", GitHubRepo: "r", Actor: "x",
	}
	_, err := h.Execute(ctx, root, create_release_from_main.Params{
		TagName: "v1", ReleaseName: "R", ChangelogSection: "body",
	})
	if !errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatalf("%v", err)
	}
}

func TestExecute_DryRunSuccess(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initRepoOnMain(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	f := &fakeRel{}
	h := create_release_from_main.Handler{
		Pipeline: pipe, GitHub: f,
		GitHubOwner: "o", GitHubRepo: "r", Actor: "x",
	}
	res, err := h.Execute(ctx, root, create_release_from_main.Params{
		TagName: "v1", ReleaseName: "R", ReleaseNotes: "n", DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Mode != "dry-run" || f.calls != 0 {
		t.Fatalf("%+v", res)
	}
}

func TestExecute_GitHubAPIFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	root := initRepoOnMain(t)
	pipe := testPipeline(t, root, policy.AlwaysApprove{})
	f := &fakeRel{err: errors.New("API down")}
	h := create_release_from_main.Handler{
		Pipeline: pipe, GitHub: f,
		GitHubOwner: "o", GitHubRepo: "r", Actor: "x",
	}
	res, err := h.Execute(ctx, root, create_release_from_main.Params{
		TagName: "v1", ReleaseName: "R", ReleaseNotes: "n",
	})
	if err == nil || res.OK {
		t.Fatal("expected failure")
	}
	if f.calls != 1 {
		t.Fatal("release should be attempted")
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

func initRepoOnMain(t *testing.T) string {
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
	if out, err := exec.Command("git", "-C", dir, "branch", "-M", "main").CombinedOutput(); err != nil {
		t.Fatalf("git branch -M: %v\n%s", err, out)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func initRepoOnDevelop(t *testing.T) string {
	t.Helper()
	mainRoot := initRepoOnMain(t)
	if out, err := exec.Command("git", "-C", mainRoot, "checkout", "-b", "develop").CombinedOutput(); err != nil {
		t.Fatalf("checkout develop: %v\n%s", err, out)
	}
	return mainRoot
}

type fakeRel struct {
	rel   *github.Release
	err   error
	calls int
	last  github.CreateReleaseInput
}

func (f *fakeRel) CreatePullRequest(context.Context, github.CreatePullRequestInput) (*github.PullRequest, error) {
	return nil, github.ErrNotImplemented
}

func (f *fakeRel) GetPullRequestStatus(context.Context, string, string, int) (*github.PullRequestStatus, error) {
	return nil, github.ErrNotImplemented
}

func (f *fakeRel) MergePullRequest(context.Context, string, string, int, github.MergePullRequestInput) (*github.MergePullRequestResult, error) {
	return nil, github.ErrNotImplemented
}

func (f *fakeRel) CreateRelease(_ context.Context, in github.CreateReleaseInput) (*github.Release, error) {
	f.calls++
	f.last = in
	if f.err != nil {
		return nil, f.err
	}
	if f.rel != nil {
		return f.rel, nil
	}
	return &github.Release{ID: 1}, nil
}
