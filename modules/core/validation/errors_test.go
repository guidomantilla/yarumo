package validation

import (
	"testing"
)

func TestErrors_ErrorMethods(t *testing.T) {
	t.Parallel()

	t.Run("engine error", func(t *testing.T) {
		t.Parallel()

		err := ErrEngine(ErrBadRule)

		msg := err.Error()
		if msg == "" {
			t.Fatalf("expected non-empty message")
		}
	})

	t.Run("load error", func(t *testing.T) {
		t.Parallel()

		err := ErrLoad(ErrDataNil)

		msg := err.Error()
		if msg == "" {
			t.Fatalf("expected non-empty message")
		}
	})
}
