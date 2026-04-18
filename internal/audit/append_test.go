// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package audit_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ipolyzos/git-write-mcp/internal/audit"
)

func TestAppend_AuditEntrySuccessfulExecution(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	var got audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		got = r
		return nil
	})
	lg := audit.NewLogger(sink)
	ts := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	meta, err := audit.Append(ctx, lg, audit.AppendInput{
		Timestamp:     ts,
		Actor:         "actor1",
		Action:        "create_branch",
		Target:        "feature/x",
		Mode:          audit.ModeExecute,
		Outcome:       audit.OutcomeSuccess,
		AuditID:       "audit-fixed-exec",
		CorrelationID: "corr-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if meta.AuditID != "audit-fixed-exec" || meta.CorrelationID != "corr-1" {
		t.Fatalf("metadata: %+v", meta)
	}
	if got.Mode != audit.ModeExecute || got.Outcome != audit.OutcomeSuccess {
		t.Fatalf("record: %+v", got)
	}
	if got.AuditID != "audit-fixed-exec" || got.CorrelationID != "corr-1" {
		t.Fatalf("record ids: %+v", got)
	}
}

func TestAppend_AuditEntryDeniedApproval(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	var got audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		got = r
		return nil
	})
	lg := audit.NewLogger(sink)
	ts := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	_, err := audit.Append(ctx, lg, audit.AppendInput{
		Timestamp: ts,
		Actor:     "a",
		Action:    "stage_files",
		Target:    "x.txt",
		Mode:      audit.ModeBlocked,
		Outcome:   audit.OutcomeFailure,
		Detail:    "approval denied",
		AuditID:   "audit-deny",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != audit.ModeBlocked || got.Outcome != audit.OutcomeFailure {
		t.Fatalf("record: %+v", got)
	}
	if !strings.Contains(got.Detail, "approval") {
		t.Fatalf("detail: %q", got.Detail)
	}
}

func TestAppend_AuditEntryDryRun(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	var got audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		got = r
		return nil
	})
	lg := audit.NewLogger(sink)
	ts := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	_, err := audit.Append(ctx, lg, audit.AppendInput{
		Timestamp: ts,
		Actor:     "u",
		Action:    "commit_changes",
		Target:    "msg",
		Mode:      audit.ModeDryRun,
		Outcome:   audit.OutcomeSuccess,
		AuditID:   "audit-dry",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != audit.ModeDryRun {
		t.Fatalf("mode = %q", got.Mode)
	}
}

func TestAppend_DeterministicFieldPresence(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error { return nil })
	lg := audit.NewLogger(sink)
	ts := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	_, err := audit.Append(ctx, lg, audit.AppendInput{
		Timestamp: ts,
		Actor:     "a",
		Action:    "tag_revision",
		Target:    "v1",
		Mode:      audit.ModeExecute,
		Outcome:   audit.OutcomeSuccess,
		AuditID:   "id-presence",
	})
	if err != nil {
		t.Fatal(err)
	}
	r := audit.Record{
		Timestamp: ts,
		Actor:     "a",
		Action:    "tag_revision",
		Target:    "v1",
		Mode:      audit.ModeExecute,
		Outcome:   audit.OutcomeSuccess,
		AuditID:   "id-presence",
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"timestamp", "actor", "action", "target", "mode", "outcome", "audit_id"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing json key %q in %s", k, string(b))
		}
	}
}

func TestAppend_CorrelationIDPassthrough(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	var got audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		got = r
		return nil
	})
	lg := audit.NewLogger(sink)
	meta, err := audit.Append(ctx, lg, audit.AppendInput{
		Timestamp:     time.Now().UTC(),
		Actor:         "x",
		Action:        "create_branch",
		Target:        "y",
		Mode:          audit.ModeExecute,
		Outcome:       audit.OutcomeSuccess,
		CorrelationID: "trace-abc",
		AuditID:       "fixed-corr",
	})
	if err != nil {
		t.Fatal(err)
	}
	if meta.CorrelationID != "trace-abc" || got.CorrelationID != "trace-abc" {
		t.Fatalf("want correlation trace-abc, meta=%+v got=%+v", meta, got)
	}
}

func TestAppend_GeneratesAuditIDWhenEmpty(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	var got audit.Record
	sink := audit.SinkFunc(func(_ context.Context, r audit.Record) error {
		got = r
		return nil
	})
	lg := audit.NewLogger(sink)
	meta, err := audit.Append(ctx, lg, audit.AppendInput{
		Timestamp: time.Now().UTC(),
		Actor:     "a",
		Action:    "create_branch",
		Target:    "z",
		Mode:      audit.ModeExecute,
		Outcome:   audit.OutcomeSuccess,
	})
	if err != nil {
		t.Fatal(err)
	}
	if meta.AuditID == "" || got.AuditID == "" {
		t.Fatal("expected generated audit id")
	}
	if meta.AuditID != got.AuditID {
		t.Fatalf("meta vs record audit id: %q vs %q", meta.AuditID, got.AuditID)
	}
}

func TestAppend_SinkFailure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	wantErr := errors.New("disk full")
	sink := audit.SinkFunc(func(context.Context, audit.Record) error {
		return wantErr
	})
	lg := audit.NewLogger(sink)
	_, err := audit.Append(ctx, lg, audit.AppendInput{
		Timestamp: time.Now().UTC(),
		Actor:     "a",
		Action:    "create_branch",
		Target:    "b",
		Mode:      audit.ModeExecute,
		Outcome:   audit.OutcomeSuccess,
		AuditID:   "audit-err",
	})
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("want sink error, got %v", err)
	}
}

func TestAppend_NilLoggerStillReturnsMetadata(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	meta, err := audit.Append(ctx, nil, audit.AppendInput{
		Timestamp: time.Now().UTC(),
		Actor:     "a",
		Action:    "create_branch",
		Target:    "c",
		Mode:      audit.ModeExecute,
		Outcome:   audit.OutcomeSuccess,
		AuditID:   "meta-only",
	})
	if err != nil {
		t.Fatal(err)
	}
	if meta.AuditID != "meta-only" {
		t.Fatalf("metadata: %+v", meta)
	}
}
