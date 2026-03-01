package passwords

import (
	"testing"
)

func TestNewGeneratorOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions()

		if opts.passwordLength != DefaultPasswordLength {
			t.Fatalf("expected passwordLength %d, got %d", DefaultPasswordLength, opts.passwordLength)
		}
		if opts.minSpecialChar != DefaultMinSpecialChar {
			t.Fatalf("expected minSpecialChar %d, got %d", DefaultMinSpecialChar, opts.minSpecialChar)
		}
		if opts.minNum != DefaultMinNum {
			t.Fatalf("expected minNum %d, got %d", DefaultMinNum, opts.minNum)
		}
		if opts.minUpperCase != DefaultMinUpperCase {
			t.Fatalf("expected minUpperCase %d, got %d", DefaultMinUpperCase, opts.minUpperCase)
		}
	})
}

func TestWithPasswordLength(t *testing.T) {
	t.Parallel()

	t.Run("sets length when greater than or equal to default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithPasswordLength(30))

		if opts.passwordLength != 30 {
			t.Fatalf("expected 30, got %d", opts.passwordLength)
		}
	})

	t.Run("keeps default when less than default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithPasswordLength(10))

		if opts.passwordLength != DefaultPasswordLength {
			t.Fatalf("expected %d, got %d", DefaultPasswordLength, opts.passwordLength)
		}
	})

	t.Run("accepts exact default value", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithPasswordLength(DefaultPasswordLength))

		if opts.passwordLength != DefaultPasswordLength {
			t.Fatalf("expected %d, got %d", DefaultPasswordLength, opts.passwordLength)
		}
	})
}

func TestWithMinSpecialChar(t *testing.T) {
	t.Parallel()

	t.Run("sets min when greater than or equal to default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithMinSpecialChar(8))

		if opts.minSpecialChar != 8 {
			t.Fatalf("expected 8, got %d", opts.minSpecialChar)
		}
	})

	t.Run("keeps default when less than default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithMinSpecialChar(1))

		if opts.minSpecialChar != DefaultMinSpecialChar {
			t.Fatalf("expected %d, got %d", DefaultMinSpecialChar, opts.minSpecialChar)
		}
	})
}

func TestWithMinNum(t *testing.T) {
	t.Parallel()

	t.Run("sets min when greater than or equal to default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithMinNum(10))

		if opts.minNum != 10 {
			t.Fatalf("expected 10, got %d", opts.minNum)
		}
	})

	t.Run("keeps default when less than default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithMinNum(1))

		if opts.minNum != DefaultMinNum {
			t.Fatalf("expected %d, got %d", DefaultMinNum, opts.minNum)
		}
	})
}

func TestWithMinUpperCase(t *testing.T) {
	t.Parallel()

	t.Run("sets min when greater than or equal to default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithMinUpperCase(10))

		if opts.minUpperCase != 10 {
			t.Fatalf("expected 10, got %d", opts.minUpperCase)
		}
	})

	t.Run("keeps default when less than default", func(t *testing.T) {
		t.Parallel()

		opts := NewGeneratorOptions(WithMinUpperCase(1))

		if opts.minUpperCase != DefaultMinUpperCase {
			t.Fatalf("expected %d, got %d", DefaultMinUpperCase, opts.minUpperCase)
		}
	})
}
