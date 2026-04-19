// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_release_from_main

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// RegisterMCPTool registers create_release_from_main on the MCP server.
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
		Name:        "create_release_from_main",
		Description: "Create a GitHub release from the configured release branch (default main): requires local checkout on that branch (GIT_WRITE_MCP_RELEASE_SOURCE_BRANCH), tag_name, release_name, and release_notes or changelog_section. Approval, dry-run, audit via runtime.Pipeline.",
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
