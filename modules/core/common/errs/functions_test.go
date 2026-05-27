package errs

import (
	"errors"
	"fmt"
	"testing"
)

// --- test helper types ---

type errA struct{ msg string }

func (e *errA) Error() string { return e.msg }

type errB struct{ inner error }

func (e *errB) Error() string { return "errB: " + fmt.Sprint(e.inner) }

func (e *errB) Unwrap() error { return e.inner }

type errMulti struct{ inners []error }

func (e *errMulti) Error() string { return "errMulti" }

func (e *errMulti) Unwrap() []error { return e.inners }

type cyc struct{ inner error }

func (e *cyc) Error() string { return "cyc" }

func (e *cyc) Unwrap() error { return e.inner }

// errTyped embeds TypedError to test ErrorType propagation in AsErrorInfo.
type errTyped struct {
	TypedError
}

func (e *errTyped) Error() string {
	return fmt.Sprintf("typed(%s): %s", e.Type, e.Err)
}

// --- tests ---

func TestAs(t *testing.T) {
	t.Parallel()

	t.Run("matches concrete error type", func(t *testing.T) {
		t.Parallel()

		original := &errA{msg: "hello"}
		wrapped := fmt.Errorf("wrap: %w", original)

		got, ok := As[*errA](wrapped)
		if !ok {
			t.Fatal("expected match")
		}

		if got.msg != "hello" {
			t.Fatalf("got msg %q, want %q", got.msg, "hello")
		}
	})

	t.Run("does not match wrong type", func(t *testing.T) {
		t.Parallel()

		err := &errA{msg: "hello"}

		_, ok := As[*errB](err)
		if ok {
			t.Fatal("expected no match")
		}
	})

	t.Run("nil error returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := As[*errA](nil)
		if ok {
			t.Fatal("expected no match for nil")
		}
	})
}

func TestMatch(t *testing.T) {
	t.Parallel()

	t.Run("matches by type", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("wrap: %w", &errA{msg: "x"})
		if !Match[*errA](err) {
			t.Fatal("expected type match")
		}
	})

	t.Run("matches by value", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("sentinel")
		err := fmt.Errorf("wrap: %w", sentinel)

		if !Match[*errB](err, sentinel) {
			t.Fatal("expected value match")
		}
	})

	t.Run("no match returns false", func(t *testing.T) {
		t.Parallel()

		err := errors.New("plain")
		if Match[*errA](err) {
			t.Fatal("expected no match")
		}
	})

	t.Run("nil error returns false", func(t *testing.T) {
		t.Parallel()

		if Match[*errA](nil) {
			t.Fatal("expected no match for nil")
		}
	})

	t.Run("matches value without type match", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("specific")
		err := fmt.Errorf("wrap: %w", sentinel)

		if !Match[*errA](err, sentinel) {
			t.Fatal("expected value match even without type match")
		}
	})
}

