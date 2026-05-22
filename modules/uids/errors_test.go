package uids

import (
	"errors"
	"strings"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

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

func TestErrGeneration(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil error", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(errors.New("entropy failed"))
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	})

	t.Run("error is of type Error", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(errors.New("entropy failed"))

		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("matches ErrGenerationFailed sentinel via errors.Is", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(errors.New("entropy failed"))
		if !errors.Is(err, ErrGenerationFailed) {
			t.Fatalf("expected errors.Is(err, ErrGenerationFailed); got false")
		}
	})

	t.Run("preserves wrapped underlying error", func(t *testing.T) {
		t.Parallel()

		underlying := errors.New("entropy failed")
		err := ErrGeneration(underlying)

		if !errors.Is(err, underlying) {
			t.Fatalf("expected errors.Is(err, underlying); got false")
		}
	})

	t.Run("classifies via AsErrorInfo", func(t *testing.T) {
		t.Parallel()

		err := ErrGeneration(errors.New("entropy failed"))
		infos := cerrs.AsErrorInfo(err)

		if len(infos) == 0 {
			t.Fatal("expected at least one ErrorInfo entry")
		}

		found := false

		for _, info := range infos {
			if info.Type == UidGenerationError {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected ErrorInfo with type %q, got %+v", UidGenerationError, infos)
		}
	})

	t.Run("supports multiple wrapped errors", func(t *testing.T) {
		t.Parallel()

		a := errors.New("inner-a")
		b := errors.New("inner-b")
		err := ErrGeneration(a, b)

		if !errors.Is(err, a) {
			t.Fatal("expected err to wrap a")
		}

		if !errors.Is(err, b) {
			t.Fatal("expected err to wrap b")
		}
	})
}
