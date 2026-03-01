package random

import (
	"errors"
	"io"
	"math/big"
	"strings"
	"testing"
)

// withRandInt temporarily replaces the package-level randInt and restores it after fn returns.
func withRandInt(temp func(reader io.Reader, limit *big.Int) (*big.Int, error), fn func()) {
	orig := randInt
	randInt = temp

	defer func() { randInt = orig }()

	fn()
}

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

		got := Bytes(32)
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
	t.Run("negative limit returns zero", func(t *testing.T) {
		t.Parallel()

		n, err := Number(-5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n != 0 {
			t.Fatalf("got %d, want 0", n)
		}
	})

	t.Run("zero limit returns zero", func(t *testing.T) {
		t.Parallel()

		n, err := Number(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n != 0 {
			t.Fatalf("got %d, want 0", n)
		}
	})

	t.Run("max of one always returns zero", func(t *testing.T) {
		t.Parallel()

		n, err := Number(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n != 0 {
			t.Fatalf("got %d, want 0", n)
		}
	})

	t.Run("positive max returns value in range", func(t *testing.T) {
		t.Parallel()

		n, err := Number(100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n < 0 || n >= 100 {
			t.Fatalf("got %d, want value in [0, 100)", n)
		}
	})

	t.Run("error from rand", func(t *testing.T) {
		withRandInt(func(_ io.Reader, _ *big.Int) (*big.Int, error) {
			return nil, errors.New("boom")
		}, func() {
			_, err := Number(10)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	})
}

func TestString(t *testing.T) {
	t.Run("negative size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := String(-1, "abc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := String(0, "abc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("empty charset returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := String(5, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("single char charset is deterministic", func(t *testing.T) {
		t.Parallel()

		got, err := String(5, "x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "xxxxx" {
			t.Fatalf("got %q, want %q", got, "xxxxx")
		}
	})

	t.Run("correct length and charset", func(t *testing.T) {
		t.Parallel()

		got, err := String(10, "abcdef")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 10 {
			t.Fatalf("got length %d, want 10", len(got))
		}

		if !allRunesIn(got, "abcdef") {
			t.Fatalf("string %q contains chars outside charset", got)
		}
	})

	t.Run("error from rand", func(t *testing.T) {
		withRandInt(func(_ io.Reader, _ *big.Int) (*big.Int, error) {
			return nil, errors.New("boom")
		}, func() {
			_, err := String(1, "ab")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	})
}

func TestTextLower(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := TextLower(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got, err := TextLower(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got, err := TextLower(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allRunesIn(got, LowerCharSet) {
			t.Fatalf("string %q contains chars outside LowerCharSet", got)
		}
	})
}

func TestTextUpper(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := TextUpper(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got, err := TextUpper(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got, err := TextUpper(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allRunesIn(got, UpperCharSet) {
			t.Fatalf("string %q contains chars outside UpperCharSet", got)
		}
	})
}

func TestTextNumber(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := TextNumber(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got, err := TextNumber(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got, err := TextNumber(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allRunesIn(got, NumberSet) {
			t.Fatalf("string %q contains chars outside NumberSet", got)
		}
	})
}

func TestTextSpecial(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := TextSpecial(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got, err := TextSpecial(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got, err := TextSpecial(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allRunesIn(got, SpecialCharSet) {
			t.Fatalf("string %q contains chars outside SpecialCharSet", got)
		}
	})
}

func TestTextAlpha(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := TextAlpha(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got, err := TextAlpha(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got, err := TextAlpha(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allRunesIn(got, AlphaSet) {
			t.Fatalf("string %q contains chars outside AlphaSet", got)
		}
	})
}

func TestTextAlphaNum(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := TextAlphaNum(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got, err := TextAlphaNum(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got, err := TextAlphaNum(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allRunesIn(got, AlphaNumSet) {
			t.Fatalf("string %q contains chars outside AlphaNumSet", got)
		}
	})
}

func TestTextAll(t *testing.T) {
	t.Parallel()

	t.Run("zero size returns empty", func(t *testing.T) {
		t.Parallel()

		got, err := TextAll(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("correct length", func(t *testing.T) {
		t.Parallel()

		got, err := TextAll(20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("got length %d, want 20", len(got))
		}
	})

	t.Run("all chars from charset", func(t *testing.T) {
		t.Parallel()

		got, err := TextAll(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allRunesIn(got, AllCharSet) {
			t.Fatalf("string %q contains chars outside AllCharSet", got)
		}
	})
}
