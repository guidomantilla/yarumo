package http

import (
    "errors"
    "testing"
)

func TestNoopRetryIf_AlwaysFalse(t *testing.T) {
    // nil error
    if got := NoopRetryIf(nil); got {
        t.Fatalf("NoopRetryIf(nil) = %v, want false", got)
    }
    // non-nil error
    if got := NoopRetryIf(errors.New("boom")); got {
        t.Fatalf("NoopRetryIf(non-nil) = %v, want false", got)
    }
}

func TestNoopRetryHook_NoOp(t *testing.T) {
    // Should not panic or have any side effects; just invoke with a few values
    NoopRetryHook(0, nil)
    NoopRetryHook(3, errors.New("x"))
}
