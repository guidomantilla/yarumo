package passwords

import (
	"errors"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	t.Parallel()

	t.Run("creates generator with defaults", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()

		if g == nil {
			t.Fatal("expected non-nil generator")
		}
		if g.passwordLength != DefaultPasswordLength {
			t.Fatalf("expected passwordLength %d, got %d", DefaultPasswordLength, g.passwordLength)
		}
	})

	t.Run("creates generator with custom options", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator(WithPasswordLength(30))

		if g.passwordLength != 30 {
			t.Fatalf("expected passwordLength 30, got %d", g.passwordLength)
		}
	})
}

func TestGenerator_Generate(t *testing.T) {
	t.Parallel()

	t.Run("generates password of expected length", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		password := g.Generate()

		if len(password) != DefaultPasswordLength {
			t.Fatalf("expected length %d, got %d", DefaultPasswordLength, len(password))
		}
	})

	t.Run("generated password passes validation", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		password := g.Generate()

		err := g.Validate(password)
		if err != nil {
			t.Fatalf("expected generated password to pass validation: %v", err)
		}
	})
}

func TestGenerator_Validate(t *testing.T) {
	t.Parallel()

	t.Run("returns error for short password", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		err := g.Validate("short")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPasswordLength) {
			t.Fatalf("expected ErrPasswordLength, got %v", err)
		}
	})

	t.Run("returns error for password without enough special chars", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		// 26 chars, 6 upper, 6 nums, 0 special
		err := g.Validate("AABBCC112233aabbccddeeffgg")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPasswordSpecialChars) {
			t.Fatalf("expected ErrPasswordSpecialChars, got %v", err)
		}
	})

	t.Run("returns error for password without enough numbers", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		// 26 chars, 6 upper, 4 special, 0 nums
		err := g.Validate("AABBCC@@##$$aabbccddeeffgg")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPasswordNumbers) {
			t.Fatalf("expected ErrPasswordNumbers, got %v", err)
		}
	})

	t.Run("returns error for password without enough uppercase", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		// 26 chars, 4 special, 6 nums, 0 upper
		err := g.Validate("@@##$$112233aabbccddeeffgg")

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPasswordUppercaseChars) {
			t.Fatalf("expected ErrPasswordUppercaseChars, got %v", err)
		}
	})

	t.Run("accepts valid password", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		// 26 chars, 6 upper, 4 special, 6 nums, rest lower
		err := g.Validate("AABBCCDD@@##$$112233aabbcc")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("validation error wraps as domain error", func(t *testing.T) {
		t.Parallel()

		g := NewGenerator()
		err := g.Validate("short")

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !strings.Contains(err.Error(), ErrValidationFailed.Error()) {
			t.Fatalf("expected validation failed in error, got %q", err.Error())
		}
	})
}
