// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
	"github.com/ipolyzos/git-write-mcp/plugins/commit_changes"
	"github.com/ipolyzos/git-write-mcp/plugins/create_branch"
	"github.com/ipolyzos/git-write-mcp/plugins/stage_files"
	"github.com/ipolyzos/git-write-mcp/plugins/tag_revision"
)

const stubToolDescription = "Controlled Git write operation (behavior not implemented yet)."

// NewMCPServer builds an MCP server. Tools with dedicated implementations skip the generic stub.
func NewMCPServer(version string, reg *plugin.Registry, cfg ServeConfig) *mcp.Server {
	if reg == nil {
		panic("server: NewMCPServer with nil registry")
	}
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "git-write-mcp",
		Version: version,
	}, nil)

	implemented := map[string]struct{}{}
	if cfg.Pipeline != nil && cfg.RepoRoot != "" {
		implemented["commit_changes"] = struct{}{}
		implemented["create_branch"] = struct{}{}
		implemented["stage_files"] = struct{}{}
		implemented["tag_revision"] = struct{}{}
		commit_changes.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		create_branch.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		stage_files.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		tag_revision.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
	}

	for _, name := range reg.Names() {
		name := name
		if _, skip := implemented[name]; skip {
			continue
		}
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

// RunStdio serves MCP on stdin/stdout using cfg for repository-backed tools.
func RunStdio(ctx context.Context, version string, reg *plugin.Registry, cfg ServeConfig) error {
	s := NewMCPServer(version, reg, cfg)
	return s.Run(ctx, &mcp.StdioTransport{})
}
