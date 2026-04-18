// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/audit"
	"github.com/ipolyzos/git-write-mcp/internal/git"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// ServeConfig carries dependencies for stdio MCP (repository + mutation pipeline).
type ServeConfig struct {
	RepoRoot string
	Pipeline *runtime.Pipeline
	Actor    string
}

// BuildServeConfig constructs [ServeConfig] from an absolute or relative repository root path.
func BuildServeConfig(repoRoot string) (ServeConfig, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return ServeConfig{}, fmt.Errorf("server: empty repo root")
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return ServeConfig{}, err
	}
	eng, err := policy.NewEngine(abs)
	if err != nil {
		return ServeConfig{}, err
	}
	ex, err := git.NewExecutor(git.ExecutorOptions{RepoRoot: abs})
	if err != nil {
		return ServeConfig{}, err
	}
	lg := audit.NewLogger(nil)
	pipe := &runtime.Pipeline{
		RepoRoot:  abs,
		Policy:    eng,
		RepoProbe: git.NewRepoProbe(ex),
		Approver:  ApproverFromEnv(),
		Git:       ex,
		Audit:     lg,
	}
	actor := strings.TrimSpace(os.Getenv("GIT_WRITE_MCP_ACTOR"))
	if actor == "" {
		actor = "mcp"
	}
	return ServeConfig{RepoRoot: abs, Pipeline: pipe, Actor: actor}, nil
}
