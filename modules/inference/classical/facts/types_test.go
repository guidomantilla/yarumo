package facts

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestTypeCompliance(t *testing.T) {
	t.Parallel()

	t.Run("factBase implements FactBase", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBase()
		fb.Assert("A", true)

		got := fb.Len()
		if got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	})

	t.Run("factBase from initial implements FactBase", func(t *testing.T) {
		t.Parallel()

		fb := NewFactBaseFrom(logic.Fact{"A": true})

		got := fb.Len()
		if got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	})
}
