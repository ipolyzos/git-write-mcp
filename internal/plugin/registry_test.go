// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"reflect"
	"testing"
)

type fakePlugin struct {
	name string
}

func (f *fakePlugin) Name() string { return f.name }

func TestRegistryNamesSorted(t *testing.T) {
	t.Helper()
	r := NewRegistry()
	r.Register(&fakePlugin{name: "zebra"})
	r.Register(&fakePlugin{name: "alpha"})
	got := r.Names()
	want := []string{"alpha", "zebra"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
}
