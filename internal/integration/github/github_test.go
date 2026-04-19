// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	gh "github.com/ipolyzos/git-write-mcp/internal/integration/github"
)

func TestConfig_Validate_Success(t *testing.T) {
	t.Helper()
	c := gh.Config{
		Repository: gh.Repository{Owner: "octocat", Name: "Hello-World"},
		Token:      "ghp_secret_example",
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestConfig_Validate_MissingOwner(t *testing.T) {
	t.Helper()
	c := gh.Config{
		Repository: gh.Repository{Owner: " ", Name: "Hello-World"},
		Token:      "tok",
	}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *gh.ValidationError
	if !errors.As(err, &ve) || ve.Field != "owner" {
		t.Fatalf("got %v", err)
	}
}

func TestConfig_Validate_MissingName(t *testing.T) {
	t.Helper()
	c := gh.Config{
		Repository: gh.Repository{Owner: "octocat", Name: ""},
		Token:      "tok",
	}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *gh.ValidationError
	if !errors.As(err, &ve) || ve.Field != "name" {
		t.Fatalf("got %v", err)
	}
}

func TestConfig_Validate_MissingToken(t *testing.T) {
	t.Helper()
	c := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "",
	}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *gh.ValidationError
	if !errors.As(err, &ve) || ve.Field != "token" {
		t.Fatalf("got %v", err)
	}
}

func TestConfig_RedactedSummary_NeverContainsToken(t *testing.T) {
	t.Helper()
	secret := "ghp_super_secret_value_must_not_appear"
	c := gh.Config{
		Repository: gh.Repository{Owner: "acme", Name: "demo"},
		Token:      secret,
	}
	s := c.RedactedSummary()
	if strings.Contains(s, secret) {
		t.Fatalf("summary leaked token: %q", s)
	}
	if !strings.Contains(s, `owner="acme"`) || !strings.Contains(s, `name="demo"`) {
		t.Fatalf("unexpected summary: %q", s)
	}
}

func TestNewClient_ValidConfig(t *testing.T) {
	t.Helper()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}"))}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cl == nil {
		t.Fatal("nil client")
	}
}

