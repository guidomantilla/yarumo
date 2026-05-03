package logic

import "testing"

func TestIsSatisfiable(t *testing.T) {
	t.Parallel()

	t.Run("simple variable", func(t *testing.T) {
		t.Parallel()

		if !IsSatisfiable(Var("A")) {
			t.Fatal("expected satisfiable")
		}
	})

	t.Run("tautology", func(t *testing.T) {
		t.Parallel()

		if !IsSatisfiable(OrF{L: Var("A"), R: NotF{F: Var("A")}}) {
			t.Fatal("expected satisfiable")
		}
	})

	t.Run("contradiction", func(t *testing.T) {
		t.Parallel()

		if IsSatisfiable(AndF{L: Var("A"), R: NotF{F: Var("A")}}) {
			t.Fatal("expected unsatisfiable")
		}
	})

	t.Run("true constant", func(t *testing.T) {
		t.Parallel()

		if !IsSatisfiable(TrueF{}) {
			t.Fatal("expected satisfiable")
		}
	})

	t.Run("false constant", func(t *testing.T) {
		t.Parallel()

		if IsSatisfiable(FalseF{}) {
			t.Fatal("expected unsatisfiable")
		}
	})

	t.Run("with registered solver", func(t *testing.T) { //nolint:paralleltest // modifies global satSolver
		original := satSolver

		defer func() { satSolver = original }()

		RegisterSATSolver(func(f Formula) (bool, Fact) {
			return true, Fact{"A": true}
		})

		if !IsSatisfiable(Var("A")) {
			t.Fatal("expected satisfiable via solver")
		}
	})
}

func TestIsContradiction(t *testing.T) {
	t.Parallel()

	t.Run("contradiction", func(t *testing.T) {
		t.Parallel()

		if !IsContradiction(AndF{L: Var("A"), R: NotF{F: Var("A")}}) {
			t.Fatal("expected contradiction")
		}
	})

	t.Run("not contradiction", func(t *testing.T) {
		t.Parallel()

		if IsContradiction(Var("A")) {
			t.Fatal("expected not contradiction")
		}
	})

	t.Run("false constant", func(t *testing.T) {
		t.Parallel()

		if !IsContradiction(FalseF{}) {
			t.Fatal("expected contradiction")
		}
	})
}

func TestIsTautology(t *testing.T) {
	t.Parallel()

	t.Run("tautology", func(t *testing.T) {
		t.Parallel()

		if !IsTautology(OrF{L: Var("A"), R: NotF{F: Var("A")}}) {
			t.Fatal("expected tautology")
		}
	})

	t.Run("not tautology", func(t *testing.T) {
		t.Parallel()

		if IsTautology(Var("A")) {
			t.Fatal("expected not tautology")
		}
	})

	t.Run("true constant", func(t *testing.T) {
		t.Parallel()

		if !IsTautology(TrueF{}) {
			t.Fatal("expected tautology")
		}
	})

	t.Run("implication tautology", func(t *testing.T) {
		t.Parallel()

		// A => A is always true
		if !IsTautology(ImplF{L: Var("A"), R: Var("A")}) {
			t.Fatal("expected tautology")
		}
	})
}
