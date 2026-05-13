package uids

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUID(t *testing.T) {
	t.Parallel()

	t.Run("creates UID with name and function", func(t *testing.T) {
		t.Parallel()

		fn := func() (string, error) { return "test-id", nil }

		u := NewUID("TEST", fn)
		if u == nil {
			t.Fatal("expected non-nil UID")
		}

		if u.Name() != "TEST" {
			t.Fatalf("Name() = %q, want %q", u.Name(), "TEST")
		}

		got, err := u.Generate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != "test-id" {
			t.Fatalf("Generate() = %q, want %q", got, "test-id")
		}
	})

	t.Run("different instances are independent", func(t *testing.T) {
		t.Parallel()

		u1 := NewUID("A", func() (string, error) { return "a", nil })
		u2 := NewUID("B", func() (string, error) { return "b", nil })

		if u1.Name() == u2.Name() {
			t.Fatal("expected different names")
		}

		gotA, errA := u1.Generate()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		gotB, errB := u2.Generate()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if gotA == gotB {
			t.Fatal("expected different generated values")
		}
	})

	t.Run("propagates error from underlying generator", func(t *testing.T) {
		t.Parallel()

		want := errors.New("entropy source failed")
		u := NewUID("FAIL", func() (string, error) { return "", want })

		got, err := u.Generate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, want) {
			t.Fatalf("expected wrapped error, got %v", err)
		}

		if got != "" {
			t.Fatalf("expected empty string on error, got %q", got)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats uid error message", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("ABC")

		got := err.Error()
		if !strings.Contains(got, "uid") {
			t.Fatalf("expected 'uid' in error: %q", got)
		}

		if !strings.Contains(got, "ABC") {
			t.Fatalf("expected 'ABC' in error: %q", got)
		}

		if !strings.Contains(got, UidNotFound) {
			t.Fatalf("expected %q in error: %q", UidNotFound, got)
		}
	})
}

func TestErrAlgorithmNotSupported(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil error", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("XYZ")
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	})

	t.Run("error is of type Error", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("XYZ")

		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("contains algorithm name in message", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("UNKNOWN")
		if !strings.Contains(err.Error(), "UNKNOWN") {
			t.Fatalf("expected algorithm name in error: %q", err.Error())
		}
	})

	t.Run("uses uid algorithm wording", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("FOO")
		if !strings.Contains(err.Error(), "uid algorithm") {
			t.Fatalf("expected 'uid algorithm' in error: %q", err.Error())
		}
	})
}