func TestWrap(t *testing.T) {
	t.Parallel()

	t.Run("joins multiple errors", func(t *testing.T) {
		t.Parallel()

		e1 := errors.New("a")
		e2 := errors.New("b")

		joined := Wrap(e1, e2)
		if joined == nil {
			t.Fatal("expected non-nil error")
		}

		if !errors.Is(joined, e1) {
			t.Fatal("joined does not contain e1")
		}

		if !errors.Is(joined, e2) {
			t.Fatal("joined does not contain e2")
		}
	})

	t.Run("single error", func(t *testing.T) {
		t.Parallel()

		e := errors.New("only")

		joined := Wrap(e)
		if joined == nil {
			t.Fatal("expected non-nil error")
		}
	})

	t.Run("all nil returns nil", func(t *testing.T) {
		t.Parallel()

		if Wrap(nil, nil) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("no args returns nil", func(t *testing.T) {
		t.Parallel()

		if Wrap() != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("mixed nil and non-nil", func(t *testing.T) {
		t.Parallel()

		e := errors.New("real")

		joined := Wrap(nil, e, nil)
		if joined == nil {
			t.Fatal("expected non-nil")
		}

		if !errors.Is(joined, e) {
			t.Fatal("joined does not contain the error")
		}
	})
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	t.Run("leaf error returns itself", func(t *testing.T) {
		t.Parallel()

		e := errors.New("leaf")
		got := Unwrap(e)

		if len(got) != 1 || !errors.Is(got[0], e) {
			t.Fatalf("expected [leaf], got %v", got)
		}
	})

	t.Run("single-unwrap chain", func(t *testing.T) {
		t.Parallel()

		leaf := &errA{msg: "bottom"}
		chain := &errB{inner: leaf}

		got := Unwrap(chain)
		if len(got) != 1 || !errors.Is(got[0], leaf) {
			t.Fatalf("expected [bottom], got %v", got)
		}
	})

	t.Run("multi-unwrap", func(t *testing.T) {
		t.Parallel()

		a := &errA{msg: "a"}
		b := &errA{msg: "b"}
		multi := &errMulti{inners: []error{a, b}}

		got := Unwrap(multi)
		if len(got) != 2 {
			t.Fatalf("expected 2 leaves, got %d", len(got))
		}

		if !errors.Is(got[0], a) || !errors.Is(got[1], b) {
			t.Fatalf("unexpected leaves: %v", got)
		}
	})

	t.Run("nested single inside multi", func(t *testing.T) {
		t.Parallel()

		leaf := &errA{msg: "deep"}
		chain := &errB{inner: leaf}
		multi := &errMulti{inners: []error{chain}}

		got := Unwrap(multi)
		if len(got) != 1 || !errors.Is(got[0], leaf) {
			t.Fatalf("expected [deep], got %v", got)
		}
	})

	t.Run("cycle detection", func(t *testing.T) {
		t.Parallel()

		c := &cyc{}
		c.inner = c

		got := Unwrap(c)
		if len(got) != 0 {
			t.Fatalf("expected empty (cycle), got %v", got)
		}
	})

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()

		got := Unwrap(nil)
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("multi-unwrap with nil inner", func(t *testing.T) {
		t.Parallel()

		a := &errA{msg: "a"}
		multi := &errMulti{inners: []error{a, nil}}

		got := Unwrap(multi)
		if len(got) != 1 || !errors.Is(got[0], a) {
			t.Fatalf("expected [a], got %v", got)
		}
	})
}

func TestErrorMessages(t *testing.T) {
	t.Parallel()

	t.Run("collects leaf messages", func(t *testing.T) {
		t.Parallel()

		multi := &errMulti{inners: []error{
			&errA{msg: "first"},
			&errA{msg: "second"},
		}}

		got := ErrorMessages(multi)
		if len(got) != 2 || got[0] != "first" || got[1] != "second" {
			t.Fatalf("got %v, want [first second]", got)
		}
	})

	t.Run("single error", func(t *testing.T) {
		t.Parallel()

		got := ErrorMessages(errors.New("solo"))
		if len(got) != 1 || got[0] != "solo" {
			t.Fatalf("got %v, want [solo]", got)
		}
	})

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()

		got := ErrorMessages(nil)
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func TestHasErrorMessage(t *testing.T) {
	t.Parallel()

	t.Run("finds substring in leaf", func(t *testing.T) {
		t.Parallel()

		err := &errMulti{inners: []error{
			&errA{msg: "connection timeout"},
			&errA{msg: "dial failed"},
		}}

		if !HasErrorMessage(err, "timeout") {
			t.Fatal("expected true")
		}
	})

	t.Run("no match returns false", func(t *testing.T) {
		t.Parallel()

		err := errors.New("something else")
		if HasErrorMessage(err, "timeout") {
			t.Fatal("expected false")
		}
	})

	t.Run("nil error returns false", func(t *testing.T) {
		t.Parallel()

		if HasErrorMessage(nil, "anything") {
			t.Fatal("expected false")
		}
	})

	t.Run("empty substring matches any error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("any message")
		if !HasErrorMessage(err, "") {
			t.Fatal("expected true for empty substring")
		}
	})
}

func TestAsErrorInfo(t *testing.T) {
	t.Parallel()

	t.Run("groups leaves by TypedError type", func(t *testing.T) {
		t.Parallel()

		err := &TypedError{
			Type: "http-request",
			Err: errors.Join(
				errors.New("timeout"),
				errors.New("dial failed"),
			),
		}

		got := AsErrorInfo(err)
		if len(got) != 1 {
			t.Fatalf("expected 1 group, got %d", len(got))
		}

		if got[0].Type != "http-request" {
			t.Fatalf("Type = %q, want %q", got[0].Type, "http-request")
		}

		if len(got[0].Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(got[0].Messages))
		}

		if got[0].Messages[0] != "timeout" || got[0].Messages[1] != "dial failed" {
			t.Fatalf("unexpected messages: %v", got[0].Messages)
		}
	})

	t.Run("propagates type through embedded TypedError", func(t *testing.T) {
		t.Parallel()

		err := &errTyped{
			TypedError: TypedError{
				Type: "service",
				Err:  errors.New("down"),
			},
		}

		got := AsErrorInfo(err)
		if len(got) != 1 {
			t.Fatalf("expected 1 group, got %d", len(got))
		}

		if got[0].Type != "service" {
			t.Fatalf("Type = %q, want %q", got[0].Type, "service")
		}

		if len(got[0].Messages) != 1 || got[0].Messages[0] != "down" {
			t.Fatalf("unexpected messages: %v", got[0].Messages)
		}
	})

	t.Run("multiple typed errors produce multiple groups", func(t *testing.T) {
		t.Parallel()

		httpErr := &TypedError{Type: "http", Err: errors.New("timeout")}
		dbErr := &TypedError{Type: "db", Err: errors.New("conn refused")}
		combined := errors.Join(httpErr, dbErr)

		got := AsErrorInfo(combined)
		if len(got) != 2 {
			t.Fatalf("expected 2 groups, got %d", len(got))
		}

		if got[0].Type != "http" || got[0].Messages[0] != "timeout" {
			t.Fatalf("first group: %+v", got[0])
		}

		if got[1].Type != "db" || got[1].Messages[0] != "conn refused" {
			t.Fatalf("second group: %+v", got[1])
		}
	})

	t.Run("untyped errors fall back to reflect type", func(t *testing.T) {
		t.Parallel()

		err := &errA{msg: "plain"}
		got := AsErrorInfo(err)

		if len(got) != 1 {
			t.Fatalf("expected 1 group, got %d", len(got))
		}

		if got[0].Type != "*errs.errA" {
			t.Fatalf("Type = %q, want %q", got[0].Type, "*errs.errA")
		}

		if got[0].Messages[0] != "plain" {
			t.Fatalf("Messages[0] = %q, want %q", got[0].Messages[0], "plain")
		}
	})

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()

		got := AsErrorInfo(nil)
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("nested typed error overrides parent type", func(t *testing.T) {
		t.Parallel()

		inner := &TypedError{Type: "specific", Err: errors.New("detail")}
		outer := &TypedError{Type: "general", Err: inner}

		got := AsErrorInfo(outer)
		if len(got) != 1 {
			t.Fatalf("expected 1 group, got %d", len(got))
		}

		if got[0].Type != "specific" {
			t.Fatalf("Type = %q, want %q (innermost should win)", got[0].Type, "specific")
		}
	})

	t.Run("cycle in error tree does not loop", func(t *testing.T) {
		t.Parallel()

		c := &cyc{}
		c.inner = c

		got := AsErrorInfo(c)
		if len(got) != 0 {
			t.Fatalf("expected empty for cycle, got %v", got)
		}
	})

	t.Run("leafTypeName uses currentType when provided", func(t *testing.T) {
		t.Parallel()

		got := leafTypeName(errors.New("x"), "custom-type")
		if got != "custom-type" {
			t.Fatalf("leafTypeName = %q, want %q", got, "custom-type")
		}
	})

	t.Run("leafTypeName falls back to reflect type", func(t *testing.T) {
		t.Parallel()

		got := leafTypeName(&errA{msg: "x"}, "")
		if got != "*errs.errA" {
			t.Fatalf("leafTypeName = %q, want %q", got, "*errs.errA")
		}
	})

	t.Run("multi-unwrap with nil inner skips nil", func(t *testing.T) {
		t.Parallel()

		a := &errA{msg: "valid"}
		multi := &errMulti{inners: []error{a, nil}}

		got := AsErrorInfo(multi)
		if len(got) != 1 {
			t.Fatalf("expected 1 group, got %d", len(got))
		}

		if got[0].Messages[0] != "valid" {
			t.Fatalf("Messages[0] = %q, want %q", got[0].Messages[0], "valid")
		}
	})
}
