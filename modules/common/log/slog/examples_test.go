package slog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"

	cslog "github.com/guidomantilla/yarumo/common/log/slog"
	"github.com/guidomantilla/yarumo/common/log/slog/slogctx"
)

// keys returns the sorted keys of m. Helps make example output deterministic.
func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}

// ExampleNewLogger_contextAttrs demonstrates how context-bound attributes are
// added to every log record automatically without the caller needing to pass
// them on every Info/Warn/Error call.
func ExampleNewLogger_contextAttrs() {
	buf := &bytes.Buffer{}
	logger := cslog.NewLogger(
		cslog.WithWriter(buf),
		cslog.WithLevel(cslog.LevelInfo),
		cslog.WithContextExtractors(cslog.SlogctxExtractor),
	)

	ctx := slogctx.WithAttrs(context.Background(),
		slog.String("request_id", "req-001"),
		slog.String("user_id", "u-42"),
	)

	logger.Info(ctx, "request received", "method", "GET")

	var got map[string]any

	err := json.Unmarshal(buf.Bytes(), &got)
	if err != nil {
		fmt.Println("parse error:", err)

		return
	}

	for _, k := range keys(got) {
		if k == "time" {
			continue
		}

		fmt.Printf("%s=%v\n", k, got[k])
	}

	// Output:
	// level=INFO
	// method=GET
	// msg=request received
	// request_id=req-001
	// user_id=u-42
}

// ExampleNewLogger_contextAttrsSetAfterBinding shows how middleware can
// enrich a context-bound bag after the context has already been propagated
// to inner handlers.
func ExampleNewLogger_contextAttrsSetAfterBinding() {
	buf := &bytes.Buffer{}
	logger := cslog.NewLogger(
		cslog.WithWriter(buf),
		cslog.WithLevel(cslog.LevelInfo),
		cslog.WithContextExtractors(cslog.SlogctxExtractor),
	)

	ctx := slogctx.WithAttrs(context.Background(), slog.String("request_id", "req-001"))

	// Later, middleware resolves the user and enriches the bag in place.
	slogctx.SetAttrs(ctx, slog.String("user_id", "u-42"))

	logger.Info(ctx, "request completed")

	var got map[string]any

	err := json.Unmarshal(buf.Bytes(), &got)
	if err != nil {
		fmt.Println("parse error:", err)

		return
	}

	fmt.Println("request_id=" + got["request_id"].(string))
	fmt.Println("user_id=" + got["user_id"].(string))

	// Output:
	// request_id=req-001
	// user_id=u-42
}
