// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package get_pull_request_status_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/plugins/get_pull_request_status"
)

func TestExecute_ValidStatusRetrieval(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	mTrue := true
	status := &github.PullRequestStatus{
		Number:         7,
		Title:          "Test",
		State:          "open",
		Merged:         false,
		Draft:          false,
		Mergeable:      &mTrue,
		MergeableState: "clean",
		HTMLURL:        "https://github.com/o/r/pull/7",
		HeadRef:        "feature/a",
		BaseRef:        "develop",
		HeadSHA:        "abc",
		ChecksSummary: github.ChecksSummary{
			State: "success", TotalCount: 1,
			Contexts: []github.StatusContext{{Context: "ci", State: "success"}},
		},
		ReviewsSummary: github.ReviewsSummary{Approved: 1},
	}
	r := &fakeReader{status: status}
	h := get_pull_request_status.Handler{GitHub: r, Owner: "o", Repo: "r"}
	res, err := h.Execute(ctx, get_pull_request_status.Params{PullRequestNumber: 7})
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.PullRequest != status {
		t.Fatalf("%+v", res)
	}
	if len(r.calls) != 1 || r.calls[0] != 7 {
		t.Fatalf("calls %v", r.calls)
	}
}

func TestExecute_PRNotFound(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	h := get_pull_request_status.Handler{
		GitHub: &fakeReader{err: github.ErrPullRequestNotFound},
		Owner:  "o", Repo: "r",
	}
	res, err := h.Execute(ctx, get_pull_request_status.Params{PullRequestNumber: 404})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.OK || !errors.Is(err, github.ErrPullRequestNotFound) {
		t.Fatalf("got %v res=%+v", err, res)
	}
}

func TestExecute_GitHubAPIFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	h := get_pull_request_status.Handler{
		GitHub: &fakeReader{err: errors.New("upstream failure")},
		Owner:  "o", Repo: "r",
	}
	res, err := h.Execute(ctx, get_pull_request_status.Params{PullRequestNumber: 1})
	if err == nil {
		t.Fatal("expected error")
	}
	if res.OK {
		t.Fatal("expected ok=false")
	}
}

func TestExecute_NilGitHub(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	h := get_pull_request_status.Handler{GitHub: nil, Owner: "o", Repo: "r"}
	_, err := h.Execute(ctx, get_pull_request_status.Params{PullRequestNumber: 1})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecute_InvalidNumber(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	h := get_pull_request_status.Handler{GitHub: &fakeReader{}, Owner: "o", Repo: "r"}
	_, err := h.Execute(ctx, get_pull_request_status.Params{PullRequestNumber: 0})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestToolResult_JSONDeterministic(t *testing.T) {
	t.Helper()
	mTrue := true
	st := &github.PullRequestStatus{
		Number: 1, State: "open", HeadRef: "f", BaseRef: "develop",
		HeadSHA: "x", Mergeable: &mTrue, MergeableState: "clean",
		ChecksSummary: github.ChecksSummary{State: "pending", TotalCount: 0},
	}
	a := get_pull_request_status.ToolResult{OK: true, PullRequest: st}
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

type fakeReader struct {
	status *github.PullRequestStatus
	err    error
	calls  []int
}

func (f *fakeReader) GetPullRequestStatus(ctx context.Context, owner, repo string, number int) (*github.PullRequestStatus, error) {
	_ = ctx
	f.calls = append(f.calls, number)
	if f.err != nil {
		return nil, f.err
	}
	return f.status, nil
}
