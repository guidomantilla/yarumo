package messaging

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestSilentErrorHandler_OptsOutOfLogging(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithErrorHandler(SilentErrorHandler))

	want := reflect.ValueOf(SilentErrorHandler).Pointer()
	got := reflect.ValueOf(opts.errorHandler).Pointer()
	if got != want {
		t.Fatalf("expected SilentErrorHandler installed, got different function")
	}

	// Smoke-test the silent hook actually does nothing: should not
	// panic and should not write anywhere observable.
	SilentErrorHandler(context.Background(), nil, errors.New("test"))
}

func TestDefaultErrorHandler_DoesNotPanicOnNilMsg(t *testing.T) {
	t.Parallel()

	// Smoke-test: handler must tolerate nil msg + non-nil err and not
	// panic. Output goes through common/log's global slot — invisible
	// to this test but the call must complete cleanly.
	DefaultErrorHandler(context.Background(), nil, errors.New("smoke"))
}
