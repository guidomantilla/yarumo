package parser

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
)

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("simple variable", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "A" {
			t.Fatalf("expected A, got %s", f.String())
		}
	})

	t.Run("true constant", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "true" {
			t.Fatalf("expected true, got %s", f.String())
		}
	})

	t.Run("false constant", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "false" {
			t.Fatalf("expected false, got %s", f.String())
		}
	})

	t.Run("negation", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("!A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "!A" {
			t.Fatalf("expected !A, got %s", f.String())
		}
	})

	t.Run("conjunction", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A & B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("disjunction", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A | B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", f.String())
		}
	})

	t.Run("implication", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A => B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A => B)" {
			t.Fatalf("expected (A => B), got %s", f.String())
		}
	})

	t.Run("biconditional", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A <=> B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A <=> B)" {
			t.Fatalf("expected (A <=> B), got %s", f.String())
		}
	})

	t.Run("parenthesized", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("(A)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A)" {
			t.Fatalf("expected (A), got %s", f.String())
		}
	})

	t.Run("precedence and over or", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A | B & C")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// AND binds tighter than OR: A | (B & C)
		if f.String() != "(A | (B & C))" {
			t.Fatalf("expected (A | (B & C)), got %s", f.String())
		}
	})

	t.Run("complex formula", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("(A & B) => !C | D")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "(((A & B)) => (!C | D))"
		if f.String() != expected {
			t.Fatalf("expected %s, got %s", expected, f.String())
		}
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("")
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("   ")
		if err == nil {
			t.Fatal("expected error for whitespace-only input")
		}
	})

	t.Run("unclosed parenthesis", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("(A & B")
		if err == nil {
			t.Fatal("expected error for unclosed paren")
		}
	})

	t.Run("unexpected token", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A & & B")
		if err == nil {
			t.Fatal("expected error for unexpected token")
		}
	})

	t.Run("trailing operator", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A &")
		if err == nil {
			t.Fatal("expected error for trailing operator")
		}
	})
}

func TestParse_unicodeSynonyms(t *testing.T) {
	t.Parallel()

	t.Run("unicode and", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A ∧ B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("unicode or", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A ∨ B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", f.String())
		}
	})

	t.Run("unicode not", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("¬A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "!A" {
			t.Fatalf("expected !A, got %s", f.String())
		}
	})

	t.Run("unicode implies", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A → B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A => B)" {
			t.Fatalf("expected (A => B), got %s", f.String())
		}
	})

	t.Run("unicode iff", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A ↔ B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A <=> B)" {
			t.Fatalf("expected (A <=> B), got %s", f.String())
		}
	})

	t.Run("unicode true", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("⊤")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "true" {
			t.Fatalf("expected true, got %s", f.String())
		}
	})

	t.Run("unicode false", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("⊥")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "false" {
			t.Fatalf("expected false, got %s", f.String())
		}
	})
}

func TestParse_textSynonyms(t *testing.T) {
	t.Parallel()

	t.Run("keyword AND", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A AND B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("keyword OR", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A OR B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", f.String())
		}
	})

	t.Run("keyword NOT", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("NOT A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "!A" {
			t.Fatalf("expected !A, got %s", f.String())
		}
	})

	t.Run("tilde as not", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("~A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "!A" {
			t.Fatalf("expected !A, got %s", f.String())
		}
	})

	t.Run("arrow implies", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A -> B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A => B)" {
			t.Fatalf("expected (A => B), got %s", f.String())
		}
	})

	t.Run("double arrow iff", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A <-> B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A <=> B)" {
			t.Fatalf("expected (A <=> B), got %s", f.String())
		}
	})

	t.Run("double ampersand", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A && B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("double pipe", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A || B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", f.String())
		}
	})

	t.Run("caret as and", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A ^ B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("keyword TRUE", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("TRUE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "true" {
			t.Fatalf("expected true, got %s", f.String())
		}
	})

	t.Run("keyword FALSE", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("FALSE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "false" {
			t.Fatalf("expected false, got %s", f.String())
		}
	})

	t.Run("keyword T", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("T")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "true" {
			t.Fatalf("expected true, got %s", f.String())
		}
	})

	t.Run("keyword lowercase not", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("not A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "!A" {
			t.Fatalf("expected !A, got %s", f.String())
		}
	})

	t.Run("keyword lowercase and", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A and B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("keyword lowercase or", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A or B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A | B)" {
			t.Fatalf("expected (A | B), got %s", f.String())
		}
	})
}

func TestParse_roundTrip(t *testing.T) {
	t.Parallel()

	t.Run("complex round trip", func(t *testing.T) {
		t.Parallel()

		original := "(A & B) => (!C | D)"

		f, err := Parse(original)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		reparsed, err := Parse(f.String())
		if err != nil {
			t.Fatalf("unexpected error on reparse: %v", err)
		}

		if !logic.Equivalent(f, reparsed) {
			t.Fatalf("round trip failed: %s != %s", f.String(), reparsed.String())
		}
	})

	t.Run("iff round trip", func(t *testing.T) {
		t.Parallel()

		f, err := Parse("A <=> B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		reparsed, err := Parse(f.String())
		if err != nil {
			t.Fatalf("unexpected error on reparse: %v", err)
		}

		if !logic.Equivalent(f, reparsed) {
			t.Fatalf("round trip failed: %s != %s", f.String(), reparsed.String())
		}
	})
}

func TestParseWith(t *testing.T) {
	t.Parallel()

	t.Run("with default options", func(t *testing.T) {
		t.Parallel()

		f, err := ParseWith("A & B")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("with strict option", func(t *testing.T) {
		t.Parallel()

		f, err := ParseWith("A & B", WithStrict(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})
}

func TestMustParse(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()

		f := MustParse("A & B")
		if f.String() != "(A & B)" {
			t.Fatalf("expected (A & B), got %s", f.String())
		}
	})

	t.Run("panics on invalid input", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
		}()

		MustParse("")
	})
}

func TestParse_errors(t *testing.T) {
	t.Parallel()

	t.Run("error in iff right side", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A <=> &")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in impl right side", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A => &")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in or right side", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A | &")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in and right side", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A & |")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in not operand", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("!")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("error in group content", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("(&)")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unexpected operator at start", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("& A")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("trailing tokens after valid formula", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A B")
		if err == nil {
			t.Fatal("expected error for trailing tokens")
		}
	})
}

func TestParseError(t *testing.T) {
	t.Parallel()

	t.Run("error message", func(t *testing.T) {
		t.Parallel()

		e := &ParseError{Pos: 5, Col: 6, Msg: "test error"}

		got := e.Error()
		if got != "test error" {
			t.Fatalf("expected test error, got %s", got)
		}
	})

	t.Run("position preserved", func(t *testing.T) {
		t.Parallel()

		_, err := Parse("A & & B")
		if err == nil {
			t.Fatal("expected error")
		}

		var pe *ParseError

		ok := errors.As(err, &pe)
		if !ok {
			t.Fatal("expected *ParseError")
		}

		if pe.Pos < 0 {
			t.Fatalf("expected positive position, got %d", pe.Pos)
		}
	})
}
