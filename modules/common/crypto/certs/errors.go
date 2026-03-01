package certs

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	// CertificateType is the error type for certificate operations.
	CertificateType = "certificate"
)

var (
	_ error = (*Error)(nil)
)

// Error is the domain error for certificate operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error message.
func (e *Error) Error() string {
	return fmt.Sprintf("certificate %s error: %s", e.Type, e.Err)
}

// Sentinel errors for certificate operations.
var (
	ErrServerNameEmpty                = errors.New("server name is empty")
	ErrCaCertificateEmpty             = errors.New("ca certificate path is empty")
	ErrClientCertificateEmpty         = errors.New("client certificate path is empty")
	ErrClientKeyEmpty                 = errors.New("client key path is empty")
	ErrCaCertificateReadFailed        = errors.New("ca certificate read failed")
	ErrCaCertificateParseFailed       = errors.New("ca certificate parse failed")
	ErrClientCertificateLoadFailed    = errors.New("client certificate load failed")
	ErrTlsConfigFailed                = errors.New("tls config failed")
	ErrServerCertificateEmpty         = errors.New("server certificate path is empty")
	ErrServerKeyEmpty                 = errors.New("server key path is empty")
	ErrServerCertificateLoadFailed    = errors.New("server certificate load failed")
	ErrClientCACertificateReadFailed  = errors.New("client ca certificate read failed")
	ErrClientCACertificateParseFailed = errors.New("client ca certificate parse failed")
	ErrServerTlsConfigFailed          = errors.New("server tls config failed")
	ErrCaCertificatePEMEmpty          = errors.New("ca certificate pem is empty")
	ErrClientCertificatePEMEmpty      = errors.New("client certificate pem is empty")
	ErrClientKeyPEMEmpty              = errors.New("client key pem is empty")
	ErrClientKeyPairFailed            = errors.New("client key pair failed")
	ErrServerCertificatePEMEmpty      = errors.New("server certificate pem is empty")
	ErrServerKeyPEMEmpty              = errors.New("server key pem is empty")
	ErrServerKeyPairFailed            = errors.New("server key pair failed")
	ErrKeyGenerationFailed            = errors.New("key generation failed")
	ErrCertificateCreationFailed      = errors.New("certificate creation failed")
	ErrKeyMarshalFailed               = errors.New("key marshal failed")
	ErrSelfSignedFailed               = errors.New("self signed failed")
	ErrPoolCertificateEmpty           = errors.New("pool certificate is empty")
	ErrPoolCertificateParseFailed     = errors.New("pool certificate parse failed")
	ErrPoolFailed                     = errors.New("pool failed")
)

// ErrClientTls creates a certificate error for client TLS operations.
func ErrClientTls(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CertificateType,
			Err:  errors.Join(append(errs, ErrTlsConfigFailed)...),
		},
	}
}

// ErrServerTls creates a certificate error for server TLS operations.
func ErrServerTls(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CertificateType,
			Err:  errors.Join(append(errs, ErrServerTlsConfigFailed)...),
		},
	}
}

// ErrSelfSigned creates a certificate error for self-signed certificate operations.
func ErrSelfSigned(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CertificateType,
			Err:  errors.Join(append(errs, ErrSelfSignedFailed)...),
		},
	}
}

// ErrPool creates a certificate error for certificate pool operations.
func ErrPool(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CertificateType,
			Err:  errors.Join(append(errs, ErrPoolFailed)...),
		},
	}
}
