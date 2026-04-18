// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
	"github.com/ipolyzos/git-write-mcp/internal/server"
	"github.com/ipolyzos/git-write-mcp/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version)
		return
	}

	repoRoot := os.Getenv("GIT_WRITE_MCP_REPO_ROOT")
	if repoRoot == "" {
		fmt.Fprintln(os.Stderr, "git-write-mcp: GIT_WRITE_MCP_REPO_ROOT is required")
		os.Exit(1)
	}
	cfg, err := server.BuildServeConfig(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-write-mcp: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := server.RunStdio(ctx, version.Version, plugin.Default, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "git-write-mcp: %v\n", err)
		os.Exit(1)
	}
}
