package logic

import "testing"

func TestVar_String(t *testing.T) {
	t.Parallel()

	t.Run("simple variable", func(t *testing.T) {
		t.Parallel()

		got := Var("A").String()
		if got != "A" {
			t.Fatalf("expected A, got %s", got)
		}
	})

	t.Run("multi-char variable", func(t *testing.T) {
		t.Parallel()

		got := Var("rain").String()
		if got != "rain" {
			t.Fatalf("expected rain, got %s", got)
		}
	})
}

func TestTrueF_String(t *testing.T) {
	t.Parallel()

	got := TrueF{}.String()
	if got != "true" {
		t.Fatalf("expected true, got %s", got)
	}
}

func TestFalseF_String(t *testing.T) {
	t.Parallel()

	got := FalseF{}.String()
	if got != "false" {
		t.Fatalf("expected false, got %s", got)
	}
}

func TestNotF_String(t *testing.T) {
	t.Parallel()

	t.Run("not variable", func(t *testing.T) {
		t.Parallel()

		got := NotF{F: Var("A")}.String()
		if got != "!A" {
			t.Fatalf("expected !A, got %s", got)
		}
	})

	t.Run("not nested", func(t *testing.T) {
		t.Parallel()

		got := NotF{F: NotF{F: Var("A")}}.String()
		if got != "!!A" {
			t.Fatalf("expected !!A, got %s", got)
		}
	})
}

func TestAndF_String(t *testing.T) {
	t.Parallel()

	got := AndF{L: Var("A"), R: Var("B")}.String()
	if got != "(A & B)" {
		t.Fatalf("expected (A & B), got %s", got)
	}
}

func TestOrF_String(t *testing.T) {
	t.Parallel()

	got := OrF{L: Var("A"), R: Var("B")}.String()
	if got != "(A | B)" {
		t.Fatalf("expected (A | B), got %s", got)
	}
}

func TestImplF_String(t *testing.T) {
	t.Parallel()

	got := ImplF{L: Var("A"), R: Var("B")}.String()
	if got != "(A => B)" {
		t.Fatalf("expected (A => B), got %s", got)
	}
}

func TestIffF_String(t *testing.T) {
	t.Parallel()

	got := IffF{L: Var("A"), R: Var("B")}.String()
	if got != "(A <=> B)" {
		t.Fatalf("expected (A <=> B), got %s", got)
	}
}

func TestGroupF_String(t *testing.T) {
	t.Parallel()

	t.Run("group variable", func(t *testing.T) {
		t.Parallel()

		got := GroupF{F: Var("A")}.String()
		if got != "(A)" {
			t.Fatalf("expected (A), got %s", got)
		}
	})

	t.Run("group complex", func(t *testing.T) {
		t.Parallel()

		got := GroupF{F: AndF{L: Var("A"), R: Var("B")}}.String()
		if got != "((A & B))" {
			t.Fatalf("expected ((A & B)), got %s", got)
		}
	})
}

func TestString_nested(t *testing.T) {
	t.Parallel()

	t.Run("impl with and", func(t *testing.T) {
		t.Parallel()

		f := ImplF{L: AndF{L: Var("A"), R: Var("B")}, R: Var("C")}

		got := f.String()
		if got != "((A & B) => C)" {
			t.Fatalf("expected ((A & B) => C), got %s", got)
		}
	})

	t.Run("not and or", func(t *testing.T) {
		t.Parallel()

		f := NotF{F: OrF{L: Var("A"), R: AndF{L: Var("B"), R: Var("C")}}}

		got := f.String()
		if got != "!(A | (B & C))" {
			t.Fatalf("expected !(A | (B & C)), got %s", got)
		}
	})
}
