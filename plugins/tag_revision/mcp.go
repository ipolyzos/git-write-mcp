// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package tag_revision

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// RegisterMCPTool registers the tag_revision tool on the MCP server.
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
		Name:        "tag_revision",
		Description: "Create an annotated tag with git tag -a -m (no -f/-d). Optional target_ref defaults to HEAD.",
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
