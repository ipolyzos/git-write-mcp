// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

// MCPGitHub is the GitHub REST surface used by stdio MCP (PR create + PR status reads).
type MCPGitHub interface {
	PRWriter
	PRStatusReader
}

var _ MCPGitHub = (*Client)(nil)
