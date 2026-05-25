package zerolog

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEmit(t *testing.T) {
	t.Parallel()

	t.Run("nil event is a no-op", func(t *testing.T) {
		t.Parallel()

		// Should not panic when event is nil.
		emit(nil, context.Background(), "msg")
	})

	t.Run("nil context is tolerated", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		//nolint:staticcheck // explicitly testing nil-context tolerance.
		l.Info(nil, "nil-ctx-emit", "k", "v")

		if !strings.Contains(buf.String(), "nil-ctx-emit") {
			t.Fatalf("output %q does not contain %q", buf.String(), "nil-ctx-emit")
		}
	})
}

func TestApplyArgs_AllTypes(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))

	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	dur := 5 * time.Second
	boom := errors.New("boom")

	l.Info(
		context.Background(),
		"all-types",
		"str", "value",
		"bool", true,
		"int", int(1),
		"int8", int8(2),
		"int16", int16(3),
		"int32", int32(4),
		"int64", int64(5),
		"uint", uint(6),
		"uint8", uint8(7),
		"uint16", uint16(8),
		"uint32", uint32(9),
		"uint64", uint64(10),
		"float32", float32(1.5),
		"float64", float64(2.5),
		"time", now,
		"dur", dur,
		"err", boom,
		"nil", nil,
		"other", struct{ A int }{A: 7},
	)

	out := buf.String()

	wants := []string{
		`"str":"value"`,
		`"bool":true`,
		`"int":1`,
		`"int8":2`,
		`"int16":3`,
		`"int32":4`,
		`"int64":5`,
		`"uint":6`,
		`"uint8":7`,
		`"uint16":8`,
		`"uint32":9`,
		`"uint64":10`,
		`"float32":1.5`,
		`"float64":2.5`,
		`"time":`,
		`"dur":`,
		`"err":"boom"`,
		`"nil":null`,
		`"other":`,
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("output %q is missing %q", out, w)
		}
	}
}

func TestApplyArgs_BadKeys(t *testing.T) {
	t.Parallel()

	t.Run("odd trailing arg captured under BADKEY", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		l.Info(context.Background(), "odd", "k", "v", "stray")

		out := buf.String()
		if !strings.Contains(out, `"!BADKEY":"stray"`) {
			t.Fatalf("output %q does not contain !BADKEY", out)
		}

		if !strings.Contains(out, `"k":"v"`) {
			t.Fatalf("output %q does not contain valid pair", out)
		}
	})

	t.Run("non-string key falls back to indexed BADKEY", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		l.Info(context.Background(), "non-str", 123, "value")

		out := buf.String()
		if !strings.Contains(out, `"!BADKEY_0":"value"`) {
			t.Fatalf("output %q does not contain !BADKEY_0", out)
		}
	})
}
