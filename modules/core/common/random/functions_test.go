package random

import (
	"strings"
	"testing"
)

// allRunesIn returns true when every rune in s exists in charset.
func allRunesIn(s, charset string) bool {
	for _, r := range s {
		if !strings.ContainsRune(charset, r) {
			return false
		}
	}

	return true
}

func TestBytes(t *testing.T) {
	t.Parallel()

	t.Run("negative size returns nil", func(t *testing.T) {
		t.Parallel()

		got := Bytes(-1)
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("zero size returns nil", func(t *testing.T) {
		t.Parallel()

		got := Bytes(0)
		if got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("positive size returns correct length", func(t *testing.T) {
		t.Parallel()

		got := Bytes(16)
		if len(got) != 16 {
			t.Fatalf("got length %d, want 16", len(got))
		}
	})

	t.Run("output contains non-zero bytes", func(t *testing.T) {
		t.Parallel()

		got := Bytes(64)

		allZero := true

		for _, b := range got {
			if b != 0 {
				allZero = false
				break
			}
		}

		if allZero {
			t.Fatal("expected non-zero bytes in random output")
		}
	})
}

func TestNumber(t *testing.T) {
	t.Parallel()

	t.Run("negative limit returns zero", func(t *testing.T) {
		t.Parallel()

		n := Number(-5)
		if n != 0 {
			t.Fatalf("got %d, want 0", n)
		}
	})

	t.Run("zero limit returns zero", func(t *testing.T) {
		t.Parallel()

		n := Number(0)
		if n != 0 {
			t.Fatalf("got %d, want 0", n)
		}
	})

	t.Run("limit of one always returns zero", func(t *testing.T) {
		t.Parallel()

		n := Number(1)
		if n != 0 {
			t.Fatalf("got %d, want 0", n)
		}
	})

	t.Run("positive limit returns value in range", func(t *testing.T) {
		t.Parallel()

		n := Number(100)
		if n < 0 || n >= 100 {
			t.Fatalf("got %d, want value in [0, 100)", n)
		}
	})
}

func TestString(t *testing.T) {
	t.Parallel()

	t.Run("negative size returns empty", func(t *testing.T) {
		t.Parallel()

		got := String(-1, "abc")
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := String(0, "abc")
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty charset returns empty", func(t *testing.T) {
		t.Parallel()

		got := String(5, "")
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("single char charset is deterministic", func(t *testing.T) {
		t.Parallel()

		got := String(5, "x")
		if got != "xxxxx" {
			t.Fatalf("got %q, want %q", got, "xxxxx")
		}
	})

	t.Run("correct length and charset", func(t *testing.T) {
		t.Parallel()

		got := String(10, "abcdef")
		if len(got) != 10 {
			t.Fatalf("got length %d, want 10", len(got))
		}

		if !allRunesIn(got, "abcdef") {
			t.Fatalf("string %q contains chars outside charset", got)
		}
	})
}

func TestTextLower(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := TextLower(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got := TextLower(20)
		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got := TextLower(32)
		if !allRunesIn(got, LowerCharSet) {
			t.Fatalf("string %q contains chars outside LowerCharSet", got)
		}
	})
}

func TestTextUpper(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := TextUpper(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got := TextUpper(20)
		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got := TextUpper(32)
		if !allRunesIn(got, UpperCharSet) {
			t.Fatalf("string %q contains chars outside UpperCharSet", got)
		}
	})
}

func TestTextNumber(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := TextNumber(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got := TextNumber(20)
		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got := TextNumber(32)
		if !allRunesIn(got, NumberSet) {
			t.Fatalf("string %q contains chars outside NumberSet", got)
		}
	})
}

func TestTextSpecial(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := TextSpecial(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got := TextSpecial(20)
		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got := TextSpecial(32)
		if !allRunesIn(got, SpecialCharSet) {
			t.Fatalf("string %q contains chars outside SpecialCharSet", got)
		}
	})
}

func TestTextAlpha(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := TextAlpha(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got := TextAlpha(20)
		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got := TextAlpha(32)
		if !allRunesIn(got, AlphaSet) {
			t.Fatalf("string %q contains chars outside AlphaSet", got)
		}
	})
}

func TestTextAlphaNum(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := TextAlphaNum(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got := TextAlphaNum(20)
		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got := TextAlphaNum(32)
		if !allRunesIn(got, AlphaNumSet) {
			t.Fatalf("string %q contains chars outside AlphaNumSet", got)
		}
	})
}

func TestTextAll(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got := TextAll(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got := TextAll(20)
		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got := TextAll(32)
		if !allRunesIn(got, AllCharSet) {
			t.Fatalf("string %q contains chars outside AllCharSet", got)
		}
	})
}
