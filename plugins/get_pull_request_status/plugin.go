// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package get_pull_request_status registers the get_pull_request_status MCP tool.
package get_pull_request_status

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(meta{})
}

type meta struct{}

func (meta) Name() string { return "get_pull_request_status" }