func TestNewClient_MissingRequiredConfig(t *testing.T) {
	t.Helper()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "", Name: "r"},
		Token:      "t",
	}
	_, err := gh.NewClient(cfg, fakeHTTP(nil))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClient_NilHTTPClient(t *testing.T) {
	t.Helper()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	_, err := gh.NewClient(cfg, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_ImplementsIntegration(t *testing.T) {
	t.Helper()
	var _ gh.Integration = (*gh.Client)(nil)
}

func TestDecodePullRequestResponse_Deterministic(t *testing.T) {
	t.Helper()
	body := []byte(`{
  "number": 42,
  "title": "Example",
  "state": "open",
  "html_url": "https://github.com/o/r/pull/42",
  "head": {"sha": "abc123"}
}`)
	a, err := gh.DecodePullRequestResponse(body)
	if err != nil {
		t.Fatal(err)
	}
	b, err := gh.DecodePullRequestResponse(body)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("non-deterministic decode: %+v vs %+v", a, b)
	}
	if a.Number != 42 || a.Title != "Example" || a.State != "open" || a.HeadSHA != "abc123" {
		t.Fatalf("unexpected mapping: %+v", a)
	}
}

func TestClient_DoAuthorizedGET_DeterministicRequest(t *testing.T) {
	t.Helper()
	var seenMethod, seenPath string
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "octocat", Name: "Hello-World"},
		Token:      "secret-token-for-auth-header-only",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(req *http.Request) (*http.Response, error) {
		seenMethod = req.Method
		seenPath = req.URL.Path
		auth := req.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Fatalf("want Bearer scheme")
		}
		if strings.TrimPrefix(auth, "Bearer ") != cfg.Token {
			t.Fatal("authorization bearer token mismatch")
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"number":1}`)),
			Header:     make(http.Header),
		}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	res, err := cl.DoAuthorizedGET(ctx, "/repos/octocat/Hello-World/pulls/1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if seenMethod != http.MethodGet {
		t.Fatalf("method = %q", seenMethod)
	}
	if seenPath != "/repos/octocat/Hello-World/pulls/1" {
		t.Fatalf("path = %q", seenPath)
	}
}

func TestClient_DoAuthorizedGET_ErrorBodyOmitsToken(t *testing.T) {
	t.Helper()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "ultra-secret-x",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = cl.DoAuthorizedGET(ctx, "/fail")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), cfg.Token) {
		t.Fatal("error leaked token")
	}
}

func TestLoadConfigFromEnv_Success(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GITHUB_REPOSITORY_OWNER", "acme")
	t.Setenv("GITHUB_REPOSITORY_NAME", "proj")
	c, err := gh.LoadConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.Repository.Owner != "acme" || c.Repository.Name != "proj" || c.Token != "tok" {
		t.Fatalf("%+v", c)
	}
}

func TestLoadConfigFromEnv_FromGITHUB_REPOSITORY(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GITHUB_REPOSITORY", "octocat/Spoon-Knife")
	t.Setenv("GITHUB_REPOSITORY_OWNER", "")
	t.Setenv("GITHUB_REPOSITORY_NAME", "")
	c, err := gh.LoadConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if c.Repository.Owner != "octocat" || c.Repository.Name != "Spoon-Knife" {
		t.Fatalf("%+v", c)
	}
}

func TestLoadConfigFromEnv_MissingToken(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_REPOSITORY_OWNER", "o")
	t.Setenv("GITHUB_REPOSITORY_NAME", "r")
	_, err := gh.LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_CreatePullRequest_Success(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/repos/o/r/pulls" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		return &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(strings.NewReader(`{"number":5,"title":"Hi","state":"open","html_url":"https://example.com/5","head":{"sha":"sha1"}}`)),
		}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	pr, err := cl.CreatePullRequest(ctx, gh.CreatePullRequestInput{
		Title: "Hi",
		Head:  "feature/a",
		Base:  "develop",
	})
	if err != nil {
		t.Fatal(err)
	}
	if pr.Number != 5 || pr.HeadSHA != "sha1" || pr.Title != "Hi" || pr.HTMLURL != "https://example.com/5" || pr.State != "open" {
		t.Fatalf("%+v", pr)
	}
}

func TestClient_GetPullRequestStatus_Success(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	prJSON := `{"number":3,"title":"T","state":"open","merged":false,"draft":false,"mergeable":true,"mergeable_state":"clean","html_url":"https://ex/3","head":{"ref":"feature/x","sha":"shaabc"},"base":{"ref":"develop"}}`
	statJSON := `{"state":"success","total_count":2,"statuses":[{"state":"success","context":"ci"},{"state":"failure","context":"audit"}]}`
	revJSON := `[{"state":"APPROVED"},{"state":"COMMENTED"}]`
	cl, err := gh.NewClient(cfg, fakeHTTP(func(req *http.Request) (*http.Response, error) {
		p := req.URL.Path
		switch {
		case p == "/repos/o/r/pulls/3" && req.Method == http.MethodGet:
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(prJSON))}, nil
		case p == "/repos/o/r/commits/shaabc/status":
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(statJSON))}, nil
		case p == "/repos/o/r/pulls/3/reviews":
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(revJSON))}, nil
		default:
			t.Fatalf("unexpected %s %s", req.Method, p)
			return nil, nil
		}
	}))
	if err != nil {
		t.Fatal(err)
	}
	st, err := cl.GetPullRequestStatus(ctx, "o", "r", 3)
	if err != nil {
		t.Fatal(err)
	}
	if st.Number != 3 || st.HeadRef != "feature/x" || st.BaseRef != "develop" || st.HeadSHA != "shaabc" {
		t.Fatalf("%+v", st)
	}
	if st.MergeableState != "clean" || st.Mergeable == nil || !*st.Mergeable {
		t.Fatalf("%+v", st)
	}
	if st.ChecksSummary.State != "success" || st.ChecksSummary.TotalCount != 2 || len(st.ChecksSummary.Contexts) != 2 {
		t.Fatalf("%+v", st.ChecksSummary)
	}
	if st.ChecksSummary.Contexts[0].Context != "audit" || st.ChecksSummary.Contexts[1].Context != "ci" {
		t.Fatalf("contexts must be sorted: %+v", st.ChecksSummary.Contexts)
	}
	if st.ReviewsSummary.Approved != 1 || st.ReviewsSummary.Commented != 1 {
		t.Fatalf("%+v", st.ReviewsSummary)
	}
}

func TestClient_GetPullRequestStatus_NotFound(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
		}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = cl.GetPullRequestStatus(ctx, "o", "r", 99)
	if !errors.Is(err, gh.ErrPullRequestNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestClient_GetPullRequestStatus_HTTPErrorOnPR(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`boom`)),
		}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = cl.GetPullRequestStatus(ctx, "o", "r", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_MergePullRequest_Success(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPut || req.URL.Path != "/repos/o/r/pulls/9/merge" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"merged":true,"message":"Merged","sha":"zzz"}`)),
		}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	res, err := cl.MergePullRequest(ctx, "o", "r", 9, gh.MergePullRequestInput{MergeMethod: "squash"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Merged || res.SHA != "zzz" || res.Message != "Merged" {
		t.Fatalf("%+v", res)
	}
}

func TestClient_CreateRelease_Success(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/repos/o/r/releases" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(`{"id":7,"html_url":"https://ex/7","tag_name":"v1","name":"R"}`)),
		}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	res, err := cl.CreateRelease(ctx, gh.CreateReleaseInput{
		TagName: "v1", Name: "R", Body: "b", TargetCommitish: "main",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != 7 || res.HTMLURL != "https://ex/7" {
		t.Fatalf("%+v", res)
	}
}

func TestClient_GetPullRequestStatus_RepoMismatch(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cfg := gh.Config{
		Repository: gh.Repository{Owner: "o", Name: "r"},
		Token:      "t",
	}
	cl, err := gh.NewClient(cfg, fakeHTTP(nil))
	if err != nil {
		t.Fatal(err)
	}
	_, err = cl.GetPullRequestStatus(ctx, "other", "r", 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

type fakeHTTP func(*http.Request) (*http.Response, error)

func (f fakeHTTP) Do(req *http.Request) (*http.Response, error) {
	if f == nil {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))}, nil
	}
	return f(req)
}
