package generator

import (
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns documented defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.passwordLength != DefaultPasswordLength {
			t.Fatalf("passwordLength: got %d, want %d", opts.passwordLength, DefaultPasswordLength)
		}
		if opts.minSpecialChar != DefaultMinSpecialChar {
			t.Fatalf("minSpecialChar: got %d, want %d", opts.minSpecialChar, DefaultMinSpecialChar)
		}
		if opts.minNumber != DefaultMinNumber {
			t.Fatalf("minNumber: got %d, want %d", opts.minNumber, DefaultMinNumber)
		}
		if opts.minUpperCase != DefaultMinUpperCase {
			t.Fatalf("minUpperCase: got %d, want %d", opts.minUpperCase, DefaultMinUpperCase)
		}
		if opts.minLowerCase != DefaultMinLowerCase {
			t.Fatalf("minLowerCase: got %d, want %d", opts.minLowerCase, DefaultMinLowerCase)
		}
	})

	t.Run("applies overrides in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithPasswordLength(32),
			WithMinSpecialChar(2),
			WithMinNumber(2),
			WithMinUpperCase(2),
			WithMinLowerCase(2),
		)

		if opts.passwordLength != 32 {
			t.Fatalf("passwordLength: got %d, want 32", opts.passwordLength)
		}
		if opts.minSpecialChar != 2 {
			t.Fatalf("minSpecialChar: got %d, want 2", opts.minSpecialChar)
		}
		if opts.minNumber != 2 {
			t.Fatalf("minNumber: got %d, want 2", opts.minNumber)
		}
		if opts.minUpperCase != 2 {
			t.Fatalf("minUpperCase: got %d, want 2", opts.minUpperCase)
		}
		if opts.minLowerCase != 2 {
			t.Fatalf("minLowerCase: got %d, want 2", opts.minLowerCase)
		}
	})
}

func TestWithPasswordLength(t *testing.T) {
	t.Parallel()

	t.Run("accepts smaller-than-default value (no silent rejection)", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPasswordLength(10))

		if opts.passwordLength != 10 {
			t.Fatalf("got %d, want 10 — legacy silent rejection regression", opts.passwordLength)
		}
	})

	t.Run("accepts larger-than-default value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPasswordLength(64))

		if opts.passwordLength != 64 {
			t.Fatalf("got %d, want 64", opts.passwordLength)
		}
	})

	t.Run("accepts zero value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPasswordLength(0))

		if opts.passwordLength != 0 {
			t.Fatalf("got %d, want 0", opts.passwordLength)
		}
	})

	t.Run("clamps negative value to zero", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPasswordLength(-5))

		if opts.passwordLength != 0 {
			t.Fatalf("got %d, want 0 (negative clamped)", opts.passwordLength)
		}
	})
}

func TestWithMinSpecialChar(t *testing.T) {
	t.Parallel()

	t.Run("accepts custom value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinSpecialChar(8))

		if opts.minSpecialChar != 8 {
			t.Fatalf("got %d, want 8", opts.minSpecialChar)
		}
	})

	t.Run("accepts smaller-than-default value (no silent rejection)", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinSpecialChar(1))

		if opts.minSpecialChar != 1 {
			t.Fatalf("got %d, want 1 — legacy silent rejection regression", opts.minSpecialChar)
		}
	})

	t.Run("clamps negative value to zero", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinSpecialChar(-3))

		if opts.minSpecialChar != 0 {
			t.Fatalf("got %d, want 0 (negative clamped)", opts.minSpecialChar)
		}
	})
}

func TestWithMinNumber(t *testing.T) {
	t.Parallel()

	t.Run("accepts custom value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinNumber(10))

		if opts.minNumber != 10 {
			t.Fatalf("got %d, want 10", opts.minNumber)
		}
	})

	t.Run("accepts smaller-than-default value (no silent rejection)", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinNumber(1))

		if opts.minNumber != 1 {
			t.Fatalf("got %d, want 1", opts.minNumber)
		}
	})

	t.Run("clamps negative value to zero", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinNumber(-1))

		if opts.minNumber != 0 {
			t.Fatalf("got %d, want 0", opts.minNumber)
		}
	})
}

func TestWithMinUpperCase(t *testing.T) {
	t.Parallel()

	t.Run("accepts custom value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinUpperCase(10))

		if opts.minUpperCase != 10 {
			t.Fatalf("got %d, want 10", opts.minUpperCase)
		}
	})

	t.Run("accepts smaller-than-default value (no silent rejection)", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinUpperCase(1))

		if opts.minUpperCase != 1 {
			t.Fatalf("got %d, want 1", opts.minUpperCase)
		}
	})

	t.Run("clamps negative value to zero", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinUpperCase(-2))

		if opts.minUpperCase != 0 {
			t.Fatalf("got %d, want 0", opts.minUpperCase)
		}
	})
}

func TestWithMinLowerCase(t *testing.T) {
	t.Parallel()

	t.Run("accepts custom value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinLowerCase(10))

		if opts.minLowerCase != 10 {
			t.Fatalf("got %d, want 10", opts.minLowerCase)
		}
	})

	t.Run("accepts zero (no lower-case required)", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinLowerCase(0))

		if opts.minLowerCase != 0 {
			t.Fatalf("got %d, want 0", opts.minLowerCase)
		}
	})

	t.Run("clamps negative value to zero", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinLowerCase(-4))

		if opts.minLowerCase != 0 {
			t.Fatalf("got %d, want 0", opts.minLowerCase)
		}
	})
}
