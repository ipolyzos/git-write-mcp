// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_pull_request

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// RegisterMCPTool registers the create_pull_request tool on the MCP server.
func RegisterMCPTool(s *mcp.Server, pipe *runtime.Pipeline, repoRoot string, actor string, gh github.PRWriter, owner, repo string) {
	if s == nil || pipe == nil {
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
		Name:        "create_pull_request",
		Description: "Open a GitHub pull request via REST API from source_branch into target_branch (default develop). Requires GITHUB_TOKEN and repo env; target allowlist GIT_WRITE_MCP_ALLOWED_PR_TARGET_BRANCHES. Approval and dry-run via runtime.Pipeline.",
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
