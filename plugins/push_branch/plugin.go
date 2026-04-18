// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package push_branch registers the push_branch MCP tool and handler.
package push_branch

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(meta{})
}

type meta struct{}

func (meta) Name() string { return "push_branch" }
