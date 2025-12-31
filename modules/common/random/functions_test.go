package random

import (
	"errors"
	"io"
	"math/big"
	"testing"
)

// helper to restore randInt after test
func withRandInt(temp func(reader io.Reader, max *big.Int) (*big.Int, error), fn func()) {
	orig := randInt
	randInt = temp

	defer func() { randInt = orig }()

	fn()
}

func TestBytes_Size(t *testing.T) {
	for _, n := range []int{0, 1, 16, 64} {
		b := Bytes(n)
		if len(b) != n {
			t.Fatalf("expected Bytes length %d, got %d", n, len(b))
		}
	}
}

func TestNumber_Normal(t *testing.T) {
	for _, max := range []int64{1, 2, 10, 100} {
		n, err := Number(max)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n < 0 || n >= max {
			t.Fatalf("number out of range: %d not in [0,%d)", n, max)
		}
	}
}

func TestNumber_Error(t *testing.T) {
	withRandInt(func(r io.Reader, max *big.Int) (*big.Int, error) {
		return nil, errors.New("boom")
	}, func() {
		_, err := Number(10)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestString_EmptyInputs(t *testing.T) {
	s, err := String(0, "abc")
	if err != nil || s != "" {
		t.Fatalf("expected empty string and nil error, got %q, %v", s, err)
	}

	s, err = String(5, "")
	if err != nil || s != "" {
		t.Fatalf("expected empty string and nil error, got %q, %v", s, err)
	}
}

func TestString_NormalAndDeterministic(t *testing.T) {
	// deterministic charset of size 1
	s, err := String(5, "x")
	if err != nil || s != "xxxxx" {
		t.Fatalf("expected xxxxx, got %q, err=%v", s, err)
	}
}

func TestString_Error(t *testing.T) {
	withRandInt(func(r io.Reader, max *big.Int) (*big.Int, error) {
		return nil, errors.New("boom")
	}, func() {
		_, err := String(1, "ab")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func containsAll(set string, s string) bool {
	for _, r := range s {
		if !stringsContainsRune(set, r) {
			return false
		}
	}

	return true
}

func stringsContainsRune(set string, r rune) bool {
	for _, x := range set {
		if x == r {
			return true
		}
	}

	return false
}

func TestTextHelpers(t *testing.T) {
	tests := []struct {
		name string
		fn   func(int) (string, error)
		set  string
	}{
		{"lower", TextLower, LowerCharSet},
		{"upper", TextUpper, UpperCharSet},
		{"number", TextNumber, NumberSet},
		{"special", TextSpecial, SpecialCharSet},
		{"alpha", TextAlpha, AlphaSet},
		{"alphanum", TextAlphaNum, AlphaNumSet},
		{"all", TextAll, AllCharSet},
	}
	for _, tc := range tests {
		got, err := tc.fn(32)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}

		if len(got) != 32 {
			t.Fatalf("%s: expected length 32, got %d", tc.name, len(got))
		}

		if !containsAll(tc.set, got) {
			t.Fatalf("%s: generated string contains characters outside the set", tc.name)
		}
	}
}
