// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package get_pull_request_status

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
)

// RegisterMCPTool registers get_pull_request_status on the MCP server.
func RegisterMCPTool(s *mcp.Server, gh github.PRStatusReader, owner, repo string) {
	if s == nil || gh == nil {
		return
	}
	h := &Handler{GitHub: gh, Owner: owner, Repo: repo}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_pull_request_status",
		Description: "Read-only GitHub PR readiness: state, mergeability, combined checks for head SHA, review counts, head/base refs. Requires GITHUB_TOKEN and repository env at server start. Does not mutate git or GitHub.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in Params) (*mcp.CallToolResult, ToolResult, error) {
		res, err := h.Execute(ctx, in)
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
