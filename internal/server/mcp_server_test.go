// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
	"github.com/ipolyzos/git-write-mcp/internal/server"
)

func TestNewMCPServerExposesExpectedTools(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	srv := server.NewMCPServer("0.0.0-test", plugin.Default)

	ct, st := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}

	cli := mcp.NewClient(&mcp.Implementation{Name: "git-write-mcp-test-client", Version: "v1"}, nil)
	cs, err := cli.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cs.Close(); err != nil {
			t.Errorf("Close client session: %v", err)
		}
	})

	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	sort.Strings(names)

	want := []string{"commit_changes", "create_branch", "stage_files", "tag_revision"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("ListTools names = %v, want %v", names, want)
	}
}
