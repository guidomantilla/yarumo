// Package certs provides TLS certificate loading and configuration utilities.
package certs

import (
	"crypto/tls"
	"crypto/x509"
)

var (
	_ ClientTlsFn        = ClientTls
	_ ServerTlsFn        = ServerTls
	_ ClientTlsFromPEMFn = ClientTlsFromPEM
	_ ServerTlsFromPEMFn = ServerTlsFromPEM
	_ SelfSignedFn       = SelfSigned
	_ NewPoolFn          = NewPool
)

// ClientTlsFn is the function type for ClientTls.
type ClientTlsFn func(serverName string, caCertificate string, clientCertificate string, clientKey string, insecureSkipVerify bool) (*tls.Config, error)

// ServerTlsFn is the function type for ServerTls.
type ServerTlsFn func(certFile string, keyFile string, clientCACertFile string) (*tls.Config, error)

// ClientTlsFromPEMFn is the function type for ClientTlsFromPEM.
type ClientTlsFromPEMFn func(serverName string, caCertPEM []byte, clientCertPEM []byte, clientKeyPEM []byte, insecureSkipVerify bool) (*tls.Config, error)

// ServerTlsFromPEMFn is the function type for ServerTlsFromPEM.
type ServerTlsFromPEMFn func(certPEM []byte, keyPEM []byte, clientCACertPEM []byte) (*tls.Config, error)

// SelfSignedFn is the function type for SelfSigned.
type SelfSignedFn func(opts ...SelfSignedOption) ([]byte, []byte, error)

// NewPoolFn is the function type for NewPool.
type NewPoolFn func(pemCerts ...[]byte) (*x509.CertPool, error)
