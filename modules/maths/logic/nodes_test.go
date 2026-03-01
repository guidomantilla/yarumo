package logic

import "testing"

func TestTypeCompliance(t *testing.T) {
	t.Parallel()

	t.Run("Var implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = Var("A")

		got := f.String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("TrueF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = TrueF{}

		got := f.String()
		if got != "true" {
			t.Fatalf("expected true, got %s", got)
		}
	})

	t.Run("FalseF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = FalseF{}

		got := f.String()
		if got != "false" {
			t.Fatalf("expected false, got %s", got)
		}
	})

	t.Run("NotF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = NotF{F: Var("A")}

		got := f.String()
		if got != "!A" {
			t.Fatalf("expected !A, got %s", got)
		}
	})

	t.Run("AndF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = AndF{L: Var("A"), R: Var("B")}

		got := f.String()
		if got != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", got)
		}
	})

	t.Run("OrF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = OrF{L: Var("A"), R: Var("B")}

		got := f.String()
		if got != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", got)
		}
	})

	t.Run("ImplF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = ImplF{L: Var("A"), R: Var("B")}

		got := f.String()
		if got != "(A => B)" {
			t.Fatalf("expected (A => B), got %s", got)
		}
	})

	t.Run("IffF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = IffF{L: Var("A"), R: Var("B")}

		got := f.String()
		if got != "(A <=> B)" {
			t.Fatalf("expected (A <=> B), got %s", got)
		}
	})

	t.Run("GroupF implements Formula", func(t *testing.T) {
		t.Parallel()

		var f Formula = GroupF{F: Var("A")}

		got := f.String()
		if got != "(A)" {
			t.Fatalf("expected (A), got %s", got)
		}
	})
}
