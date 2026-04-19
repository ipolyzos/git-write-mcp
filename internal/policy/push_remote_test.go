// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func TestAllowedPushRemotes_Default(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPushRemotes, "")
	got := policy.AllowedPushRemotes()
	want := []string{"origin"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAllowedPushRemotes_FromEnv_DedupSorted(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPushRemotes, "fork, origin , fork, upstream")
	got := policy.AllowedPushRemotes()
	want := []string{"fork", "origin", "upstream"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestValidateAllowedPushRemote_Allowed(t *testing.T) {
	t.Helper()
	if err := policy.ValidateAllowedPushRemote("origin", []string{"fork", "origin"}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAllowedPushRemote_NotAllowed(t *testing.T) {
	t.Helper()
	err := policy.ValidateAllowedPushRemote("evil", []string{"origin"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "remote" {
		t.Fatalf("got %v", err)
	}
}

func TestValidateAllowedPushRemote_EmptyRemote(t *testing.T) {
	t.Helper()
	err := policy.ValidateAllowedPushRemote("  ", []string{"origin"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "remote" {
		t.Fatalf("got %v", err)
	}
}
