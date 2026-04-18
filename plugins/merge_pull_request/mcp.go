// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package merge_pull_request

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// RegisterMCPTool registers merge_pull_request on the MCP server.
func RegisterMCPTool(s *mcp.Server, pipe *runtime.Pipeline, repoRoot string, actor string, gh github.MCPGitHub, owner, repo string) {
	if s == nil || pipe == nil || gh == nil {
		return
	}
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return
	}
	h := &Handler{
		Pipeline:    pipe,
		GitHub:      gh,
		GitHubOwner: owner,
		GitHubRepo:  repo,
		Actor:       actor,
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "merge_pull_request",
		Description: "Merge a GitHub PR via REST only when merge policy passes (checks, reviews, mergeability, base branch allowlist GIT_WRITE_MCP_ALLOWED_PR_TARGET_BRANCHES, merge_method allowlist GIT_WRITE_MCP_MERGE_ALLOWED_METHODS). Requires approval and GITHUB_* env; dry_run supported.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in Params) (*mcp.CallToolResult, ToolResult, error) {
		res, err := h.Execute(ctx, absRoot, in)
		if err != nil {
			b, _ := json.Marshal(res)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(b)},
				},
			}, res, nil
		}
		return nil, res, nil
	})
}
