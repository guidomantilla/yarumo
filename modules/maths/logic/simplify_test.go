package logic

import "testing"

func TestSimplify_constants(t *testing.T) {
	t.Parallel()

	t.Run("variable unchanged", func(t *testing.T) {
		t.Parallel()

		got := Simplify(Var("A")).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("true unchanged", func(t *testing.T) {
		t.Parallel()

		got := Simplify(TrueF{}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("false unchanged", func(t *testing.T) {
		t.Parallel()

		got := Simplify(FalseF{}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("not true", func(t *testing.T) {
		t.Parallel()

		got := Simplify(NotF{F: TrueF{}}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("not false", func(t *testing.T) {
		t.Parallel()

		got := Simplify(NotF{F: FalseF{}}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("not variable unchanged", func(t *testing.T) {
		t.Parallel()

		got := Simplify(NotF{F: Var("A")}).String()
		if got != "!A" {
			t.Fatalf("expected !A, got %s", got)
		}
	})

	t.Run("double negation", func(t *testing.T) {
		t.Parallel()

		got := Simplify(NotF{F: NotF{F: Var("A")}}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("group eliminated", func(t *testing.T) {
		t.Parallel()

		got := Simplify(GroupF{F: Var("A")}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})
}

func TestSimplify_rules(t *testing.T) {
	t.Parallel()

	t.Run("and true left", func(t *testing.T) {
		t.Parallel()

		got := Simplify(AndF{L: TrueF{}, R: Var("A")}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("and true right", func(t *testing.T) {
		t.Parallel()

		got := Simplify(AndF{L: Var("A"), R: TrueF{}}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("and false left", func(t *testing.T) {
		t.Parallel()

		got := Simplify(AndF{L: FalseF{}, R: Var("A")}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("and false right", func(t *testing.T) {
		t.Parallel()

		got := Simplify(AndF{L: Var("A"), R: FalseF{}}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("or true left", func(t *testing.T) {
		t.Parallel()

		got := Simplify(OrF{L: TrueF{}, R: Var("A")}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("or true right", func(t *testing.T) {
		t.Parallel()

		got := Simplify(OrF{L: Var("A"), R: TrueF{}}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("or false left", func(t *testing.T) {
		t.Parallel()

		got := Simplify(OrF{L: FalseF{}, R: Var("A")}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("or false right", func(t *testing.T) {
		t.Parallel()

		got := Simplify(OrF{L: Var("A"), R: FalseF{}}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("and idempotent", func(t *testing.T) {
		t.Parallel()

		got := Simplify(AndF{L: Var("A"), R: Var("A")}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("or idempotent", func(t *testing.T) {
		t.Parallel()

		got := Simplify(OrF{L: Var("A"), R: Var("A")}).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("and complement", func(t *testing.T) {
		t.Parallel()

		got := Simplify(AndF{L: Var("A"), R: NotF{F: Var("A")}}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("and complement reversed", func(t *testing.T) {
		t.Parallel()

		got := Simplify(AndF{L: NotF{F: Var("A")}, R: Var("A")}).String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("or complement", func(t *testing.T) {
		t.Parallel()

		got := Simplify(OrF{L: Var("A"), R: NotF{F: Var("A")}}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("or complement reversed", func(t *testing.T) {
		t.Parallel()

		got := Simplify(OrF{L: NotF{F: Var("A")}, R: Var("A")}).String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("implication eliminated", func(t *testing.T) {
		t.Parallel()

		got := Simplify(ImplF{L: Var("A"), R: Var("B")}).String()

		expected := "(!A | B)"
		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})

	t.Run("biconditional eliminated", func(t *testing.T) {
		t.Parallel()

		result := Simplify(IffF{L: Var("A"), R: Var("B")})

		if !Equivalent(result, IffF{L: Var("A"), R: Var("B")}) {
			t.Fatal("simplified biconditional should be equivalent to original")
		}
	})

	t.Run("nested simplification", func(t *testing.T) {
		t.Parallel()

		// (A & true) | (false & B) → A
		f := OrF{
			L: AndF{L: Var("A"), R: TrueF{}},
			R: AndF{L: FalseF{}, R: Var("B")},
		}

		got := Simplify(f).String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})
}
