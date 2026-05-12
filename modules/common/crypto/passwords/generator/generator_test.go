package generator

import (
	"errors"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	t.Parallel()

	t.Run("creates generator with defaults", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g == nil {
			t.Fatal("expected non-nil generator")
		}
		if g.PasswordLength() != DefaultPasswordLength {
			t.Fatalf("PasswordLength: got %d, want %d", g.PasswordLength(), DefaultPasswordLength)
		}
		if g.MinSpecialChar() != DefaultMinSpecialChar {
			t.Fatalf("MinSpecialChar: got %d, want %d", g.MinSpecialChar(), DefaultMinSpecialChar)
		}
		if g.MinNumber() != DefaultMinNumber {
			t.Fatalf("MinNumber: got %d, want %d", g.MinNumber(), DefaultMinNumber)
		}
		if g.MinUpperCase() != DefaultMinUpperCase {
			t.Fatalf("MinUpperCase: got %d, want %d", g.MinUpperCase(), DefaultMinUpperCase)
		}
		if g.MinLowerCase() != DefaultMinLowerCase {
			t.Fatalf("MinLowerCase: got %d, want %d", g.MinLowerCase(), DefaultMinLowerCase)
		}
	})

	t.Run("creates generator with custom options", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator(WithPasswordLength(40), WithMinLowerCase(8))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.PasswordLength() != 40 {
			t.Fatalf("PasswordLength: got %d, want 40", g.PasswordLength())
		}
		if g.MinLowerCase() != 8 {
			t.Fatalf("MinLowerCase: got %d, want 8", g.MinLowerCase())
		}
	})

	t.Run("rejects when minimums exceed total length", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator(
			WithPasswordLength(8),
			WithMinSpecialChar(4),
			WithMinNumber(4),
			WithMinUpperCase(4),
			WithMinLowerCase(4),
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if g != nil {
			t.Fatal("expected nil generator on error")
		}
		if !errors.Is(err, ErrConstraintsExceedLength) {
			t.Fatalf("expected ErrConstraintsExceedLength, got %v", err)
		}
		if !errors.Is(err, ErrInvalidOption) {
			t.Fatalf("expected ErrInvalidOption in chain, got %v", err)
		}
	})

	t.Run("rejects when zero length but positive minimums", func(t *testing.T) {
		t.Parallel()

		_, err := NewGenerator(WithPasswordLength(0), WithMinNumber(1), WithMinSpecialChar(0), WithMinUpperCase(0), WithMinLowerCase(0))
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrConstraintsExceedLength) {
			t.Fatalf("expected ErrConstraintsExceedLength, got %v", err)
		}
	})

	t.Run("accepts zero everything (degenerate empty password)", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator(
			WithPasswordLength(0),
			WithMinSpecialChar(0),
			WithMinNumber(0),
			WithMinUpperCase(0),
			WithMinLowerCase(0),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.PasswordLength() != 0 {
			t.Fatalf("got %d, want 0", g.PasswordLength())
		}
	})
}

func TestGenerator_Generate(t *testing.T) {
	t.Parallel()

	t.Run("generates password of expected length", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		pw, err := g.Generate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pw) != DefaultPasswordLength {
			t.Fatalf("got length %d, want %d", len(pw), DefaultPasswordLength)
		}
	})

	t.Run("generated password passes validation", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		pw, err := g.Generate()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}

		errValidate := g.Validate(pw)
		if errValidate != nil {
			t.Fatalf("expected generated password to pass validation: %v", errValidate)
		}
	})

	t.Run("two generations differ (basic non-constancy)", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		a, errA := g.Generate()
		if errA != nil {
			t.Fatalf("generate a: %v", errA)
		}
		b, errB := g.Generate()
		if errB != nil {
			t.Fatalf("generate b: %v", errB)
		}

		if a == b {
			t.Fatalf("two consecutive generations are identical: %q — entropy regression?", a)
		}
	})

	t.Run("respects boundary case minimums equal length", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator(
			WithPasswordLength(8),
			WithMinSpecialChar(2),
			WithMinNumber(2),
			WithMinUpperCase(2),
			WithMinLowerCase(2),
		)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		pw, err := g.Generate()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if len(pw) != 8 {
			t.Fatalf("got length %d, want 8", len(pw))
		}
		errValidate := g.Validate(pw)
		if errValidate != nil {
			t.Fatalf("validation failed: %v", errValidate)
		}
	})

	t.Run("degenerate zero-length password", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator(
			WithPasswordLength(0),
			WithMinSpecialChar(0),
			WithMinNumber(0),
			WithMinUpperCase(0),
			WithMinLowerCase(0),
		)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		pw, err := g.Generate()
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if pw != "" {
			t.Fatalf("got %q, want empty", pw)
		}
	})
}

