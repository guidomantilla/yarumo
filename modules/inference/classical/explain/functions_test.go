package explain

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestNewTrace(t *testing.T) {
	t.Parallel()

	t.Run("creates empty trace", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace()
		if len(tr.Steps) != 0 {
			t.Fatalf("expected 0 steps, got %d", len(tr.Steps))
		}

		if tr.Goal != "" {
			t.Fatalf("expected empty goal, got %q", tr.Goal)
		}
	})
}

func TestNewGoalTrace(t *testing.T) {
	t.Parallel()

	t.Run("creates trace with goal", func(t *testing.T) {
		t.Parallel()

		tr := NewGoalTrace("X")
		if tr.Goal != "X" {
			t.Fatalf("expected goal X, got %q", tr.Goal)
		}

		if len(tr.Steps) != 0 {
			t.Fatalf("expected 0 steps, got %d", len(tr.Steps))
		}
	})
}

func TestTrace_AddStep(t *testing.T) {
	t.Parallel()

	t.Run("appends step to empty trace", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace()
		step := Step{Number: 1, RuleName: "r1"}
		tr = tr.AddStep(step)

		if len(tr.Steps) != 1 {
			t.Fatalf("expected 1 step, got %d", len(tr.Steps))
		}

		if tr.Steps[0].RuleName != "r1" {
			t.Fatalf("expected r1, got %s", tr.Steps[0].RuleName)
		}
	})

	t.Run("appends multiple steps", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace()
		tr = tr.AddStep(Step{Number: 1, RuleName: "r1"})
		tr = tr.AddStep(Step{Number: 2, RuleName: "r2"})

		if len(tr.Steps) != 2 {
			t.Fatalf("expected 2 steps, got %d", len(tr.Steps))
		}

		if tr.Steps[1].RuleName != "r2" {
			t.Fatalf("expected r2, got %s", tr.Steps[1].RuleName)
		}
	})
}

func TestProvenanceOf_asserted(t *testing.T) {
	t.Parallel()

	t.Run("initial facts are asserted", func(t *testing.T) {
		t.Parallel()

		initial := logic.Fact{"A": true}
		tr := NewTrace()
		provs := ProvenanceOf(tr, initial)

		if len(provs) != 1 {
			t.Fatalf("expected 1 provenance, got %d", len(provs))
		}

		if provs[0].Origin != Asserted {
			t.Fatal("expected asserted origin")
		}

		if provs[0].Variable != "A" {
			t.Fatalf("expected variable A, got %s", provs[0].Variable)
		}

		if !provs[0].Value {
			t.Fatal("expected value true")
		}

		if provs[0].RuleName != "" {
			t.Fatalf("expected empty rule name, got %s", provs[0].RuleName)
		}

		if provs[0].Step != 0 {
			t.Fatalf("expected step 0, got %d", provs[0].Step)
		}
	})
}

func TestProvenanceOf_derived(t *testing.T) {
	t.Parallel()

	t.Run("derived facts record rule and step", func(t *testing.T) {
		t.Parallel()

		initial := logic.Fact{"A": true}
		tr := NewTrace()
		tr = tr.AddStep(Step{
			Number:   1,
			RuleName: "r1",
			Produced: map[logic.Var]bool{"B": true},
		})
		provs := ProvenanceOf(tr, initial)

		if len(provs) != 2 {
			t.Fatalf("expected 2 provenances, got %d", len(provs))
		}

		var derived *Provenance

		for i := range provs {
			if provs[i].Variable == "B" {
				derived = &provs[i]
			}
		}

		if derived == nil {
			t.Fatal("expected provenance for B")
		}

		if derived.Origin != Derived {
			t.Fatal("expected derived origin")
		}

		if derived.RuleName != "r1" {
			t.Fatalf("expected rule r1, got %s", derived.RuleName)
		}

		if derived.Step != 1 {
			t.Fatalf("expected step 1, got %d", derived.Step)
		}
	})
}

func TestProvenanceOf_duplicates(t *testing.T) {
	t.Parallel()

	t.Run("duplicate derived facts only recorded once", func(t *testing.T) {
		t.Parallel()

		initial := logic.Fact{}
		tr := NewTrace()
		tr = tr.AddStep(Step{
			Number:   1,
			RuleName: "r1",
			Produced: map[logic.Var]bool{"B": true},
		})
		tr = tr.AddStep(Step{
			Number:   2,
			RuleName: "r2",
			Produced: map[logic.Var]bool{"B": true},
		})
		provs := ProvenanceOf(tr, initial)

		count := 0

		for _, p := range provs {
			if p.Variable == "B" {
				count++
			}
		}

		if count != 1 {
			t.Fatalf("expected 1 provenance for B, got %d", count)
		}
	})
}

func TestProvenanceOf_empty(t *testing.T) {
	t.Parallel()

	t.Run("empty trace and empty initial", func(t *testing.T) {
		t.Parallel()

		provs := ProvenanceOf(NewTrace(), logic.Fact{})
		if len(provs) != 0 {
			t.Fatalf("expected 0 provenances, got %d", len(provs))
		}
	})
}
