package logic

import "testing"

func TestFormat(t *testing.T) {
	t.Parallel()

	t.Run("variable", func(t *testing.T) {
		t.Parallel()

		got := Format(Var("A"))
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("true", func(t *testing.T) {
		t.Parallel()

		got := Format(TrueF{})
		if got != "⊤" {
			t.Fatalf("expected ⊤, got %s", got)
		}
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()

		got := Format(FalseF{})
		if got != "⊥" {
			t.Fatalf("expected ⊥, got %s", got)
		}
	})

	t.Run("not", func(t *testing.T) {
		t.Parallel()

		got := Format(NotF{F: Var("A")})
		if got != "¬A" {
			t.Fatalf("expected ¬A, got %s", got)
		}
	})

	t.Run("and", func(t *testing.T) {
		t.Parallel()

		got := Format(AndF{L: Var("A"), R: Var("B")})
		if got != "(A ∧ B)" {
			t.Fatalf("expected (A ∧ B), got %s", got)
		}
	})

	t.Run("or", func(t *testing.T) {
		t.Parallel()

		got := Format(OrF{L: Var("A"), R: Var("B")})
		if got != "(A ∨ B)" {
			t.Fatalf("expected (A ∨ B), got %s", got)
		}
	})

	t.Run("implies", func(t *testing.T) {
		t.Parallel()

		got := Format(ImplF{L: Var("A"), R: Var("B")})
		if got != "(A → B)" {
			t.Fatalf("expected (A → B), got %s", got)
		}
	})

	t.Run("iff", func(t *testing.T) {
		t.Parallel()

		got := Format(IffF{L: Var("A"), R: Var("B")})
		if got != "(A ↔ B)" {
			t.Fatalf("expected (A ↔ B), got %s", got)
		}
	})

	t.Run("group", func(t *testing.T) {
		t.Parallel()

		got := Format(GroupF{F: Var("A")})
		if got != "(A)" {
			t.Fatalf("expected (A), got %s", got)
		}
	})

	t.Run("nested", func(t *testing.T) {
		t.Parallel()

		f := ImplF{L: AndF{L: Var("A"), R: Var("B")}, R: NotF{F: Var("C")}}

		got := Format(f)

		expected := "((A ∧ B) → ¬C)"
		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})

	t.Run("unknown formula type falls back to String", func(t *testing.T) {
		t.Parallel()

		got := Format(customFormula{})
		if got != "custom" {
			t.Fatalf("expected custom, got %s", got)
		}
	})
}

type customFormula struct{}

func (c customFormula) String() string { return "custom" }

func (c customFormula) Eval(_ Fact) bool { return false }

func (c customFormula) Vars() []Var { return nil }
