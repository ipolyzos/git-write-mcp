// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
)

const stubToolDescription = "Controlled Git write operation (behavior not implemented yet)."

// NewMCPServer builds an MCP [mcp.Server] with stub tool handlers for every name in reg.
func NewMCPServer(version string, reg *plugin.Registry) *mcp.Server {
	if reg == nil {
		panic("server: NewMCPServer with nil registry")
	}
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "git-write-mcp",
		Version: version,
	}, nil)

	for _, name := range reg.Names() {
		name := name
		mcp.AddTool(s, &mcp.Tool{
			Name:        name,
			Description: stubToolDescription,
		}, func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("%s: not implemented", name)},
				},
			}, nil, nil
		})
	}
	return s
}

// RunStdio serves the MCP protocol on stdin/stdout using tools registered in reg.
func RunStdio(ctx context.Context, version string, reg *plugin.Registry) error {
	s := NewMCPServer(version, reg)
	return s.Run(ctx, &mcp.StdioTransport{})
}
