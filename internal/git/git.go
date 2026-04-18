// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package git defines boundaries for allowlisted Git CLI invocation (no shell, allowlist only; logic later).
package git

// Executor runs allowlisted git commands via the external git binary (not implemented in bootstrap).
type Executor struct{}
