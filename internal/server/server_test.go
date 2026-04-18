// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"reflect"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/server"
)

func TestBootstrapRegistersExpectedPluginNames(t *testing.T) {
	t.Helper()
	b := server.NewBootstrap()
	got := b.PluginNames()
	want := []string{"commit_changes", "create_branch", "stage_files", "tag_revision"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PluginNames() = %v, want %v", got, want)
	}
}
