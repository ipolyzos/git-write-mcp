// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package create_branch registers the create_branch MCP tool and handler.
package create_branch

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(meta{})
}

type meta struct{}

func (meta) Name() string { return "create_branch" }
