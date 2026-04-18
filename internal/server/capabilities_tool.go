// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
)

func registerGetCapabilitiesTool(s *mcp.Server) {
	if s == nil {
		return
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        plugin.ToolGetCapabilities,
		Description: "Return deterministic JSON describing git-write-mcp write tools, safety behaviors (approval, dry-run, audit), and read-only Git tools intentionally delegated to other MCP servers.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		b, err := plugin.CapabilitiesJSON()
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: err.Error()},
				},
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(b)},
			},
		}, nil, nil
	})
}
