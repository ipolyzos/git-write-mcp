// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package create_branch registers the create_branch tool (logic not implemented).
package create_branch

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(stub{})
}

type stub struct{}

func (stub) Name() string { return "create_branch" }