func TestGenerator_Validate(t *testing.T) {
	t.Parallel()

	t.Run("returns error for short password", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		errValidate := g.Validate("short")
		if errValidate == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errValidate, ErrPasswordLength) {
			t.Fatalf("expected ErrPasswordLength, got %v", errValidate)
		}
	})

	t.Run("returns error for password without enough special chars", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		// 26 chars, 6 upper, 6 nums, 0 special, 14 lower
		errValidate := g.Validate("AABBCC112233aabbccddeeffgg")
		if errValidate == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errValidate, ErrPasswordSpecialChars) {
			t.Fatalf("expected ErrPasswordSpecialChars, got %v", errValidate)
		}
	})

	t.Run("returns error for password without enough numbers", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		// 26 chars, 6 upper, 4 special, 0 numbers, 16 lower
		errValidate := g.Validate("AABBCC@@##$$aabbccddeeffgg")
		if errValidate == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errValidate, ErrPasswordNumbers) {
			t.Fatalf("expected ErrPasswordNumbers, got %v", errValidate)
		}
	})

	t.Run("returns error for password without enough uppercase", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		// 26 chars, 0 upper, 4 special, 6 nums, 16 lower
		errValidate := g.Validate("@@##$$112233aabbccddeeffgg")
		if errValidate == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errValidate, ErrPasswordUppercaseChars) {
			t.Fatalf("expected ErrPasswordUppercaseChars, got %v", errValidate)
		}
	})

	t.Run("returns error for password without enough lowercase", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		// 26 chars, 6 upper, 4 special, 6 numbers, 0 lower, 10 chars filler in upper
		errValidate := g.Validate("AABBCCDDEEFF@@##$$112233ZZ")
		if errValidate == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(errValidate, ErrPasswordLowercaseChars) {
			t.Fatalf("expected ErrPasswordLowercaseChars, got %v", errValidate)
		}
	})

	t.Run("accepts valid password", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		// 26 chars, 6 upper, 4 special, 6 nums, 10 lower
		errValidate := g.Validate("AABBCCDD@@##$$112233aabbcc")
		if errValidate != nil {
			t.Fatalf("unexpected error: %v", errValidate)
		}
	})

	t.Run("validation error wraps as domain error", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		errValidate := g.Validate("short")

		var domErr *Error
		if !errors.As(errValidate, &domErr) {
			t.Fatalf("expected *Error, got %T", errValidate)
		}
		if !strings.Contains(errValidate.Error(), ErrValidationFailed.Error()) {
			t.Fatalf("expected validation failed in error, got %q", errValidate.Error())
		}
	})

	t.Run("custom-config password passes its own validator", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator(
			WithPasswordLength(12),
			WithMinSpecialChar(2),
			WithMinNumber(2),
			WithMinUpperCase(2),
			WithMinLowerCase(2),
		)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}

		pw, errGen := g.Generate()
		if errGen != nil {
			t.Fatalf("generate: %v", errGen)
		}
		errVal := g.Validate(pw)
		if errVal != nil {
			t.Fatalf("validate: %v", errVal)
		}
	})
}

func TestGenerator_Accessors(t *testing.T) {
	t.Parallel()

	t.Run("expose configured values", func(t *testing.T) {
		t.Parallel()

		g, err := NewGenerator(
			WithPasswordLength(20),
			WithMinSpecialChar(3),
			WithMinNumber(4),
			WithMinUpperCase(5),
			WithMinLowerCase(2),
		)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		if g.PasswordLength() != 20 {
			t.Fatalf("PasswordLength: %d", g.PasswordLength())
		}
		if g.MinSpecialChar() != 3 {
			t.Fatalf("MinSpecialChar: %d", g.MinSpecialChar())
		}
		if g.MinNumber() != 4 {
			t.Fatalf("MinNumber: %d", g.MinNumber())
		}
		if g.MinUpperCase() != 5 {
			t.Fatalf("MinUpperCase: %d", g.MinUpperCase())
		}
		if g.MinLowerCase() != 2 {
			t.Fatalf("MinLowerCase: %d", g.MinLowerCase())
		}
	})
}

func TestCountClasses(t *testing.T) {
	t.Parallel()

	t.Run("counts each class correctly", func(t *testing.T) {
		t.Parallel()

		counts := countClasses("Ab1@cD2#")

		if counts.special != 2 {
			t.Fatalf("special: got %d, want 2", counts.special)
		}
		if counts.number != 2 {
			t.Fatalf("number: got %d, want 2", counts.number)
		}
		if counts.upper != 2 {
			t.Fatalf("upper: got %d, want 2", counts.upper)
		}
		if counts.lower != 2 {
			t.Fatalf("lower: got %d, want 2", counts.lower)
		}
	})

	t.Run("ignores unknown characters", func(t *testing.T) {
		t.Parallel()

		counts := countClasses("éèê") // accented letters not in any charset

		if counts.special+counts.number+counts.upper+counts.lower != 0 {
			t.Fatalf("expected zero counts for non-charset runes, got %+v", counts)
		}
	})

	t.Run("empty input yields zero counts", func(t *testing.T) {
		t.Parallel()

		counts := countClasses("")

		if counts != (classCounts{}) {
			t.Fatalf("expected zero counts, got %+v", counts)
		}
	})
}
