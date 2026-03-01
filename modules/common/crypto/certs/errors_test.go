package certs

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats error with type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrClientTls(ErrServerNameEmpty)

		got := err.Error()
		if !strings.Contains(got, "certificate") {
			t.Fatalf("expected 'certificate' in error, got %q", got)
		}
		if !strings.Contains(got, CertificateType) {
			t.Fatalf("expected type %q in error, got %q", CertificateType, got)
		}
		if !strings.Contains(got, ErrServerNameEmpty.Error()) {
			t.Fatalf("expected sentinel message in error, got %q", got)
		}
	})
}

func TestErrClientTls(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrClientTls(ErrServerNameEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes sentinel in message", func(t *testing.T) {
		t.Parallel()

		err := ErrClientTls(ErrCaCertificateEmpty)

		if !strings.Contains(err.Error(), ErrCaCertificateEmpty.Error()) {
			t.Fatalf("expected sentinel message in error, got %q", err.Error())
		}
	})

	t.Run("includes tls config failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrClientTls(ErrServerNameEmpty)

		if !strings.Contains(err.Error(), ErrTlsConfigFailed.Error()) {
			t.Fatalf("expected tls config failed in error, got %q", err.Error())
		}
	})

	t.Run("wraps additional cause errors", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("some io error")
		err := ErrClientTls(ErrCaCertificateReadFailed, cause)

		if !strings.Contains(err.Error(), cause.Error()) {
			t.Fatalf("expected cause in error, got %q", err.Error())
		}
	})
}

func TestErrServerTls(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrServerTls(ErrServerCertificateEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes server tls config failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrServerTls(ErrServerCertificateEmpty)

		if !strings.Contains(err.Error(), ErrServerTlsConfigFailed.Error()) {
			t.Fatalf("expected server tls config failed in error, got %q", err.Error())
		}
	})

	t.Run("wraps additional cause errors", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("cert load error")
		err := ErrServerTls(ErrServerCertificateLoadFailed, cause)

		if !strings.Contains(err.Error(), cause.Error()) {
			t.Fatalf("expected cause in error, got %q", err.Error())
		}
	})
}

func TestErrSelfSigned(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrSelfSigned(ErrKeyGenerationFailed)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes self signed failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrSelfSigned(ErrKeyGenerationFailed)

		if !strings.Contains(err.Error(), ErrSelfSignedFailed.Error()) {
			t.Fatalf("expected self signed failed in error, got %q", err.Error())
		}
	})

	t.Run("wraps additional cause errors", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("key gen error")
		err := ErrSelfSigned(ErrKeyGenerationFailed, cause)

		if !strings.Contains(err.Error(), cause.Error()) {
			t.Fatalf("expected cause in error, got %q", err.Error())
		}
	})
}

func TestErrPool(t *testing.T) {
	t.Parallel()

	t.Run("returns domain Error type", func(t *testing.T) {
		t.Parallel()

		err := ErrPool(ErrPoolCertificateEmpty)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("includes pool failed sentinel", func(t *testing.T) {
		t.Parallel()

		err := ErrPool(ErrPoolCertificateEmpty)

		if !strings.Contains(err.Error(), ErrPoolFailed.Error()) {
			t.Fatalf("expected pool failed in error, got %q", err.Error())
		}
	})

	t.Run("wraps additional cause errors", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("parse error")
		err := ErrPool(ErrPoolCertificateParseFailed, cause)

		if !strings.Contains(err.Error(), cause.Error()) {
			t.Fatalf("expected cause in error, got %q", err.Error())
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	// ClientTls sentinels.
	t.Run("ErrServerNameEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerNameEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrCaCertificateEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCaCertificateEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientCertificateEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientCertificateEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientKeyEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientKeyEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrCaCertificateReadFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCaCertificateReadFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrCaCertificateParseFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCaCertificateParseFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientCertificateLoadFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientCertificateLoadFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrTlsConfigFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrTlsConfigFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	// ServerTls sentinels.
	t.Run("ErrServerCertificateEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerCertificateEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrServerKeyEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerKeyEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrServerCertificateLoadFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerCertificateLoadFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientCACertificateReadFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientCACertificateReadFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientCACertificateParseFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientCACertificateParseFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrServerTlsConfigFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerTlsConfigFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	// ClientTlsFromPEM sentinels.
	t.Run("ErrCaCertificatePEMEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCaCertificatePEMEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientCertificatePEMEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientCertificatePEMEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientKeyPEMEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientKeyPEMEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrClientKeyPairFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrClientKeyPairFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	// ServerTlsFromPEM sentinels.
	t.Run("ErrServerCertificatePEMEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerCertificatePEMEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrServerKeyPEMEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerKeyPEMEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrServerKeyPairFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrServerKeyPairFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	// SelfSigned sentinels.
	t.Run("ErrKeyGenerationFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrKeyGenerationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrCertificateCreationFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrCertificateCreationFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrKeyMarshalFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrKeyMarshalFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrSelfSignedFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrSelfSignedFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	// NewPool sentinels.
	t.Run("ErrPoolCertificateEmpty is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrPoolCertificateEmpty == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrPoolCertificateParseFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrPoolCertificateParseFailed == nil {
			t.Fatal("expected non-nil")
		}
	})

	t.Run("ErrPoolFailed is not nil", func(t *testing.T) {
		t.Parallel()

		if ErrPoolFailed == nil {
			t.Fatal("expected non-nil")
		}
	})
}
