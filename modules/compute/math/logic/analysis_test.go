package logic

import "testing"

func TestTruthTable(t *testing.T) {
	t.Parallel()

	t.Run("single variable", func(t *testing.T) {
		t.Parallel()

		rows := TruthTable(Var("A"))
		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
	})

	t.Run("two variables", func(t *testing.T) {
		t.Parallel()

		rows := TruthTable(AndF{L: Var("A"), R: Var("B")})
		if len(rows) != 4 {
			t.Fatalf("expected 4 rows, got %d", len(rows))
		}

		trueCount := 0

		for _, row := range rows {
			if row.Result {
				trueCount++
			}
		}

		if trueCount != 1 {
			t.Fatalf("expected 1 true row for AND, got %d", trueCount)
		}
	})

	t.Run("tautology", func(t *testing.T) {
		t.Parallel()

		// A | !A is always true
		rows := TruthTable(OrF{L: Var("A"), R: NotF{F: Var("A")}})

		for _, row := range rows {
			if !row.Result {
				t.Fatalf("expected all true for tautology, got false for %v", row.Assignment)
			}
		}
	})

	t.Run("contradiction", func(t *testing.T) {
		t.Parallel()

		// A & !A is always false
		rows := TruthTable(AndF{L: Var("A"), R: NotF{F: Var("A")}})

		for _, row := range rows {
			if row.Result {
				t.Fatalf("expected all false for contradiction, got true for %v", row.Assignment)
			}
		}
	})

	t.Run("constant true", func(t *testing.T) {
		t.Parallel()

		rows := TruthTable(TrueF{})
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}

		if !rows[0].Result {
			t.Fatal("expected true")
		}
	})

	t.Run("row assignments are populated", func(t *testing.T) {
		t.Parallel()

		rows := TruthTable(Var("A"))

		for _, row := range rows {
			_, exists := row.Assignment["A"]
			if !exists {
				t.Fatal("expected variable A in assignment")
			}
		}
	})
}

func TestEquivalent(t *testing.T) {
	t.Parallel()

	t.Run("identical formulas", func(t *testing.T) {
		t.Parallel()

		if !Equivalent(Var("A"), Var("A")) {
			t.Fatal("expected equivalent")
		}
	})

	t.Run("de morgan equivalence", func(t *testing.T) {
		t.Parallel()

		// !(A & B) ≡ (!A | !B)
		left := NotF{F: AndF{L: Var("A"), R: Var("B")}}
		right := OrF{L: NotF{F: Var("A")}, R: NotF{F: Var("B")}}

		if !Equivalent(left, right) {
			t.Fatal("expected De Morgan equivalence")
		}
	})

	t.Run("non equivalent", func(t *testing.T) {
		t.Parallel()

		if Equivalent(Var("A"), Var("B")) {
			t.Fatal("expected not equivalent")
		}
	})

	t.Run("different variable sets", func(t *testing.T) {
		t.Parallel()

		// A is not equivalent to (A & B)
		if Equivalent(Var("A"), AndF{L: Var("A"), R: Var("B")}) {
			t.Fatal("expected not equivalent with different variables")
		}
	})

	t.Run("constants", func(t *testing.T) {
		t.Parallel()

		if !Equivalent(TrueF{}, TrueF{}) {
			t.Fatal("expected true equivalent to true")
		}
	})

	t.Run("implication equivalence", func(t *testing.T) {
		t.Parallel()

		// A => B ≡ !A | B
		left := ImplF{L: Var("A"), R: Var("B")}
		right := OrF{L: NotF{F: Var("A")}, R: Var("B")}

		if !Equivalent(left, right) {
			t.Fatal("expected implication equivalence")
		}
	})
}

func TestFailCases(t *testing.T) {
	t.Parallel()

	t.Run("tautology has no fail cases", func(t *testing.T) {
		t.Parallel()

		fails := FailCases(OrF{L: Var("A"), R: NotF{F: Var("A")}})
		if len(fails) != 0 {
			t.Fatalf("expected 0 fail cases, got %d", len(fails))
		}
	})

	t.Run("contradiction all fail", func(t *testing.T) {
		t.Parallel()

		fails := FailCases(AndF{L: Var("A"), R: NotF{F: Var("A")}})
		if len(fails) != 2 {
			t.Fatalf("expected 2 fail cases, got %d", len(fails))
		}
	})

	t.Run("single variable", func(t *testing.T) {
		t.Parallel()

		fails := FailCases(Var("A"))
		if len(fails) != 1 {
			t.Fatalf("expected 1 fail case, got %d", len(fails))
		}

		if fails[0]["A"] {
			t.Fatal("expected A=false in fail case")
		}
	})

	t.Run("and has three fail cases", func(t *testing.T) {
		t.Parallel()

		fails := FailCases(AndF{L: Var("A"), R: Var("B")})
		if len(fails) != 3 {
			t.Fatalf("expected 3 fail cases for AND, got %d", len(fails))
		}
	})
}
