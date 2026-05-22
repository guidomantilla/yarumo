package slog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	cslog "github.com/guidomantilla/yarumo/log/slog"
	"go.opentelemetry.io/otel/trace"
)

// ExampleWithOtelTrace demonstrates trace ↔ log correlation. When a span is
// active in the context, the logger emits trace_id and span_id on every record.
func ExampleWithOtelTrace() {
	buf := &bytes.Buffer{}
	logger := cslog.NewLogger(
		cslog.WithWriter(buf),
		cslog.WithLevel(cslog.LevelInfo),
		WithOtelTrace(),
	)

	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("1112131415161718")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	logger.Info(ctx, "request processed")

	var got map[string]any

	err := json.Unmarshal(buf.Bytes(), &got)
	if err != nil {
		fmt.Println("parse error:", err)

		return
	}

	fmt.Printf("trace_id=%v\n", got[AttrTraceID])
	fmt.Printf("span_id=%v\n", got[AttrSpanID])
	fmt.Printf("trace_flags=%v\n", got[AttrTraceFlags])

	// Output:
	// trace_id=0102030405060708090a0b0c0d0e0f10
	// span_id=1112131415161718
	// trace_flags=01
}
