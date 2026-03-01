package entailment

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestEntails(t *testing.T) {
	t.Parallel()

	t.Run("simple valid entailment", func(t *testing.T) {
		t.Parallel()

		// A, A => B ⊨ B
		premises := []logic.Formula{
			logic.Var("A"),
			logic.ImplF{L: logic.Var("A"), R: logic.Var("B")},
		}

		result := Entails(premises, logic.Var("B"))

		if !result {
			t.Fatal("expected entailment to hold")
		}
	})

	t.Run("simple invalid entailment", func(t *testing.T) {
		t.Parallel()

		// A ⊭ B
		premises := []logic.Formula{logic.Var("A")}

		result := Entails(premises, logic.Var("B"))

		if result {
			t.Fatal("expected entailment to fail")
		}
	})

	t.Run("empty premises entail tautology", func(t *testing.T) {
		t.Parallel()

		f := logic.OrF{L: logic.Var("A"), R: logic.NotF{F: logic.Var("A")}}

		result := Entails(nil, f)

		if !result {
			t.Fatal("expected empty premises to entail tautology")
		}
	})

	t.Run("empty premises do not entail contingent", func(t *testing.T) {
		t.Parallel()

		result := Entails(nil, logic.Var("A"))

		if result {
			t.Fatal("expected empty premises not to entail contingent formula")
		}
	})

	t.Run("transitive chain", func(t *testing.T) {
		t.Parallel()

		// A, A => B, B => C ⊨ C
		premises := []logic.Formula{
			logic.Var("A"),
			logic.ImplF{L: logic.Var("A"), R: logic.Var("B")},
			logic.ImplF{L: logic.Var("B"), R: logic.Var("C")},
		}

		result := Entails(premises, logic.Var("C"))

		if !result {
			t.Fatal("expected transitive entailment to hold")
		}
	})

	t.Run("contradiction entails anything", func(t *testing.T) {
		t.Parallel()

		// A, !A ⊨ B
		premises := []logic.Formula{
			logic.Var("A"),
			logic.NotF{F: logic.Var("A")},
		}

		result := Entails(premises, logic.Var("B"))

		if !result {
			t.Fatal("expected contradiction to entail anything")
		}
	})
}

func TestEntailsWithCounterModel(t *testing.T) {
	t.Parallel()

	t.Run("valid entailment no countermodel", func(t *testing.T) {
		t.Parallel()

		premises := []logic.Formula{
			logic.Var("A"),
			logic.ImplF{L: logic.Var("A"), R: logic.Var("B")},
		}

		entailed, counter := EntailsWithCounterModel(premises, logic.Var("B"))

		if !entailed {
			t.Fatal("expected entailment to hold")
		}

		if counter != nil {
			t.Fatal("expected nil countermodel")
		}
	})

	t.Run("invalid entailment with countermodel", func(t *testing.T) {
		t.Parallel()

		// A ⊭ B — countermodel: A=true, B=false
		premises := []logic.Formula{logic.Var("A")}

		entailed, counter := EntailsWithCounterModel(premises, logic.Var("B"))

		if entailed {
			t.Fatal("expected entailment to fail")
		}

		if counter == nil {
			t.Fatal("expected non-nil countermodel")
		}

		// In countermodel: A must be true (premise), B must be false (conclusion fails)
		if !counter["A"] {
			t.Fatalf("expected A=true in countermodel, got %v", counter["A"])
		}

		if counter["B"] {
			t.Fatalf("expected B=false in countermodel, got %v", counter["B"])
		}
	})

	t.Run("empty premises countermodel", func(t *testing.T) {
		t.Parallel()

		entailed, counter := EntailsWithCounterModel(nil, logic.Var("A"))

		if entailed {
			t.Fatal("expected entailment to fail")
		}

		if counter == nil {
			t.Fatal("expected non-nil countermodel")
		}
	})

	t.Run("countermodel satisfies premises and falsifies conclusion", func(t *testing.T) {
		t.Parallel()

		// A & B ⊭ C
		premises := []logic.Formula{
			logic.AndF{L: logic.Var("A"), R: logic.Var("B")},
		}
		conclusion := logic.Var("C")

		entailed, counter := EntailsWithCounterModel(premises, conclusion)

		if entailed {
			t.Fatal("expected entailment to fail")
		}

		// Verify countermodel: premises true, conclusion false
		for _, p := range premises {
			if !p.Eval(counter) {
				t.Fatalf("countermodel does not satisfy premise: %s", p.String())
			}
		}

		if conclusion.Eval(counter) {
			t.Fatal("countermodel satisfies conclusion, but should not")
		}
	})
}
