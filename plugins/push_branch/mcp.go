// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package push_branch

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// RegisterMCPTool registers the push_branch tool on the MCP server.
func RegisterMCPTool(s *mcp.Server, pipe *runtime.Pipeline, repoRoot string, actor string) {
	if s == nil || pipe == nil {
		return
	}
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return
	}
	h := &Handler{Pipeline: pipe, Actor: actor}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "push_branch",
		Description: "Push a branch to an allowlisted remote with git push <remote> <branch> only (no force, no refspec). Remote defaults to origin; allowlist from GIT_WRITE_MCP_ALLOWED_PUSH_REMOTES. Requires approval.",
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
