// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"os"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

// ApproverFromEnv selects a centralized [policy.Approver].
// GIT_WRITE_MCP_AUTO_APPROVE=1|true|yes enables [policy.AlwaysApprove]; otherwise [policy.DenyAll].
func ApproverFromEnv() policy.Approver {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GIT_WRITE_MCP_AUTO_APPROVE"))) {
	case "1", "true", "yes", "on":
		return policy.AlwaysApprove{}
	default:
		return policy.DenyAll{}
	}
}
