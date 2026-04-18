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

	ctx := context.Background()
	if err := server.RunStdio(ctx, version.Version, plugin.Default); err != nil {
		fmt.Fprintf(os.Stderr, "git-write-mcp: %v\n", err)
		os.Exit(1)
	}
}
