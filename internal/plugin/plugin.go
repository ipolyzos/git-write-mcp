// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package plugin defines the MCP tool plugin surface and registration.
package plugin

// Plugin identifies a stable MCP tools/list and tools/call target (tool name).
// Plugins register via [Registry.Register], typically from an init function in plugins/*.
// Execution policy and Git invocation are layered above this contract later.
type Plugin interface {
	Name() string
}
