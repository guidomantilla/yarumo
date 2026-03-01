package rules

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestTypeCompliance(t *testing.T) {
	t.Parallel()

	t.Run("rule implements Rule", func(t *testing.T) {
		t.Parallel()

		r := NewRule("test", logic.Var("A"), map[logic.Var]bool{"B": true})

		got := r.Name()
		if got != "test" {
			t.Fatalf("expected test, got %s", got)
		}
	})
}
