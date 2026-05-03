package evaluate

import "testing"

func TestParadigm_String(t *testing.T) {
	t.Parallel()

	t.Run("deductive", func(t *testing.T) {
		t.Parallel()

		got := Deductive.String()
		if got != "deductive" {
			t.Fatalf("expected deductive, got %s", got)
		}
	})

	t.Run("bayesian", func(t *testing.T) {
		t.Parallel()

		got := Bayesian.String()
		if got != "bayesian" {
			t.Fatalf("expected bayesian, got %s", got)
		}
	})

	t.Run("fuzzy", func(t *testing.T) {
		t.Parallel()

		got := Fuzzy.String()
		if got != "fuzzy" {
			t.Fatalf("expected fuzzy, got %s", got)
		}
	})

	t.Run("table", func(t *testing.T) {
		t.Parallel()

		got := Table.String()
		if got != "table" {
			t.Fatalf("expected table, got %s", got)
		}
	})

	t.Run("scorecard", func(t *testing.T) {
		t.Parallel()

		got := Scorecard.String()
		if got != "scorecard" {
			t.Fatalf("expected scorecard, got %s", got)
		}
	})

	t.Run("tree", func(t *testing.T) {
		t.Parallel()

		got := Tree.String()
		if got != "tree" {
			t.Fatalf("expected tree, got %s", got)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		got := Paradigm(99).String()
		if got != "unknown" {
			t.Fatalf("expected unknown, got %s", got)
		}
	})
}
