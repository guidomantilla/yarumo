package rsaoaep

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrKeyGeneration(errors.New("cause"))

		got := err.Error()
		if !strings.Contains(got, "rsa_oaep") {
			t.Fatalf("expected 'rsa_oaep' in error, got %q", got)
		}

		if !strings.Contains(got, RsaOaepMethod) {
			t.Fatalf("expected type in error, got %q", got)
		}
	})
}

func TestErrAlgorithmNotSupported(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error with algorithm name", func(t *testing.T) {
		t.Parallel()

		err := ErrAlgorithmNotSupported("UNKNOWN")

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !strings.Contains(err.Error(), "UNKNOWN") {
			t.Fatalf("expected algorithm name, got %q", err.Error())
		}
	})
}

func TestErrKeyGeneration(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrKeyGenerationFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrKeyGeneration(errors.New("cause"))

		if !errors.Is(err, ErrKeyGenerationFailed) {
			t.Fatal("expected ErrKeyGenerationFailed")
		}
	})
}

func TestErrEncryption(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrEncryptionFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrEncryption(errors.New("cause"))

		if !errors.Is(err, ErrEncryptionFailed) {
			t.Fatal("expected ErrEncryptionFailed")
		}
	})
}

func TestErrDecryption(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrDecryptionFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrDecryption(errors.New("cause"))

		if !errors.Is(err, ErrDecryptionFailed) {
			t.Fatal("expected ErrDecryptionFailed")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrMethodIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrMethodIsNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrKeyIsNil is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrKeyIsNil == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrHashNotAvailable is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrHashNotAvailable == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrKeySizeNotAllowed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrKeySizeNotAllowed == nil {
			t.Fatal("expected non-nil")
		}
	})
}
