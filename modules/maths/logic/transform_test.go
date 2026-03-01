package logic

import "testing"

func TestToNNF(t *testing.T) {
	t.Parallel()

	t.Run("variable unchanged", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(Var("A")).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("true unchanged", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(TrueF{}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("false unchanged", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(FalseF{}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("negated true becomes false", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(NotF{F: TrueF{}}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("negated false becomes true", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(NotF{F: FalseF{}}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("double negation eliminated", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(NotF{F: NotF{F: Var("A")}}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("de morgan and", func(t *testing.T) {
		t.Parallel()

		// !(A & B) → (!A | !B)
		got := ToNNF(NotF{F: AndF{L: Var("A"), R: Var("B")}}).String()
		if got != "(!A | !B)" {
			t.Fatalf("expected (!A | !B), got %s", got)
		}
	})

	t.Run("de morgan or", func(t *testing.T) {
		t.Parallel()

		// !(A | B) → (!A & !B)
		got := ToNNF(NotF{F: OrF{L: Var("A"), R: Var("B")}}).String()
		if got != "(!A & !B)" {
			t.Fatalf("expected (!A & !B), got %s", got)
		}
	})

	t.Run("implication eliminated", func(t *testing.T) {
		t.Parallel()

		// A => B → (!A | B)
		got := ToNNF(ImplF{L: Var("A"), R: Var("B")}).String()
		if got != "(!A | B)" {
			t.Fatalf("expected (!A | B), got %s", got)
		}
	})

	t.Run("negated implication", func(t *testing.T) {
		t.Parallel()

		// !(A => B) → (A & !B)
		got := ToNNF(NotF{F: ImplF{L: Var("A"), R: Var("B")}}).String()
		if got != "(A & !B)" {
			t.Fatalf("expected (A & !B), got %s", got)
		}
	})

	t.Run("biconditional eliminated", func(t *testing.T) {
		t.Parallel()

		// A <=> B → (!A | B) & (!B | A)
		got := ToNNF(IffF{L: Var("A"), R: Var("B")}).String()
		if got != "((!A | B) & (!B | A))" {
			t.Fatalf("expected ((!A | B) & (!B | A)), got %s", got)
		}
	})

	t.Run("negated biconditional", func(t *testing.T) {
		t.Parallel()

		// !(A <=> B) → (A & !B) | (!A & B)
		got := ToNNF(NotF{F: IffF{L: Var("A"), R: Var("B")}}).String()
		if got != "((A & !B) | (!A & B))" {
			t.Fatalf("expected ((A & !B) | (!A & B)), got %s", got)
		}
	})

	t.Run("group eliminated", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(GroupF{F: Var("A")}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("and preserved", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(AndF{L: Var("A"), R: Var("B")}).String()
		if got != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", got)
		}
	})

	t.Run("or preserved", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(OrF{L: Var("A"), R: Var("B")}).String()
		if got != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", got)
		}
	})

	t.Run("semantic equivalence", func(t *testing.T) {
		t.Parallel()

		f := NotF{F: ImplF{L: Var("A"), R: AndF{L: Var("B"), R: Var("C")}}}
		nnf := ToNNF(f)

		if !Equivalent(f, nnf) {
			t.Fatal("NNF should be semantically equivalent to original")
		}
	})

	t.Run("unknown formula type passes through", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(customFormula{}).String()
		if got != "custom" {
			t.Fatalf("expected custom, got %s", got)
		}
	})

	t.Run("negated group", func(t *testing.T) {
		t.Parallel()

		got := ToNNF(NotF{F: GroupF{F: Var("A")}}).String()
		if got != "!A" {
			t.Fatalf("expected !A, got %s", got)
		}
	})
}

func TestToCNF(t *testing.T) {
	t.Parallel()

	t.Run("variable unchanged", func(t *testing.T) {
		t.Parallel()

		got := ToCNF(Var("A")).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("or distributed over and", func(t *testing.T) {
		t.Parallel()

		// A | (B & C) → (A | B) & (A | C)
		f := OrF{L: Var("A"), R: AndF{L: Var("B"), R: Var("C")}}

		got := ToCNF(f).String()
		if got != "((A | B) & (A | C))" {
			t.Fatalf("expected ((A | B) & (A | C)), got %s", got)
		}
	})

	t.Run("and preserved", func(t *testing.T) {
		t.Parallel()

		f := AndF{L: Var("A"), R: Var("B")}

		got := ToCNF(f).String()
		if got != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", got)
		}
	})

	t.Run("semantic equivalence", func(t *testing.T) {
		t.Parallel()

		f := ImplF{L: OrF{L: Var("A"), R: Var("B")}, R: Var("C")}
		cnf := ToCNF(f)

		if !Equivalent(f, cnf) {
			t.Fatal("CNF should be semantically equivalent to original")
		}
	})

	t.Run("left and distributed", func(t *testing.T) {
		t.Parallel()

		// (A & B) | C → (A | C) & (B | C)
		f := OrF{L: AndF{L: Var("A"), R: Var("B")}, R: Var("C")}

		got := ToCNF(f).String()
		if got != "((A | C) & (B | C))" {
			t.Fatalf("expected ((A | C) & (B | C)), got %s", got)
		}
	})
}

func TestToDNF(t *testing.T) {
	t.Parallel()

	t.Run("variable unchanged", func(t *testing.T) {
		t.Parallel()

		got := ToDNF(Var("A")).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("and distributed over or", func(t *testing.T) {
		t.Parallel()

		// A & (B | C) → (A & B) | (A & C)
		f := AndF{L: Var("A"), R: OrF{L: Var("B"), R: Var("C")}}

		got := ToDNF(f).String()
		if got != "((A & B) | (A & C))" {
			t.Fatalf("expected ((A & B) | (A & C)), got %s", got)
		}
	})

	t.Run("or preserved", func(t *testing.T) {
		t.Parallel()

		f := OrF{L: Var("A"), R: Var("B")}

		got := ToDNF(f).String()
		if got != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", got)
		}
	})

	t.Run("semantic equivalence", func(t *testing.T) {
		t.Parallel()

		f := AndF{L: Var("A"), R: ImplF{L: Var("B"), R: Var("C")}}
		dnf := ToDNF(f)

		if !Equivalent(f, dnf) {
			t.Fatal("DNF should be semantically equivalent to original")
		}
	})

	t.Run("left or distributed", func(t *testing.T) {
		t.Parallel()

		// (A | B) & C → (A & C) | (B & C)
		f := AndF{L: OrF{L: Var("A"), R: Var("B")}, R: Var("C")}

		got := ToDNF(f).String()
		if got != "((A & C) | (B & C))" {
			t.Fatalf("expected ((A & C) | (B & C)), got %s", got)
		}
	})
}
