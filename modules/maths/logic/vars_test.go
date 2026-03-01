package logic

import (
	"slices"
	"testing"
)

func TestVar_Vars(t *testing.T) {
	t.Parallel()

	t.Run("returns single variable", func(t *testing.T) {
		t.Parallel()

		got := Var("A").Vars()
		if len(got) != 1 {
			t.Fatalf("expected 1 variable, got %d", len(got))
		}

		if got[0] != "A" {
			t.Fatalf("expected A, got %s", got[0])
		}
	})
}

func TestTrueF_Vars(t *testing.T) {
	t.Parallel()

	got := TrueF{}.Vars()
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestFalseF_Vars(t *testing.T) {
	t.Parallel()

	got := FalseF{}.Vars()
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestNotF_Vars(t *testing.T) {
	t.Parallel()

	t.Run("single variable", func(t *testing.T) {
		t.Parallel()

		got := NotF{F: Var("A")}.Vars()
		if len(got) != 1 || got[0] != "A" {
			t.Fatalf("expected [A], got %v", got)
		}
	})

	t.Run("nested", func(t *testing.T) {
		t.Parallel()

		got := NotF{F: AndF{L: Var("A"), R: Var("B")}}.Vars()

		expected := []Var{"A", "B"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}

func TestAndF_Vars(t *testing.T) {
	t.Parallel()

	t.Run("distinct variables", func(t *testing.T) {
		t.Parallel()

		got := AndF{L: Var("A"), R: Var("B")}.Vars()

		expected := []Var{"A", "B"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})

	t.Run("duplicate variables", func(t *testing.T) {
		t.Parallel()

		got := AndF{L: Var("A"), R: Var("A")}.Vars()

		expected := []Var{"A"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})

	t.Run("sorted output", func(t *testing.T) {
		t.Parallel()

		got := AndF{L: Var("Z"), R: Var("A")}.Vars()

		expected := []Var{"A", "Z"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}

func TestOrF_Vars(t *testing.T) {
	t.Parallel()

	t.Run("distinct variables", func(t *testing.T) {
		t.Parallel()

		got := OrF{L: Var("A"), R: Var("B")}.Vars()

		expected := []Var{"A", "B"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})

	t.Run("duplicate variables", func(t *testing.T) {
		t.Parallel()

		got := OrF{L: Var("A"), R: Var("A")}.Vars()

		expected := []Var{"A"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}

func TestImplF_Vars(t *testing.T) {
	t.Parallel()

	t.Run("distinct variables", func(t *testing.T) {
		t.Parallel()

		got := ImplF{L: Var("A"), R: Var("B")}.Vars()

		expected := []Var{"A", "B"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})

	t.Run("overlapping variables", func(t *testing.T) {
		t.Parallel()

		got := ImplF{L: AndF{L: Var("A"), R: Var("B")}, R: Var("A")}.Vars()

		expected := []Var{"A", "B"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}

func TestIffF_Vars(t *testing.T) {
	t.Parallel()

	t.Run("distinct variables", func(t *testing.T) {
		t.Parallel()

		got := IffF{L: Var("A"), R: Var("B")}.Vars()

		expected := []Var{"A", "B"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})

	t.Run("nested with duplicates", func(t *testing.T) {
		t.Parallel()

		got := IffF{L: OrF{L: Var("A"), R: Var("C")}, R: AndF{L: Var("B"), R: Var("A")}}.Vars()

		expected := []Var{"A", "B", "C"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}

func TestGroupF_Vars(t *testing.T) {
	t.Parallel()

	t.Run("single variable", func(t *testing.T) {
		t.Parallel()

		got := GroupF{F: Var("A")}.Vars()
		if len(got) != 1 || got[0] != "A" {
			t.Fatalf("expected [A], got %v", got)
		}
	})

	t.Run("complex inner", func(t *testing.T) {
		t.Parallel()

		got := GroupF{F: OrF{L: Var("X"), R: Var("Y")}}.Vars()

		expected := []Var{"X", "Y"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}

func TestVars_complex(t *testing.T) {
	t.Parallel()

	t.Run("deeply nested formula", func(t *testing.T) {
		t.Parallel()
		// (A & B) => (C | !A)
		f := ImplF{
			L: AndF{L: Var("A"), R: Var("B")},
			R: OrF{L: Var("C"), R: NotF{F: Var("A")}},
		}
		got := f.Vars()

		expected := []Var{"A", "B", "C"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})

	t.Run("constants have no variables", func(t *testing.T) {
		t.Parallel()

		f := AndF{L: TrueF{}, R: FalseF{}}

		got := f.Vars()
		if len(got) != 0 {
			t.Fatalf("expected empty, got %v", got)
		}
	})

	t.Run("mixed constants and variables", func(t *testing.T) {
		t.Parallel()

		f := OrF{L: TrueF{}, R: Var("A")}
		got := f.Vars()

		expected := []Var{"A"}
		if !slices.Equal(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	})
}
