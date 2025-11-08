package certs

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/modules/common/errs"
)

const (
	CertificateTLSType       = "TLS"
	CertificateCaTLSType     = "ca certificate"
	CertificateClientTLSType = "client certificate"
)

var (
	_ error = (*CertificateError)(nil)
)

type CertificateError struct {
	cerrs.TypedError
}

func (e *CertificateError) Error() string {
	return fmt.Sprintf("certificate %s error: %s", e.Type, e.Err)
}

var (
	ErrCertificateTLSFailed = errors.New("certificate tls generation failed")
)

func ErrCertificateTLS(errs ...error) error {
	return &CertificateError{
		TypedError: cerrs.TypedError{
			Type: CertificateTLSType,
			Err:  errors.Join(append(errs, ErrCertificateTLSFailed)...),
		},
	}
}

func ErrCertificateCaTLS(errs ...error) error {
	return &CertificateError{
		TypedError: cerrs.TypedError{
			Type: CertificateCaTLSType,
			Err:  errors.Join(append(errs, ErrCertificateTLSFailed)...),
		},
	}
}

func ErrCertificateClientTLS(errs ...error) error {
	return &CertificateError{
		TypedError: cerrs.TypedError{
			Type: CertificateClientTLSType,
			Err:  errors.Join(append(errs, ErrCertificateTLSFailed)...),
		},
	}
}
