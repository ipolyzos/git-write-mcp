// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

// MCPGitHub is the GitHub REST surface used by stdio MCP (PR create, reads, merge, releases).
type MCPGitHub interface {
	PRWriter
	PRStatusReader
	PRMerger
	ReleaseCreator
}

var _ MCPGitHub = (*Client)(nil)
