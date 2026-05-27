package validation

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestIsRequired(t *testing.T) {
	t.Parallel()

	t.Run("happy path string", func(t *testing.T) {
		t.Parallel()

		err := IsRequired("hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := IsRequired(42)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty string", func(t *testing.T) {
		t.Parallel()

		err := IsRequired("")
		if !errors.Is(err, ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})

	t.Run("error zero int", func(t *testing.T) {
		t.Parallel()

		err := IsRequired(0)
		if !errors.Is(err, ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})
}

func TestMustBeUndefined(t *testing.T) {
	t.Parallel()

	t.Run("happy path empty string", func(t *testing.T) {
		t.Parallel()

		err := MustBeUndefined("")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error non-empty string", func(t *testing.T) {
		t.Parallel()

		err := MustBeUndefined("hello")
		if !errors.Is(err, ErrFieldMustBeUndefined) {
			t.Fatalf("expected ErrFieldMustBeUndefined, got %v", err)
		}
	})

	t.Run("happy path zero int", func(t *testing.T) {
		t.Parallel()

		err := MustBeUndefined(0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMinLen(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MinLen("hello", 3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := MinLen("hi", 5)
		if !errors.Is(err, ErrMinLen) {
			t.Fatalf("expected ErrMinLen, got %v", err)
		}
	})

	t.Run("negative threshold accepts empty", func(t *testing.T) {
		t.Parallel()

		err := MinLen("", -3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMaxLen(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MaxLen("hello", 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := MaxLen("hello world", 5)
		if !errors.Is(err, ErrMaxLen) {
			t.Fatalf("expected ErrMaxLen, got %v", err)
		}
	})

	t.Run("negative threshold rejects non-empty", func(t *testing.T) {
		t.Parallel()

		err := MaxLen("x", -1)
		if !errors.Is(err, ErrMaxLen) {
			t.Fatalf("expected ErrMaxLen, got %v", err)
		}
	})
}

func TestMatchesRegex(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MatchesRegex("abc123", `^[a-z]+\d+$`)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error mismatch", func(t *testing.T) {
		t.Parallel()

		err := MatchesRegex("ABC", `^[a-z]+$`)
		if !errors.Is(err, ErrRegexMismatch) {
			t.Fatalf("expected ErrRegexMismatch, got %v", err)
		}
	})

	t.Run("error invalid pattern", func(t *testing.T) {
		t.Parallel()

		err := MatchesRegex("x", `[`)
		if !errors.Is(err, ErrRegexInvalid) {
			t.Fatalf("expected ErrRegexInvalid, got %v", err)
		}
	})
}

func TestIsEmail(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("a@b.com")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("")
		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("not-an-email")
		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})

	t.Run("error with display name", func(t *testing.T) {
		t.Parallel()

		err := IsEmail("Name <a@b.com>")
		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})
}

func TestIsURL(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsURL("https://example.com/path")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsURL("")
		if !errors.Is(err, ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})

	t.Run("error missing scheme", func(t *testing.T) {
		t.Parallel()

		err := IsURL("example.com")
		if !errors.Is(err, ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})

	t.Run("error unparseable", func(t *testing.T) {
		t.Parallel()

		err := IsURL("http://[::1")
		if !errors.Is(err, ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})
}

func TestMin(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := Min(10, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := Min(3, 10)
		if !errors.Is(err, ErrMinValue) {
			t.Fatalf("expected ErrMinValue, got %v", err)
		}
	})

	t.Run("happy path float", func(t *testing.T) {
		t.Parallel()

		err := Min(1.5, 1.0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMax(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := Max(3, 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := Max(20, 10)
		if !errors.Is(err, ErrMaxValue) {
			t.Fatalf("expected ErrMaxValue, got %v", err)
		}
	})
}

func TestInRange(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := InRange(5, 0, 10)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := InRange(-1, 0, 10)
		if !errors.Is(err, ErrOutOfRange) {
			t.Fatalf("expected ErrOutOfRange, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := InRange(20, 0, 10)
		if !errors.Is(err, ErrOutOfRange) {
			t.Fatalf("expected ErrOutOfRange, got %v", err)
		}
	})

	t.Run("error invalid range", func(t *testing.T) {
		t.Parallel()

		err := InRange(5, 10, 0)
		if !errors.Is(err, ErrInvalidRange) {
			t.Fatalf("expected ErrInvalidRange, got %v", err)
		}
	})
}

func TestIsUID(t *testing.T) {
	t.Parallel()

	alwaysTrue := func(string) bool { return true }
	alwaysFalse := func(string) bool { return false }

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsUID("any-string", alwaysTrue)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty string", func(t *testing.T) {
		t.Parallel()

		err := IsUID("", alwaysTrue)
		if !errors.Is(err, ErrUIDInvalid) {
			t.Fatalf("expected ErrUIDInvalid, got %v", err)
		}
	})

	t.Run("error nil predicate", func(t *testing.T) {
		t.Parallel()

		err := IsUID("any-string", nil)
		if !errors.Is(err, ErrUIDInvalid) {
			t.Fatalf("expected ErrUIDInvalid, got %v", err)
		}
	})

	t.Run("error predicate rejects", func(t *testing.T) {
		t.Parallel()

		err := IsUID("any-string", alwaysFalse)
		if !errors.Is(err, ErrUIDInvalid) {
			t.Fatalf("expected ErrUIDInvalid, got %v", err)
		}
	})
}

func TestPositive(t *testing.T) {
	t.Parallel()

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := Positive(5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path float", func(t *testing.T) {
		t.Parallel()

		err := Positive(0.001)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error zero", func(t *testing.T) {
		t.Parallel()

		err := Positive(0)
		if !errors.Is(err, ErrNotPositive) {
			t.Fatalf("expected ErrNotPositive, got %v", err)
		}
	})

	t.Run("error negative", func(t *testing.T) {
		t.Parallel()

		err := Positive(-1)
		if !errors.Is(err, ErrNotPositive) {
			t.Fatalf("expected ErrNotPositive, got %v", err)
		}
	})
}

func TestNegative(t *testing.T) {
	t.Parallel()

	t.Run("happy path int64", func(t *testing.T) {
		t.Parallel()

		err := Negative(int64(-5))
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path float", func(t *testing.T) {
		t.Parallel()

		err := Negative(-0.001)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error zero", func(t *testing.T) {
		t.Parallel()

		err := Negative(0)
		if !errors.Is(err, ErrNotNegative) {
			t.Fatalf("expected ErrNotNegative, got %v", err)
		}
	})

	t.Run("error positive", func(t *testing.T) {
		t.Parallel()

		err := Negative(5)
		if !errors.Is(err, ErrNotNegative) {
			t.Fatalf("expected ErrNotNegative, got %v", err)
		}
	})
}

func TestNonZero(t *testing.T) {
	t.Parallel()

	t.Run("happy path positive", func(t *testing.T) {
		t.Parallel()

		err := NonZero(1)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path negative", func(t *testing.T) {
		t.Parallel()

		err := NonZero(-1)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path float", func(t *testing.T) {
		t.Parallel()

		err := NonZero(0.5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error zero", func(t *testing.T) {
		t.Parallel()

		err := NonZero(0)
		if !errors.Is(err, ErrZero) {
			t.Fatalf("expected ErrZero, got %v", err)
		}
	})
}

func TestMultipleOf(t *testing.T) {
	t.Parallel()

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := MultipleOf(15, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path zero is multiple of any", func(t *testing.T) {
		t.Parallel()

		err := MultipleOf(0, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path negative", func(t *testing.T) {
		t.Parallel()

		err := MultipleOf(-10, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error not multiple", func(t *testing.T) {
		t.Parallel()

		err := MultipleOf(7, 3)
		if !errors.Is(err, ErrNotMultipleOf) {
			t.Fatalf("expected ErrNotMultipleOf, got %v", err)
		}
	})

	t.Run("error zero factor", func(t *testing.T) {
		t.Parallel()

		err := MultipleOf(5, 0)
		if !errors.Is(err, ErrNotMultipleOf) {
			t.Fatalf("expected ErrNotMultipleOf, got %v", err)
		}
	})
}

func TestIsIntegerString(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsIntegerString("12345")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path negative", func(t *testing.T) {
		t.Parallel()

		err := IsIntegerString("-42")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsIntegerString("")
		if !errors.Is(err, ErrIntegerStringInvalid) {
			t.Fatalf("expected ErrIntegerStringInvalid, got %v", err)
		}
	})

	t.Run("error not integer", func(t *testing.T) {
		t.Parallel()

		err := IsIntegerString("3.14")
		if !errors.Is(err, ErrIntegerStringInvalid) {
			t.Fatalf("expected ErrIntegerStringInvalid, got %v", err)
		}
	})

	t.Run("error garbage", func(t *testing.T) {
		t.Parallel()

		err := IsIntegerString("abc")
		if !errors.Is(err, ErrIntegerStringInvalid) {
			t.Fatalf("expected ErrIntegerStringInvalid, got %v", err)
		}
	})
}

func TestIsFloatString(t *testing.T) {
	t.Parallel()

	t.Run("happy path int form", func(t *testing.T) {
		t.Parallel()

		err := IsFloatString("42")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path decimal", func(t *testing.T) {
		t.Parallel()

		err := IsFloatString("3.14")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path scientific", func(t *testing.T) {
		t.Parallel()

		err := IsFloatString("1.5e10")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsFloatString("")
		if !errors.Is(err, ErrFloatStringInvalid) {
			t.Fatalf("expected ErrFloatStringInvalid, got %v", err)
		}
	})

	t.Run("error garbage", func(t *testing.T) {
		t.Parallel()

		err := IsFloatString("not-a-number")
		if !errors.Is(err, ErrFloatStringInvalid) {
			t.Fatalf("expected ErrFloatStringInvalid, got %v", err)
		}
	})
}

type role string

func TestEqual(t *testing.T) {
	t.Parallel()

	t.Run("happy path string", func(t *testing.T) {
		t.Parallel()

		err := Equal("admin", "admin")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := Equal(42, 42)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path custom type", func(t *testing.T) {
		t.Parallel()

		err := Equal(role("admin"), role("admin"))
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error not equal", func(t *testing.T) {
		t.Parallel()

		err := Equal("admin", "user")
		if !errors.Is(err, ErrNotEqual) {
			t.Fatalf("expected ErrNotEqual, got %v", err)
		}
	})
}

func TestNotEqual(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := NotEqual("admin", "guest")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := NotEqual(1, 2)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error equal", func(t *testing.T) {
		t.Parallel()

		err := NotEqual("admin", "admin")
		if !errors.Is(err, ErrMustNotEqual) {
			t.Fatalf("expected ErrMustNotEqual, got %v", err)
		}
	})
}

func TestEqualIgnoreCase(t *testing.T) {
	t.Parallel()

	t.Run("happy path ascii", func(t *testing.T) {
		t.Parallel()

		err := EqualIgnoreCase("Hello", "hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path unicode fold", func(t *testing.T) {
		t.Parallel()

		err := EqualIgnoreCase("Σίγμα", "σίγμα")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path both empty", func(t *testing.T) {
		t.Parallel()

		err := EqualIgnoreCase("", "")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error different", func(t *testing.T) {
		t.Parallel()

		err := EqualIgnoreCase("hello", "world")
		if !errors.Is(err, ErrNotEqual) {
			t.Fatalf("expected ErrNotEqual, got %v", err)
		}
	})

	t.Run("error empty vs non-empty", func(t *testing.T) {
		t.Parallel()

		err := EqualIgnoreCase("", "x")
		if !errors.Is(err, ErrNotEqual) {
			t.Fatalf("expected ErrNotEqual, got %v", err)
		}
	})
}

func TestOneOf(t *testing.T) {
	t.Parallel()

	t.Run("happy path string", func(t *testing.T) {
		t.Parallel()

		err := OneOf("admin", []string{"admin", "user", "guest"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path int", func(t *testing.T) {
		t.Parallel()

		err := OneOf(2, []int{1, 2, 3})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path custom type", func(t *testing.T) {
		t.Parallel()

		err := OneOf(role("admin"), []role{"admin", "user"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error not in", func(t *testing.T) {
		t.Parallel()

		err := OneOf("root", []string{"admin", "user"})
		if !errors.Is(err, ErrNotInAllowed) {
			t.Fatalf("expected ErrNotInAllowed, got %v", err)
		}
	})

	t.Run("error empty allowed", func(t *testing.T) {
		t.Parallel()

		err := OneOf("admin", []string{})
		if !errors.Is(err, ErrEmptyAllowed) {
			t.Fatalf("expected ErrEmptyAllowed, got %v", err)
		}
	})
}

func TestNotIn(t *testing.T) {
	t.Parallel()

	t.Run("happy path not present", func(t *testing.T) {
		t.Parallel()

		err := NotIn("admin", []string{"banned", "blocked"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path empty forbidden", func(t *testing.T) {
		t.Parallel()

		err := NotIn("admin", []string{})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error in forbidden", func(t *testing.T) {
		t.Parallel()

		err := NotIn("banned", []string{"banned", "blocked"})
		if !errors.Is(err, ErrInForbidden) {
			t.Fatalf("expected ErrInForbidden, got %v", err)
		}
	})

	t.Run("error int in forbidden", func(t *testing.T) {
		t.Parallel()

		err := NotIn(7, []int{7, 13})
		if !errors.Is(err, ErrInForbidden) {
			t.Fatalf("expected ErrInForbidden, got %v", err)
		}
	})
}

func TestIsJWT(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		err := IsJWT(jwt)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsJWT("")
		if !errors.Is(err, ErrJWTInvalid) {
			t.Fatalf("expected ErrJWTInvalid, got %v", err)
		}
	})

	t.Run("error wrong segment count", func(t *testing.T) {
		t.Parallel()

		err := IsJWT("only.two")
		if !errors.Is(err, ErrJWTInvalid) {
			t.Fatalf("expected ErrJWTInvalid, got %v", err)
		}
	})

	t.Run("error empty segment", func(t *testing.T) {
		t.Parallel()

		err := IsJWT("eyJhbGciOiJIUzI1NiJ9..signature")
		if !errors.Is(err, ErrJWTInvalid) {
			t.Fatalf("expected ErrJWTInvalid, got %v", err)
		}
	})

	t.Run("error header not base64url", func(t *testing.T) {
		t.Parallel()

		err := IsJWT("!!!.eyJzdWIiOiIxIn0.sig")
		if !errors.Is(err, ErrJWTInvalid) {
			t.Fatalf("expected ErrJWTInvalid, got %v", err)
		}
	})

	t.Run("error payload not json object", func(t *testing.T) {
		t.Parallel()

		err := IsJWT("eyJhbGciOiJIUzI1NiJ9.bm90LWpzb24.sig")
		if !errors.Is(err, ErrJWTInvalid) {
			t.Fatalf("expected ErrJWTInvalid, got %v", err)
		}
	})
}

func TestIsSemver(t *testing.T) {
	t.Parallel()

	t.Run("happy path basic", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("1.2.3")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path pre-release", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("1.0.0-alpha.1")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path build metadata", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("1.0.0+20130313144700")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path pre-release and build", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("1.0.0-beta+exp.sha.5114f85")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("")
		if !errors.Is(err, ErrSemverInvalid) {
			t.Fatalf("expected ErrSemverInvalid, got %v", err)
		}
	})

	t.Run("error v prefix", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("v1.2.3")
		if !errors.Is(err, ErrSemverInvalid) {
			t.Fatalf("expected ErrSemverInvalid, got %v", err)
		}
	})

	t.Run("error missing patch", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("1.2")
		if !errors.Is(err, ErrSemverInvalid) {
			t.Fatalf("expected ErrSemverInvalid, got %v", err)
		}
	})

	t.Run("error leading zero", func(t *testing.T) {
		t.Parallel()

		err := IsSemver("01.2.3")
		if !errors.Is(err, ErrSemverInvalid) {
			t.Fatalf("expected ErrSemverInvalid, got %v", err)
		}
	})
}

func TestIsIP(t *testing.T) {
	t.Parallel()

	t.Run("happy path ipv4", func(t *testing.T) {
		t.Parallel()

		err := IsIP("192.0.2.1")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path ipv6", func(t *testing.T) {
		t.Parallel()

		err := IsIP("2001:db8::1")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error garbage", func(t *testing.T) {
		t.Parallel()

		err := IsIP("not-an-ip")
		if !errors.Is(err, ErrIPInvalid) {
			t.Fatalf("expected ErrIPInvalid, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsIP("")
		if !errors.Is(err, ErrIPInvalid) {
			t.Fatalf("expected ErrIPInvalid, got %v", err)
		}
	})
}

func TestIsIPv4(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsIPv4("192.0.2.1")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error ipv6", func(t *testing.T) {
		t.Parallel()

		err := IsIPv4("2001:db8::1")
		if !errors.Is(err, ErrIPv4Invalid) {
			t.Fatalf("expected ErrIPv4Invalid, got %v", err)
		}
	})

	t.Run("error ipv4-mapped ipv6", func(t *testing.T) {
		t.Parallel()

		err := IsIPv4("::ffff:1.2.3.4")
		if !errors.Is(err, ErrIPv4Invalid) {
			t.Fatalf("expected ErrIPv4Invalid, got %v", err)
		}
	})

	t.Run("error garbage", func(t *testing.T) {
		t.Parallel()

		err := IsIPv4("999.999.999.999")
		if !errors.Is(err, ErrIPv4Invalid) {
			t.Fatalf("expected ErrIPv4Invalid, got %v", err)
		}
	})
}

func TestIsIPv6(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsIPv6("2001:db8::1")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error ipv4", func(t *testing.T) {
		t.Parallel()

		err := IsIPv6("192.0.2.1")
		if !errors.Is(err, ErrIPv6Invalid) {
			t.Fatalf("expected ErrIPv6Invalid, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsIPv6("")
		if !errors.Is(err, ErrIPv6Invalid) {
			t.Fatalf("expected ErrIPv6Invalid, got %v", err)
		}
	})
}

func TestIsCIDR(t *testing.T) {
	t.Parallel()

	t.Run("happy path ipv4", func(t *testing.T) {
		t.Parallel()

		err := IsCIDR("192.0.2.0/24")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path ipv6", func(t *testing.T) {
		t.Parallel()

		err := IsCIDR("2001:db8::/32")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error missing mask", func(t *testing.T) {
		t.Parallel()

		err := IsCIDR("192.0.2.0")
		if !errors.Is(err, ErrCIDRInvalid) {
			t.Fatalf("expected ErrCIDRInvalid, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsCIDR("")
		if !errors.Is(err, ErrCIDRInvalid) {
			t.Fatalf("expected ErrCIDRInvalid, got %v", err)
		}
	})
}

func TestIsMAC(t *testing.T) {
	t.Parallel()

	t.Run("happy path colons", func(t *testing.T) {
		t.Parallel()

		err := IsMAC("00:1a:2b:3c:4d:5e")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path hyphens", func(t *testing.T) {
		t.Parallel()

		err := IsMAC("00-1a-2b-3c-4d-5e")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error garbage", func(t *testing.T) {
		t.Parallel()

		err := IsMAC("not-a-mac")
		if !errors.Is(err, ErrMACInvalid) {
			t.Fatalf("expected ErrMACInvalid, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsMAC("")
		if !errors.Is(err, ErrMACInvalid) {
			t.Fatalf("expected ErrMACInvalid, got %v", err)
		}
	})
}

func TestIsHostname(t *testing.T) {
	t.Parallel()

	t.Run("happy path single label", func(t *testing.T) {
		t.Parallel()

		err := IsHostname("server01")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path multi label", func(t *testing.T) {
		t.Parallel()

		err := IsHostname("api.example.com")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsHostname("")
		if !errors.Is(err, ErrHostnameInvalid) {
			t.Fatalf("expected ErrHostnameInvalid, got %v", err)
		}
	})

	t.Run("error leading hyphen", func(t *testing.T) {
		t.Parallel()

		err := IsHostname("-server")
		if !errors.Is(err, ErrHostnameInvalid) {
			t.Fatalf("expected ErrHostnameInvalid, got %v", err)
		}
	})

	t.Run("error too long", func(t *testing.T) {
		t.Parallel()

		long := strings.Repeat("a", 254)

		err := IsHostname(long)
		if !errors.Is(err, ErrHostnameInvalid) {
			t.Fatalf("expected ErrHostnameInvalid, got %v", err)
		}
	})

	t.Run("error label too long", func(t *testing.T) {
		t.Parallel()

		long := strings.Repeat("a", 64) + ".com"

		err := IsHostname(long)
		if !errors.Is(err, ErrHostnameInvalid) {
			t.Fatalf("expected ErrHostnameInvalid, got %v", err)
		}
	})
}

func TestIsFQDN(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsFQDN("api.example.com")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error single label", func(t *testing.T) {
		t.Parallel()

		err := IsFQDN("localhost")
		if !errors.Is(err, ErrFQDNInvalid) {
			t.Fatalf("expected ErrFQDNInvalid, got %v", err)
		}
	})

	t.Run("error numeric tld", func(t *testing.T) {
		t.Parallel()

		err := IsFQDN("192.0.2.1")
		if !errors.Is(err, ErrFQDNInvalid) {
			t.Fatalf("expected ErrFQDNInvalid, got %v", err)
		}
	})

	t.Run("error not a hostname", func(t *testing.T) {
		t.Parallel()

		err := IsFQDN("not valid")
		if !errors.Is(err, ErrFQDNInvalid) {
			t.Fatalf("expected ErrFQDNInvalid, got %v", err)
		}
	})
}

func TestIsPort(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsPort(443)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path uint16", func(t *testing.T) {
		t.Parallel()

		err := IsPort(uint16(8080))
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path edge low", func(t *testing.T) {
		t.Parallel()

		err := IsPort(1)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path edge high", func(t *testing.T) {
		t.Parallel()

		err := IsPort(65535)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error zero", func(t *testing.T) {
		t.Parallel()

		err := IsPort(0)
		if !errors.Is(err, ErrPortInvalid) {
			t.Fatalf("expected ErrPortInvalid, got %v", err)
		}
	})

	t.Run("error above max", func(t *testing.T) {
		t.Parallel()

		err := IsPort(70000)
		if !errors.Is(err, ErrPortInvalid) {
			t.Fatalf("expected ErrPortInvalid, got %v", err)
		}
	})

	t.Run("error negative", func(t *testing.T) {
		t.Parallel()

		err := IsPort(-1)
		if !errors.Is(err, ErrPortInvalid) {
			t.Fatalf("expected ErrPortInvalid, got %v", err)
		}
	})
}

func TestIsRFC3339(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsRFC3339("2026-05-24T12:34:56Z")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path offset", func(t *testing.T) {
		t.Parallel()

		err := IsRFC3339("2026-05-24T12:34:56-05:00")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsRFC3339("")
		if !errors.Is(err, ErrDateInvalid) {
			t.Fatalf("expected ErrDateInvalid, got %v", err)
		}
	})

	t.Run("error not rfc3339", func(t *testing.T) {
		t.Parallel()

		err := IsRFC3339("2026/05/24")
		if !errors.Is(err, ErrDateInvalid) {
			t.Fatalf("expected ErrDateInvalid, got %v", err)
		}
	})
}

func TestIsDate(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsDate("2026-05-24", "2006-01-02")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty layout", func(t *testing.T) {
		t.Parallel()

		err := IsDate("2026-05-24", "")
		if !errors.Is(err, ErrLayoutInvalid) {
			t.Fatalf("expected ErrLayoutInvalid, got %v", err)
		}
	})

	t.Run("error empty input", func(t *testing.T) {
		t.Parallel()

		err := IsDate("", "2006-01-02")
		if !errors.Is(err, ErrDateInvalid) {
			t.Fatalf("expected ErrDateInvalid, got %v", err)
		}
	})

	t.Run("error mismatch", func(t *testing.T) {
		t.Parallel()

		err := IsDate("not-a-date", "2006-01-02")
		if !errors.Is(err, ErrDateInvalid) {
			t.Fatalf("expected ErrDateInvalid, got %v", err)
		}
	})
}

func TestBefore(t *testing.T) {
	t.Parallel()

	ref := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := Before(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), ref)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error equal", func(t *testing.T) {
		t.Parallel()

		err := Before(ref, ref)
		if !errors.Is(err, ErrTimeBefore) {
			t.Fatalf("expected ErrTimeBefore, got %v", err)
		}
	})

	t.Run("error after", func(t *testing.T) {
		t.Parallel()

		err := Before(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), ref)
		if !errors.Is(err, ErrTimeBefore) {
			t.Fatalf("expected ErrTimeBefore, got %v", err)
		}
	})
}

