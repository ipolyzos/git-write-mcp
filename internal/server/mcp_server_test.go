// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"encoding/json"
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

	srv := server.NewMCPServer("0.0.0-test", plugin.Default, server.ServeConfig{})

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

	want := []string{"commit_changes", "create_branch", "get_capabilities", "stage_files", "tag_revision"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("ListTools names = %v, want %v", names, want)
	}
}

func TestGetCapabilitiesTool_ReturnsStableCapabilitiesJSON(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	srv := server.NewMCPServer("0.0.0-test", plugin.Default, server.ServeConfig{})

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

	wantJSON, err := plugin.CapabilitiesJSON()
	if err != nil {
		t.Fatal(err)
	}

	tr, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: plugin.ToolGetCapabilities, Arguments: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(tr.Content) != 1 {
		t.Fatalf("content: %+v", tr.Content)
	}
	tc, ok := tr.Content[0].(*mcp.TextContent)
	if !ok || tc.Text == "" {
		t.Fatal("expected text content")
	}
	if tc.Text != string(wantJSON) {
		t.Fatalf("capabilities text mismatch")
	}

	var parsed plugin.Capabilities
	if err := json.Unmarshal([]byte(tc.Text), &parsed); err != nil {
		t.Fatal(err)
	}
	doc := plugin.DocumentCapabilities()
	if !reflect.DeepEqual(parsed, doc) {
		t.Fatalf("decoded != document: %+v vs %+v", parsed, doc)
	}
}
