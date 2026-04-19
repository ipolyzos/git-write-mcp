// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
	"github.com/ipolyzos/git-write-mcp/internal/server"
)

// TestMCP_AllRegisteredTools_AutoApprove exercises every ListTools entry with
// GIT_WRITE_MCP_AUTO_APPROVE=1: git mutations use dry_run (or a no-op error
// for create_pull_request when GitHub is not configured). When GITHUB_TOKEN is
// unset, GitHub-only tools are the generic plugin stubs ("not implemented").
func TestMCP_AllRegisteredTools_AutoApprove(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_WRITE_MCP_AUTO_APPROVE", "1")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_REPOSITORY", "")

	ctx := context.Background()
	root := gitInitEmptyCommit(t)
	writeFile(t, root, "track.txt", "track")

	cfg, err := server.BuildServeConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GitHub != nil {
		t.Fatal("test expects no GitHub client (clear GITHUB_TOKEN in environment)")
	}

	srv := server.NewMCPServer("0.0.0-mcp-all", plugin.Default, cfg)
	ct, st := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	cli := mcp.NewClient(&mcp.Implementation{Name: "mcp-all-test", Version: "v1"}, nil)
	cs, err := cli.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tools) < 1 {
		t.Fatalf("ListTools: want at least one tool, got %d", len(res.Tools))
	}
	for _, tool := range res.Tools {
		if tool == nil || tool.Name == "" {
			t.Fatalf("ListTools: bad tool entry: %#v", tool)
		}
	}

	cur := gitCurrentBranch(t, root)

	// 1) get_capabilities
	m1 := callToolJSON(t, ctx, cs, "get_capabilities", map[string]any{})
	if _, ok := m1["write_tools"]; !ok {
		t.Fatalf("get_capabilities: missing write_tools: %#v", m1)
	}

	// 2–8) Git-backed tools (dry_run or one stage execute for commit dry-run path)
	assertToolOK(t, callToolJSON(t, ctx, cs, "create_branch", map[string]any{
		"branch_name": "feature/mcp-all-tools",
		"dry_run":     true,
	}))
	assertToolOK(t, callToolJSON(t, ctx, cs, "stage_files", map[string]any{
		"paths":   []string{"track.txt"},
		"dry_run": true,
	}))
	assertToolOK(t, callToolJSON(t, ctx, cs, "stage_files", map[string]any{
		"paths":   []string{"track.txt"},
		"dry_run": false,
	}))
	assertToolOK(t, callToolJSON(t, ctx, cs, "commit_changes", map[string]any{
		"message": "mcp all-tools dry-run",
		"dry_run": true,
	}))
	assertToolOK(t, callToolJSON(t, ctx, cs, "unstage_files", map[string]any{
		"paths":   []string{"track.txt"},
		"dry_run": true,
	}))
	assertToolOK(t, callToolJSON(t, ctx, cs, "tag_revision", map[string]any{
		"tag_name": "v0.0.0-mcp-all",
		"message":  "smoke tag",
		"dry_run":  true,
	}))
	assertToolOK(t, callToolJSON(t, ctx, cs, "push_branch", map[string]any{
		"branch_name": cur,
		"dry_run":     true,
	}))

	// 9) create_pull_request — real handler; GitHub nil → validation error (still callable).
	mPR := callToolJSON(t, ctx, cs, "create_pull_request", map[string]any{
		"source_branch": "feature/foo",
		"title":         "smoke",
		"dry_run":       true,
	})
	if ok, _ := mPR["ok"].(bool); ok {
		t.Fatalf("create_pull_request: expected ok=false without GitHub, got %#v", mPR)
	}
	errStr, _ := mPR["error"].(string)
	if !strings.Contains(errStr, "GitHub") {
		t.Fatalf("create_pull_request: expected GitHub configuration error, got %#v", mPR)
	}

	// 10–12) Stub plugins when cfg.GitHub == nil (merge, PR status, release).
	for _, name := range []string{"get_pull_request_status", "merge_pull_request", "create_release_from_main"} {
		tr := callToolRaw(t, ctx, cs, name, map[string]any{"pull_request_number": json.Number("1")})
		if len(tr.Content) != 1 {
			t.Fatalf("%s: unexpected content count", name)
		}
		tc, ok := tr.Content[0].(*mcp.TextContent)
		if !ok || !strings.Contains(tc.Text, "not implemented") {
			t.Fatalf("%s: want stub body, got %#v", name, tr.Content)
		}
	}
}

func assertToolOK(t *testing.T, m map[string]any) {
	t.Helper()
	ok, _ := m["ok"].(bool)
	if !ok {
		t.Fatalf("expected ok=true, got %#v", m)
	}
}

func gitCurrentBranch(t *testing.T, root string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", root, "branch", "--show-current").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func callToolRaw(t *testing.T, ctx context.Context, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	tr, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return tr
}

func callToolJSON(t *testing.T, ctx context.Context, cs *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()
	tr := callToolRaw(t, ctx, cs, name, args)
	text := toolResultText(tr)
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("%s: unmarshal: %v body=%q", name, err, text)
	}
	return m
}

func toolResultText(tr *mcp.CallToolResult) string {
	if tr == nil || len(tr.Content) == 0 {
		return "{}"
	}
	tc, ok := tr.Content[0].(*mcp.TextContent)
	if !ok || tc.Text == "" {
		return "{}"
	}
	return tc.Text
}