func TestAfter(t *testing.T) {
	t.Parallel()

	ref := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := After(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), ref)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error equal", func(t *testing.T) {
		t.Parallel()

		err := After(ref, ref)
		if !errors.Is(err, ErrTimeAfter) {
			t.Fatalf("expected ErrTimeAfter, got %v", err)
		}
	})

	t.Run("error before", func(t *testing.T) {
		t.Parallel()

		err := After(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), ref)
		if !errors.Is(err, ErrTimeAfter) {
			t.Fatalf("expected ErrTimeAfter, got %v", err)
		}
	})
}

func TestBetweenTime(t *testing.T) {
	t.Parallel()

	lo := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	hi := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := BetweenTime(time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), lo, hi)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path inclusive lo", func(t *testing.T) {
		t.Parallel()

		err := BetweenTime(lo, lo, hi)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error before lo", func(t *testing.T) {
		t.Parallel()

		err := BetweenTime(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), lo, hi)
		if !errors.Is(err, ErrTimeOutOfRange) {
			t.Fatalf("expected ErrTimeOutOfRange, got %v", err)
		}
	})

	t.Run("error after hi", func(t *testing.T) {
		t.Parallel()

		err := BetweenTime(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), lo, hi)
		if !errors.Is(err, ErrTimeOutOfRange) {
			t.Fatalf("expected ErrTimeOutOfRange, got %v", err)
		}
	})

	t.Run("error invalid range", func(t *testing.T) {
		t.Parallel()

		err := BetweenTime(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), hi, lo)
		if !errors.Is(err, ErrInvalidTimeRange) {
			t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
		}
	})
}

