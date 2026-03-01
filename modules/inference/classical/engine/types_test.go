package engine

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

func TestTypeCompliance(t *testing.T) {
	t.Parallel()

	t.Run("engine implements Engine", func(t *testing.T) {
		t.Parallel()

		e := NewEngine()
		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r})
		if result.Steps == 0 {
			t.Fatal("expected at least one step")
		}
	})
}

func TestStrategy(t *testing.T) {
	t.Parallel()

	t.Run("PriorityOrder is zero", func(t *testing.T) {
		t.Parallel()

		if PriorityOrder != 0 {
			t.Fatalf("expected 0, got %d", PriorityOrder)
		}
	})

	t.Run("FirstMatch is one", func(t *testing.T) {
		t.Parallel()

		if FirstMatch != 1 {
			t.Fatalf("expected 1, got %d", FirstMatch)
		}
	})
}
