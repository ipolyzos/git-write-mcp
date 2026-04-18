// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

func TestBuildDryRunPreview_CreateBranch(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:  "create_branch",
		Target:  "feature/preview",
		WorkDir: root,
		GitArgs: []string{"branch", "feature/preview"},
	}
	got := runtime.BuildDryRunPreview(req)
	if got.Action != "create_branch" {
		t.Fatalf("action = %q", got.Action)
	}
	if !reflect.DeepEqual(got.Targets, []string{"feature/preview"}) {
		t.Fatalf("targets = %#v", got.Targets)
	}
	if len(got.SecondaryRefs) != 0 {
		t.Fatalf("secondary_refs = %#v", got.SecondaryRefs)
	}
	wantSummary := `Would create branch "feature/preview" from the current HEAD commit.`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
	if len(got.GitCommands) != 1 || got.GitCommands[0] == "" {
		t.Fatalf("git_commands = %#v", got.GitCommands)
	}
}

func TestBuildDryRunPreview_PushBranch(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:     "push_branch",
		Target:     "feature/push",
		PushRemote: "origin",
		WorkDir:    root,
		GitArgs:    []string{"push", "origin", "feature/push"},
	}
	got := runtime.BuildDryRunPreview(req)
	if !reflect.DeepEqual(got.Targets, []string{"feature/push"}) {
		t.Fatalf("targets = %#v", got.Targets)
	}
	if !reflect.DeepEqual(got.SecondaryRefs, []string{"origin"}) {
		t.Fatalf("secondary_refs = %#v", got.SecondaryRefs)
	}
	wantSummary := `Would push branch "feature/push" to remote "origin" (no force, no refspec).`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
}

func TestBuildDryRunPreview_CreateBranch_FromRef(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:  "create_branch",
		Target:  "feature/from-main",
		WorkDir: root,
		GitArgs: []string{"branch", "feature/from-main", "main"},
		FromRef: "main",
	}
	got := runtime.BuildDryRunPreview(req)
	wantSummary := `Would create branch "feature/from-main" starting from ref "main".`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
	if !reflect.DeepEqual(got.SecondaryRefs, []string{"main"}) {
		t.Fatalf("secondary_refs = %#v", got.SecondaryRefs)
	}
}

func TestBuildDryRunPreview_UnstageFiles(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:  "unstage_files",
		Target:  "a.txt,b.txt",
		WorkDir: root,
		GitArgs: []string{"reset", "HEAD", "--", "a.txt", "b.txt"},
	}
	got := runtime.BuildDryRunPreview(req)
	if !reflect.DeepEqual(got.Targets, []string{"a.txt", "b.txt"}) {
		t.Fatalf("targets = %#v", got.Targets)
	}
	wantSummary := `Would unstage 2 path(s) from the index (working tree unchanged): "a.txt", "b.txt".`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
}

func TestBuildDryRunPreview_StageFiles(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:  "stage_files",
		Target:  "a.txt,b.txt",
		WorkDir: root,
		GitArgs: []string{"add", "--", "a.txt", "b.txt"},
	}
	got := runtime.BuildDryRunPreview(req)
	if !reflect.DeepEqual(got.Targets, []string{"a.txt", "b.txt"}) {
		t.Fatalf("targets = %#v", got.Targets)
	}
	wantSummary := `Would stage 2 path(s) for the next commit: "a.txt", "b.txt".`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
}

func TestBuildDryRunPreview_CommitChanges(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:  "commit_changes",
		Target:  "Add widget",
		WorkDir: root,
		GitArgs: []string{"commit", "-m", "Add widget\n"},
	}
	got := runtime.BuildDryRunPreview(req)
	wantSummary := `Would create a new commit with the staged changes using the provided message.`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
	if !reflect.DeepEqual(got.Targets, []string{"Add widget"}) {
		t.Fatalf("targets = %#v", got.Targets)
	}
}

func TestBuildDryRunPreview_TagRevision(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:    "tag_revision",
		Target:    "v1.0.0",
		WorkDir:   root,
		GitArgs:   []string{"tag", "-a", "v1.0.0", "-m", "Release"},
		TargetRef: "",
	}
	got := runtime.BuildDryRunPreview(req)
	wantSummary := `Would create annotated tag "v1.0.0" at HEAD.`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
	if !reflect.DeepEqual(got.Targets, []string{"v1.0.0"}) {
		t.Fatalf("targets = %#v", got.Targets)
	}
}

func TestBuildDryRunPreview_TagRevision_TargetRef(t *testing.T) {
	t.Helper()
	root := filepath.Clean("/tmp/repo")
	req := &runtime.MutationRequest{
		Action:    "tag_revision",
		Target:    "v1.0.0@main",
		WorkDir:   root,
		GitArgs:   []string{"tag", "-a", "v1.0.0", "-m", "Release", "main"},
		TargetRef: "main",
	}
	got := runtime.BuildDryRunPreview(req)
	wantSummary := `Would create annotated tag "v1.0.0" at ref "main".`
	if got.EffectSummary != wantSummary {
		t.Fatalf("effect_summary = %q, want %q", got.EffectSummary, wantSummary)
	}
	if !reflect.DeepEqual(got.SecondaryRefs, []string{"main"}) {
		t.Fatalf("secondary_refs = %#v", got.SecondaryRefs)
	}
}

func TestDryRunPreview_JSONDeterministic(t *testing.T) {
	t.Helper()
	req := &runtime.MutationRequest{
		Action:  "create_branch",
		Target:  "x",
		WorkDir: "/tmp/r",
		GitArgs: []string{"branch", "x"},
	}
	a := runtime.BuildDryRunPreview(req)
	b := runtime.BuildDryRunPreview(req)
	j1, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	j2, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(j1) != string(j2) {
		t.Fatalf("json differs: %s vs %s", j1, j2)
	}
}
