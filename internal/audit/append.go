// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// AppendInput is the stable contract for recording one mutation audit event.
type AppendInput struct {
	Timestamp time.Time
	Actor     string
	Action    string
	Target    string
	Mode      string
	Outcome   string
	Detail    string

	// CorrelationID optional trace key supplied by the caller (e.g. MCP request id).
	CorrelationID string
	// AuditID optional; when empty a new unique id is generated for this append.
	AuditID string
}

// AppendMetadata is returned to callers so pipelines and tools can reference the
// emitted audit entry without rereading the sink.
type AppendMetadata struct {
	AuditID       string `json:"audit_id"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

// Append validates input, assigns an audit id when missing, forwards to the logger's
// sink, and returns metadata. A nil logger is treated as no sink: metadata is still
// returned and no error unless the caller's sink fails.
func Append(ctx context.Context, lg *Logger, in AppendInput) (AppendMetadata, error) {
	rec := toRecord(in)
	meta := AppendMetadata{
		AuditID:       rec.AuditID,
		CorrelationID: rec.CorrelationID,
	}
	var sink Sink
	if lg != nil {
		sink = lg.sink
	}
	if sink == nil {
		return meta, nil
	}
	err := sink.Emit(ctx, rec)
	return meta, err
}

func toRecord(in AppendInput) Record {
	ts := in.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	} else {
		ts = ts.UTC()
	}
	id := strings.TrimSpace(in.AuditID)
	if id == "" {
		id = newAuditID()
	}
	corr := strings.TrimSpace(in.CorrelationID)
	return Record{
		Timestamp:     ts,
		Actor:         in.Actor,
		Action:        in.Action,
		Target:        in.Target,
		Mode:          in.Mode,
		Outcome:       in.Outcome,
		Detail:        in.Detail,
		AuditID:       id,
		CorrelationID: corr,
	}
}

func newAuditID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "audit-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b[:])
}
