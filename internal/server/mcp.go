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
	"github.com/ipolyzos/git-write-mcp/plugins/create_pull_request"
	"github.com/ipolyzos/git-write-mcp/plugins/create_release_from_main"
	"github.com/ipolyzos/git-write-mcp/plugins/get_pull_request_status"
	"github.com/ipolyzos/git-write-mcp/plugins/merge_pull_request"
	"github.com/ipolyzos/git-write-mcp/plugins/push_branch"
	"github.com/ipolyzos/git-write-mcp/plugins/stage_files"
	"github.com/ipolyzos/git-write-mcp/plugins/tag_revision"
	"github.com/ipolyzos/git-write-mcp/plugins/unstage_files"
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

	registerGetCapabilitiesTool(s)

	implemented := map[string]struct{}{
		plugin.ToolGetCapabilities: {},
	}
	if cfg.Pipeline != nil && cfg.RepoRoot != "" {
		implemented["commit_changes"] = struct{}{}
		implemented["create_branch"] = struct{}{}
		implemented["push_branch"] = struct{}{}
		implemented["create_pull_request"] = struct{}{}
		implemented["stage_files"] = struct{}{}
		implemented["tag_revision"] = struct{}{}
		implemented["unstage_files"] = struct{}{}
		commit_changes.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		create_branch.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		push_branch.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		create_pull_request.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor, cfg.GitHub, cfg.GitHubOwner, cfg.GitHubRepo)
		stage_files.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		tag_revision.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		unstage_files.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor)
		if cfg.GitHub != nil {
			implemented["get_pull_request_status"] = struct{}{}
			get_pull_request_status.RegisterMCPTool(s, cfg.GitHub, cfg.GitHubOwner, cfg.GitHubRepo)
			implemented["merge_pull_request"] = struct{}{}
			merge_pull_request.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor, cfg.GitHub, cfg.GitHubOwner, cfg.GitHubRepo)
			implemented["create_release_from_main"] = struct{}{}
			create_release_from_main.RegisterMCPTool(s, cfg.Pipeline, cfg.RepoRoot, cfg.Actor, cfg.GitHub, cfg.GitHubOwner, cfg.GitHubRepo)
		}
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
