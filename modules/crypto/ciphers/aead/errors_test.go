package aead

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
		if !strings.Contains(got, "aead") {
			t.Fatalf("expected 'aead' in error, got %q", got)
		}

		if !strings.Contains(got, AeadMethod) {
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

	t.Run("wraps with ErrEncryptFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrEncryption(errors.New("cause"))

		if !errors.Is(err, ErrEncryptFailed) {
			t.Fatal("expected ErrEncryptFailed")
		}
	})
}

func TestErrDecryption(t *testing.T) {
	t.Parallel()

	t.Run("wraps with ErrDecryptFailed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrDecryption(errors.New("cause"))

		if !errors.Is(err, ErrDecryptFailed) {
			t.Fatal("expected ErrDecryptFailed")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrMethodInvalid is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrMethodInvalid == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrKeyInvalid is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrKeyInvalid == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrCipherInitFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCipherInitFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrCiphertextTooShort is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCiphertextTooShort == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrKeySizeInvalid is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrKeySizeInvalid == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrNonceSizeInvalid is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrNonceSizeInvalid == nil {
			t.Fatal("expected non-nil")
		}
	})
}