func TestMinCount(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MinCount([]int{1, 2, 3}, 2)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := MinCount([]int{1}, 3)
		if !errors.Is(err, ErrCountBelowMin) {
			t.Fatalf("expected ErrCountBelowMin, got %v", err)
		}
	})

	t.Run("negative threshold passes empty", func(t *testing.T) {
		t.Parallel()

		err := MinCount([]int{}, -3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestMaxCount(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MaxCount([]int{1, 2}, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := MaxCount([]int{1, 2, 3, 4, 5}, 3)
		if !errors.Is(err, ErrCountAboveMax) {
			t.Fatalf("expected ErrCountAboveMax, got %v", err)
		}
	})

	t.Run("negative threshold rejects non-empty", func(t *testing.T) {
		t.Parallel()

		err := MaxCount([]int{1}, -1)
		if !errors.Is(err, ErrCountAboveMax) {
			t.Fatalf("expected ErrCountAboveMax, got %v", err)
		}
	})
}

func TestCountInRange(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := CountInRange([]int{1, 2, 3}, 2, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := CountInRange([]int{1}, 2, 5)
		if !errors.Is(err, ErrCountOutOfRange) {
			t.Fatalf("expected ErrCountOutOfRange, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := CountInRange([]int{1, 2, 3, 4, 5, 6}, 2, 5)
		if !errors.Is(err, ErrCountOutOfRange) {
			t.Fatalf("expected ErrCountOutOfRange, got %v", err)
		}
	})

	t.Run("error invalid range", func(t *testing.T) {
		t.Parallel()

		err := CountInRange([]int{1, 2}, 5, 2)
		if !errors.Is(err, ErrInvalidCountRange) {
			t.Fatalf("expected ErrInvalidCountRange, got %v", err)
		}
	})
}

func TestUnique(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := Unique([]int{1, 2, 3, 4})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path empty", func(t *testing.T) {
		t.Parallel()

		err := Unique([]int{})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error duplicate start", func(t *testing.T) {
		t.Parallel()

		err := Unique([]int{1, 1, 2, 3})
		if !errors.Is(err, ErrDuplicate) {
			t.Fatalf("expected ErrDuplicate, got %v", err)
		}
	})

	t.Run("error duplicate middle", func(t *testing.T) {
		t.Parallel()

		err := Unique([]int{1, 2, 2, 3})
		if !errors.Is(err, ErrDuplicate) {
			t.Fatalf("expected ErrDuplicate, got %v", err)
		}
	})

	t.Run("error duplicate end", func(t *testing.T) {
		t.Parallel()

		err := Unique([]int{1, 2, 3, 3})
		if !errors.Is(err, ErrDuplicate) {
			t.Fatalf("expected ErrDuplicate, got %v", err)
		}
	})
}

func TestSortedAsc(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := SortedAsc([]int{1, 2, 3, 4})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path equal adjacent", func(t *testing.T) {
		t.Parallel()

		err := SortedAsc([]int{1, 2, 2, 3})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path empty", func(t *testing.T) {
		t.Parallel()

		err := SortedAsc([]int{})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path single", func(t *testing.T) {
		t.Parallel()

		err := SortedAsc([]int{42})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error not sorted", func(t *testing.T) {
		t.Parallel()

		err := SortedAsc([]int{3, 1, 2})
		if !errors.Is(err, ErrNotSortedAsc) {
			t.Fatalf("expected ErrNotSortedAsc, got %v", err)
		}
	})
}

