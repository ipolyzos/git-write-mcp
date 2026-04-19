// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
)

func TestCapabilitiesDocument_WriteTools(t *testing.T) {
	t.Helper()
	got := plugin.DocumentCapabilities()
	want := []string{"commit_changes", "create_branch", "create_pull_request", "create_release_from_main", "merge_pull_request", "push_branch", "stage_files", "tag_revision", "unstage_files"}
	if !slices.Equal(got.WriteTools, want) {
		t.Fatalf("WriteTools = %v, want %v", got.WriteTools, want)
	}
}

func TestCapabilitiesDocument_ReadTools(t *testing.T) {
	t.Helper()
	got := plugin.DocumentCapabilities()
	want := []string{"get_pull_request_status"}
	if !slices.Equal(got.ReadTools, want) {
		t.Fatalf("ReadTools = %v, want %v", got.ReadTools, want)
	}
}

func TestCapabilitiesDocument_ExcludedReadTools(t *testing.T) {
	t.Helper()
	got := plugin.DocumentCapabilities()
	want := []string{
		"git_blame",
		"git_branch_list",
		"git_diff",
		"git_log",
		"git_show",
		"git_status",
	}
	if !slices.Equal(got.ExcludedReadTools, want) {
		t.Fatalf("ExcludedReadTools = %v, want %v", got.ExcludedReadTools, want)
	}
}

func TestCapabilitiesDocument_SharedSafetyFeatures(t *testing.T) {
	t.Helper()
	b := plugin.DocumentCapabilities().Behaviors
	if !b.ApprovalRequired || !b.DryRunSupported || !b.AuditTrailEnabled {
		t.Fatalf("behaviors: %+v", b)
	}
}

func TestCapabilitiesJSON_Deterministic(t *testing.T) {
	t.Helper()
	a, err := plugin.CapabilitiesJSON()
	if err != nil {
		t.Fatal(err)
	}
	b, err := plugin.CapabilitiesJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatalf("non-deterministic json")
	}
	var m1, m2 map[string]any
	if err := json.Unmarshal(a, &m1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &m2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m1, m2) {
		t.Fatal("decoded maps differ")
	}
}
