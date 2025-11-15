package errs

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// errA is a simple custom error type
type errA struct{ msg string }

func (e *errA) Error() string { return e.msg }

// errB wraps a single inner error and exposes Unwrap() error
type errB struct{ inner error }

func (e *errB) Error() string { return "errB: " + fmt.Sprint(e.inner) }

func (e *errB) Unwrap() error { return e.inner }

// errMulti wraps multiple inner errors and exposes Unwrap() []error
type errMulti struct{ inners []error }

func (e *errMulti) Error() string { return "errMulti" }

func (e *errMulti) Unwrap() []error { return e.inners }

// cyc is used to validate that Unwrap() doesn't loop forever on cycles
type cyc struct{ inner error }

func (e *cyc) Error() string { return "cyc" }

func (e *cyc) Unwrap() error { return e.inner }

func TestErrsFunctions(t *testing.T) {
	// --- As[T] ---
	baseA := &errA{msg: "A"}
	wrapped := fmt.Errorf("lvl2: %w", baseA)

	if got, ok := As[*errA](wrapped); !ok || got == nil || got.msg != "A" {
		t.Fatalf("As[*errA] failed: got=%v ok=%v", got, ok)
	}
	if _, ok := As[*errB](wrapped); ok {
		t.Fatalf("As[*errB] expected false")
	}

	// --- Match[T] ---
	sentinel := errors.New("sentinel")
	// match by type (errors.As)
	matchType := &errB{inner: baseA}
	if !Match[*errB](fmt.Errorf("wrap: %w", matchType)) {
		t.Fatalf("Match by type failed")
	}
	// match by value (errors.Is); pick T that does not match to force value-path
	if !Match[*errA](fmt.Errorf("with sentinel: %w", sentinel), sentinel) {
		t.Fatalf("Match by value failed")
	}
	// no match
	if Match[*errB](errors.New("other")) {
		t.Fatalf("Match expected false")
	}

	// --- Wrap ---
	if got := Wrap(); got != nil {
		t.Fatalf("Wrap() with no errors must be nil, got %v", got)
	}
	e1 := errors.New("e1")
	e2 := error(nil)
	e3 := errors.New("e3")
	joined := Wrap(e1, e2, e3)
	if joined == nil {
		t.Fatalf("Wrap with non-nil errors must return non-nil")
	}

	// --- Unwrap ---
	// 1) Simple chain Unwrap() error
	chain := &errB{inner: fmt.Errorf("leaf: %w", e1)}
	got1 := Unwrap(chain)
	// Expect to collect the final non-unwrapping errors: e1 and the fmt.Errorf("leaf: %w", e1)
	// The default case appends when it finds an error that doesn't implement Unwrap
	if len(got1) == 0 {
		t.Fatalf("Unwrap(chain) returned empty slice")
	}

	// 2) Multiple via Unwrap() []error
	multi := &errMulti{inners: []error{e1, e3}}
	got2 := Unwrap(multi)
	if !reflect.DeepEqual(got2, []error{e1, e3}) {
		t.Fatalf("Unwrap([]error) mismatch: %v", got2)
	}

	// 3) Cycle safety (should terminate and not panic)
	c1 := &cyc{}
	c2 := &cyc{inner: c1}
	c1.inner = c2
	gotCycle := Unwrap(c1)
	if len(gotCycle) != 0 { // no leaf errors, so result should be empty
		t.Fatalf("Unwrap(cycle) expected empty, got %v", gotCycle)
	}

	// 4) Nil errors are ignored
	gotNil := Unwrap(nil)
	if len(gotNil) != 0 {
		t.Fatalf("Unwrap(nil) expected empty, got %v", gotNil)
	}

	// --- ErrorMessages ---
	msgs := ErrorMessages(Wrap(e1, e3))
	if !reflect.DeepEqual(msgs, []string{"e1", "e3"}) {
		t.Fatalf("ErrorMessages mismatch: %v", msgs)
	}

	// --- HasErrorMessage ---
	if !HasErrorMessage(Wrap(e1, e3), "e3") {
		t.Fatalf("HasErrorMessage expected true for substring present")
	}
	if HasErrorMessage(Wrap(e1, e3), "missing") {
		t.Fatalf("HasErrorMessage expected false for substring absent")
	}

	// --- AsErrorInfo ---
	infos := AsErrorInfo(Wrap(baseA, e3))
	if len(infos) != 2 {
		t.Fatalf("AsErrorInfo length mismatch: %v", infos)
	}
	// Validate that type names and messages are filled
	if infos[0].Type == "" || infos[0].Message == "" {
		t.Fatalf("AsErrorInfo[0] has empty fields: %v", infos[0])
	}
	if infos[1].Type == "" || infos[1].Message != "e3" {
		t.Fatalf("AsErrorInfo[1] invalid: %v", infos[1])
	}
}