func TestSortedDesc(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := SortedDesc([]int{4, 3, 2, 1})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path equal adjacent", func(t *testing.T) {
		t.Parallel()

		err := SortedDesc([]int{3, 2, 2, 1})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path empty", func(t *testing.T) {
		t.Parallel()

		err := SortedDesc([]int{})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error not sorted", func(t *testing.T) {
		t.Parallel()

		err := SortedDesc([]int{1, 3, 2})
		if !errors.Is(err, ErrNotSortedDesc) {
			t.Fatalf("expected ErrNotSortedDesc, got %v", err)
		}
	})
}

func TestHasKey(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := HasKey(map[string]int{"a": 1, "b": 2}, "a")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error missing", func(t *testing.T) {
		t.Parallel()

		err := HasKey(map[string]int{"a": 1}, "missing")
		if !errors.Is(err, ErrKeyMissing) {
			t.Fatalf("expected ErrKeyMissing, got %v", err)
		}
	})

	t.Run("error nil map", func(t *testing.T) {
		t.Parallel()

		var m map[string]int

		err := HasKey(m, "any")
		if !errors.Is(err, ErrKeyMissing) {
			t.Fatalf("expected ErrKeyMissing, got %v", err)
		}
	})

	t.Run("error empty map", func(t *testing.T) {
		t.Parallel()

		err := HasKey(map[string]int{}, "any")
		if !errors.Is(err, ErrKeyMissing) {
			t.Fatalf("expected ErrKeyMissing, got %v", err)
		}
	})
}

func TestMinKeys(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MinKeys(map[string]int{"a": 1, "b": 2, "c": 3}, 2)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path zero threshold", func(t *testing.T) {
		t.Parallel()

		err := MinKeys(map[string]int{}, 0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path negative threshold", func(t *testing.T) {
		t.Parallel()

		err := MinKeys(map[string]int{}, -3)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error below", func(t *testing.T) {
		t.Parallel()

		err := MinKeys(map[string]int{"a": 1}, 3)
		if !errors.Is(err, ErrMinKeys) {
			t.Fatalf("expected ErrMinKeys, got %v", err)
		}
	})

	t.Run("error nil map", func(t *testing.T) {
		t.Parallel()

		var m map[string]int

		err := MinKeys(m, 1)
		if !errors.Is(err, ErrMinKeys) {
			t.Fatalf("expected ErrMinKeys, got %v", err)
		}
	})
}

func TestMaxKeys(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := MaxKeys(map[string]int{"a": 1, "b": 2}, 5)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path zero threshold empty map", func(t *testing.T) {
		t.Parallel()

		err := MaxKeys(map[string]int{}, 0)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error above", func(t *testing.T) {
		t.Parallel()

		err := MaxKeys(map[string]int{"a": 1, "b": 2, "c": 3}, 2)
		if !errors.Is(err, ErrMaxKeys) {
			t.Fatalf("expected ErrMaxKeys, got %v", err)
		}
	})

	t.Run("error negative threshold rejects non-empty", func(t *testing.T) {
		t.Parallel()

		err := MaxKeys(map[string]int{"a": 1}, -1)
		if !errors.Is(err, ErrMaxKeys) {
			t.Fatalf("expected ErrMaxKeys, got %v", err)
		}
	})
}

func TestOptional(t *testing.T) {
	t.Parallel()

	t.Run("passes on zero value", func(t *testing.T) {
		t.Parallel()

		err := Optional[string](IsEmail)("")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("delegates on non-empty pass", func(t *testing.T) {
		t.Parallel()

		err := Optional[string](IsEmail)("a@b.com")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("delegates on non-empty fail", func(t *testing.T) {
		t.Parallel()

		err := Optional[string](IsEmail)("not-an-email")
		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})
}

func TestAnyOf(t *testing.T) {
	t.Parallel()

	t.Run("passes when any passes", func(t *testing.T) {
		t.Parallel()

		err := AnyOf[string](IsURL, IsEmail)("a@b.com")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("aggregates all violations when every check fails", func(t *testing.T) {
		t.Parallel()

		err := AnyOf[string](IsURL, IsEmail)("not-valid")
		if !errors.Is(err, ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid wrapped, got %v", err)
		}

		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid wrapped, got %v", err)
		}
	})

	t.Run("empty checks trivially passes", func(t *testing.T) {
		t.Parallel()

		err := AnyOf[string]()("anything")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestAllOf(t *testing.T) {
	t.Parallel()

	t.Run("passes when all pass", func(t *testing.T) {
		t.Parallel()

		check := AllOf[string](func(s string) error { return MinLen(s, 3) }, IsAlpha)

		err := check("hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("aggregates all violations", func(t *testing.T) {
		t.Parallel()

		check := AllOf[string](func(s string) error { return MinLen(s, 5) }, IsAlpha)

		err := check("a1")
		if !errors.Is(err, ErrMinLen) {
			t.Fatalf("expected ErrMinLen wrapped, got %v", err)
		}

		if !errors.Is(err, ErrNotAlpha) {
			t.Fatalf("expected ErrNotAlpha wrapped, got %v", err)
		}
	})

	t.Run("empty checks trivially passes", func(t *testing.T) {
		t.Parallel()

		err := AllOf[string]()("anything")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestNot(t *testing.T) {
	t.Parallel()

	t.Run("passes when inner fails", func(t *testing.T) {
		t.Parallel()

		err := Not[string](IsEmail)("not-an-email")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error when inner passes", func(t *testing.T) {
		t.Parallel()

		err := Not[string](IsEmail)("a@b.com")
		if !errors.Is(err, ErrAssertionInverted) {
			t.Fatalf("expected ErrAssertionInverted, got %v", err)
		}
	})
}

func TestNonEmpty(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := NonEmpty([]int{1, 2, 3})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := NonEmpty([]int{})
		if !errors.Is(err, ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})

	t.Run("error nil slice", func(t *testing.T) {
		t.Parallel()

		var xs []string

		err := NonEmpty(xs)
		if !errors.Is(err, ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})
}

func TestEach(t *testing.T) {
	t.Parallel()

	t.Run("happy path all pass", func(t *testing.T) {
		t.Parallel()

		err := Each([]string{"a@b.com", "c@d.com"}, IsEmail)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error one fails", func(t *testing.T) {
		t.Parallel()

		err := Each([]string{"a@b.com", "not-an-email"}, IsEmail)
		if !errors.Is(err, ErrEachFailed) {
			t.Fatalf("expected ErrEachFailed, got %v", err)
		}

		if !errors.Is(err, ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid to be wrapped, got %v", err)
		}
	})

	t.Run("nil check no-ops", func(t *testing.T) {
		t.Parallel()

		err := Each([]int{1, 2, 3}, nil)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("empty slice trivially passes", func(t *testing.T) {
		t.Parallel()

		err := Each([]int{}, func(int) error { return errors.New("never called") })
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestContains(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := Contains("hello world", "world")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error missing", func(t *testing.T) {
		t.Parallel()

		err := Contains("hello", "world")
		if !errors.Is(err, ErrContainsMissing) {
			t.Fatalf("expected ErrContainsMissing, got %v", err)
		}
	})

	t.Run("empty substr passes", func(t *testing.T) {
		t.Parallel()

		err := Contains("hello", "")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("empty s rejects non-empty substr", func(t *testing.T) {
		t.Parallel()

		err := Contains("", "x")
		if !errors.Is(err, ErrContainsMissing) {
			t.Fatalf("expected ErrContainsMissing, got %v", err)
		}
	})
}

func TestHasPrefix(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := HasPrefix("hello world", "hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error missing", func(t *testing.T) {
		t.Parallel()

		err := HasPrefix("hello", "world")
		if !errors.Is(err, ErrPrefixMissing) {
			t.Fatalf("expected ErrPrefixMissing, got %v", err)
		}
	})

	t.Run("empty prefix passes", func(t *testing.T) {
		t.Parallel()

		err := HasPrefix("hello", "")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestHasSuffix(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := HasSuffix("hello world", "world")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error missing", func(t *testing.T) {
		t.Parallel()

		err := HasSuffix("hello", "world")
		if !errors.Is(err, ErrSuffixMissing) {
			t.Fatalf("expected ErrSuffixMissing, got %v", err)
		}
	})

	t.Run("empty suffix passes", func(t *testing.T) {
		t.Parallel()

		err := HasSuffix("hello", "")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
}

func TestIsLowercase(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsLowercase("hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path with digits", func(t *testing.T) {
		t.Parallel()

		err := IsLowercase("hello123")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("empty passes", func(t *testing.T) {
		t.Parallel()

		err := IsLowercase("")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error uppercase", func(t *testing.T) {
		t.Parallel()

		err := IsLowercase("Hello")
		if !errors.Is(err, ErrNotLowercase) {
			t.Fatalf("expected ErrNotLowercase, got %v", err)
		}
	})
}

func TestIsUppercase(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsUppercase("HELLO")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path with digits", func(t *testing.T) {
		t.Parallel()

		err := IsUppercase("HELLO123")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("empty passes", func(t *testing.T) {
		t.Parallel()

		err := IsUppercase("")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error lowercase", func(t *testing.T) {
		t.Parallel()

		err := IsUppercase("Hello")
		if !errors.Is(err, ErrNotUppercase) {
			t.Fatalf("expected ErrNotUppercase, got %v", err)
		}
	})
}

func TestIsAlpha(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsAlpha("hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path unicode", func(t *testing.T) {
		t.Parallel()

		err := IsAlpha("héllo")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error with digit", func(t *testing.T) {
		t.Parallel()

		err := IsAlpha("hello1")
		if !errors.Is(err, ErrNotAlpha) {
			t.Fatalf("expected ErrNotAlpha, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsAlpha("")
		if !errors.Is(err, ErrNotAlpha) {
			t.Fatalf("expected ErrNotAlpha, got %v", err)
		}
	})
}

func TestIsAlphanumeric(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsAlphanumeric("hello123")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error with symbol", func(t *testing.T) {
		t.Parallel()

		err := IsAlphanumeric("hello!")
		if !errors.Is(err, ErrNotAlphanumeric) {
			t.Fatalf("expected ErrNotAlphanumeric, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsAlphanumeric("")
		if !errors.Is(err, ErrNotAlphanumeric) {
			t.Fatalf("expected ErrNotAlphanumeric, got %v", err)
		}
	})
}

func TestIsNumeric(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsNumeric("12345")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error with letter", func(t *testing.T) {
		t.Parallel()

		err := IsNumeric("12a45")
		if !errors.Is(err, ErrNotNumeric) {
			t.Fatalf("expected ErrNotNumeric, got %v", err)
		}
	})

	t.Run("error with sign", func(t *testing.T) {
		t.Parallel()

		err := IsNumeric("-123")
		if !errors.Is(err, ErrNotNumeric) {
			t.Fatalf("expected ErrNotNumeric, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsNumeric("")
		if !errors.Is(err, ErrNotNumeric) {
			t.Fatalf("expected ErrNotNumeric, got %v", err)
		}
	})
}

func TestIsASCII(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsASCII("hello world 123!")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error non-ASCII", func(t *testing.T) {
		t.Parallel()

		err := IsASCII("héllo")
		if !errors.Is(err, ErrNotASCII) {
			t.Fatalf("expected ErrNotASCII, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsASCII("")
		if !errors.Is(err, ErrNotASCII) {
			t.Fatalf("expected ErrNotASCII, got %v", err)
		}
	})
}

func TestIsHex(t *testing.T) {
	t.Parallel()

	t.Run("happy path lower", func(t *testing.T) {
		t.Parallel()

		err := IsHex("deadbeef")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path upper", func(t *testing.T) {
		t.Parallel()

		err := IsHex("DEADBEEF")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("happy path mixed", func(t *testing.T) {
		t.Parallel()

		err := IsHex("DeadBeef123")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error non-hex char", func(t *testing.T) {
		t.Parallel()

		err := IsHex("xyz")
		if !errors.Is(err, ErrNotHex) {
			t.Fatalf("expected ErrNotHex, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsHex("")
		if !errors.Is(err, ErrNotHex) {
			t.Fatalf("expected ErrNotHex, got %v", err)
		}
	})
}

func TestIsBase64(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsBase64("aGVsbG8gd29ybGQ=")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error invalid char", func(t *testing.T) {
		t.Parallel()

		err := IsBase64("not base64!")
		if !errors.Is(err, ErrBase64Invalid) {
			t.Fatalf("expected ErrBase64Invalid, got %v", err)
		}
	})

	t.Run("error empty", func(t *testing.T) {
		t.Parallel()

		err := IsBase64("")
		if !errors.Is(err, ErrBase64Invalid) {
			t.Fatalf("expected ErrBase64Invalid, got %v", err)
		}
	})
}

func TestIsTrimmed(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		err := IsTrimmed("hello")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("empty passes", func(t *testing.T) {
		t.Parallel()

		err := IsTrimmed("")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("error leading space", func(t *testing.T) {
		t.Parallel()

		err := IsTrimmed(" hello")
		if !errors.Is(err, ErrNotTrimmed) {
			t.Fatalf("expected ErrNotTrimmed, got %v", err)
		}
	})

	t.Run("error trailing tab", func(t *testing.T) {
		t.Parallel()

		err := IsTrimmed("hello\t")
		if !errors.Is(err, ErrNotTrimmed) {
			t.Fatalf("expected ErrNotTrimmed, got %v", err)
		}
	})
}

type owner struct {
	Email string
	Tags  []string
}

type pokemon struct {
	Name  string
	Owner owner
	IDs   []int
}

func TestGetField(t *testing.T) {
	t.Parallel()

	t.Run("dotted path struct", func(t *testing.T) {
		t.Parallel()

		p := pokemon{Name: "pikachu", Owner: owner{Email: "ash@kanto.com"}}

		v, err := GetField(p, "Owner.Email")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "ash@kanto.com" {
			t.Fatalf("expected ash@kanto.com, got %v", v)
		}
	})

	t.Run("slice index", func(t *testing.T) {
		t.Parallel()

		p := pokemon{IDs: []int{10, 20, 30}}

		v, err := GetField(p, "IDs[1]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(int)
		if !ok || got != 20 {
			t.Fatalf("expected 20, got %v", v)
		}
	})

	t.Run("slice index nested", func(t *testing.T) {
		t.Parallel()

		p := pokemon{Owner: owner{Tags: []string{"trainer", "champion"}}}

		v, err := GetField(p, "Owner.Tags[0]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "trainer" {
			t.Fatalf("expected trainer, got %v", v)
		}
	})

	t.Run("pointer auto-deref", func(t *testing.T) {
		t.Parallel()

		p := &pokemon{Name: "snorlax"}

		v, err := GetField(p, "Name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "snorlax" {
			t.Fatalf("expected snorlax, got %v", v)
		}
	})

	t.Run("map lookup", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"key": "value"}

		v, err := GetField(m, "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(string)
		if !ok || got != "value" {
			t.Fatalf("expected value, got %v", v)
		}
	})

	t.Run("error nil object", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(nil, "X")
		if !errors.Is(err, ErrObjectNil) {
			t.Fatalf("expected ErrObjectNil, got %v", err)
		}
	})

	t.Run("error empty path", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error missing field", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "Nope")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error type mismatch", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{Name: "x"}, "Name.Inner")
		if !errors.Is(err, ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error index out of range", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{IDs: []int{1}}, "IDs[5]")
		if !errors.Is(err, ErrIndexOutOfRange) {
			t.Fatalf("expected ErrIndexOutOfRange, got %v", err)
		}
	})

	t.Run("error index on non-slice", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{Name: "x"}, "Name[0]")
		if !errors.Is(err, ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error malformed bracket", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error empty index", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error non-numeric index", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[abc]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error negative index", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[-1]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error leading bracket", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "[0]")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error nil pointer in path", func(t *testing.T) {
		t.Parallel()

		type wrapper struct {
			P *pokemon
		}

		_, err := GetField(wrapper{}, "P.Name")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error map non-string key", func(t *testing.T) {
		t.Parallel()

		m := map[int]string{1: "x"}

		_, err := GetField(m, "1")
		if !errors.Is(err, ErrPathTypeMismatch) {
			t.Fatalf("expected ErrPathTypeMismatch, got %v", err)
		}
	})

	t.Run("error map missing key", func(t *testing.T) {
		t.Parallel()

		m := map[string]any{"a": 1}

		_, err := GetField(m, "missing")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("error trailing chars after bracket", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "IDs[0]junk")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error empty dotted segment", func(t *testing.T) {
		t.Parallel()

		_, err := GetField(pokemon{}, "Owner..Email")
		if !errors.Is(err, ErrPathInvalid) {
			t.Fatalf("expected ErrPathInvalid, got %v", err)
		}
	})

	t.Run("error nil interface deref before index", func(t *testing.T) {
		t.Parallel()

		type box struct {
			V any
		}

		_, err := GetField(box{V: nil}, "V[0]")
		if !errors.Is(err, ErrPathNotFound) {
			t.Fatalf("expected ErrPathNotFound, got %v", err)
		}
	})

	t.Run("multiple indices", func(t *testing.T) {
		t.Parallel()

		type matrix struct {
			Rows [][]int
		}

		m := matrix{Rows: [][]int{{1, 2}, {3, 4}}}

		v, err := GetField(m, "Rows[1][0]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := v.(int)
		if !ok || got != 3 {
			t.Fatalf("expected 3, got %v", v)
		}
	})
}

func TestErrValidation(t *testing.T) {
	t.Parallel()

	t.Run("wraps causes", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("inner")
		err := ErrValidation(inner)

		if !errors.Is(err, ErrValidationFailed) {
			t.Fatalf("expected ErrValidationFailed, got %v", err)
		}

		if !errors.Is(err, inner) {
			t.Fatalf("expected inner cause wrapped, got %v", err)
		}
	})

	t.Run("error string format", func(t *testing.T) {
		t.Parallel()

		err := ErrValidation(ErrFieldRequired)
		msg := err.Error()
		if !strings.Contains(msg, "validation") {
			t.Fatalf("expected message to contain validation, got %q", msg)
		}
	})
}
