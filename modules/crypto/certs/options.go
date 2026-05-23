package certs

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"time"
)

// KeyAlgorithm enumerates the supported private-key algorithms for self-signed certificates.
type KeyAlgorithm string

// Supported key algorithms.
const (
	KeyAlgorithmECDSAP256 KeyAlgorithm = "ecdsa-p256"
	KeyAlgorithmECDSAP384 KeyAlgorithm = "ecdsa-p384"
	KeyAlgorithmEd25519   KeyAlgorithm = "ed25519"
	KeyAlgorithmRSA2048   KeyAlgorithm = "rsa-2048"
	KeyAlgorithmRSA3072   KeyAlgorithm = "rsa-3072"
)

// SelfSignedOption is a functional option for configuring SelfSigned Options.
type SelfSignedOption func(opts *SelfSignedOptions)

// SelfSignedOptions holds the configuration for self-signed certificate generation.
type SelfSignedOptions struct {
	organization string
	validity     time.Duration
	dnsNames     []string
	ipAddresses  []net.IP
	isCA         bool
	keyAlgorithm KeyAlgorithm
}

// NewSelfSignedOptions creates SelfSignedOptions with defaults.
func NewSelfSignedOptions(opts ...SelfSignedOption) *SelfSignedOptions {
	options := &SelfSignedOptions{
		organization: "Development",
		validity:     8760 * time.Hour,
		dnsNames:     []string{"localhost"},
		ipAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		isCA:         false,
		keyAlgorithm: KeyAlgorithmECDSAP256,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithOrganization sets the organization name for the certificate subject.
func WithOrganization(organization string) SelfSignedOption {
	return func(opts *SelfSignedOptions) {
		if organization != "" {
			opts.organization = organization
		}
	}
}

// WithValidity sets the validity duration for the certificate.
func WithValidity(validity time.Duration) SelfSignedOption {
	return func(opts *SelfSignedOptions) {
		if validity >= time.Hour {
			opts.validity = validity
		}
	}
}

// WithDNSNames sets the DNS names for the certificate.
func WithDNSNames(dnsNames ...string) SelfSignedOption {
	return func(opts *SelfSignedOptions) {
		if len(dnsNames) > 0 {
			opts.dnsNames = dnsNames
		}
	}
}

// WithIPAddresses sets the IP addresses for the certificate.
func WithIPAddresses(ipAddresses ...net.IP) SelfSignedOption {
	return func(opts *SelfSignedOptions) {
		if len(ipAddresses) > 0 {
			opts.ipAddresses = ipAddresses
		}
	}
}

// WithIsCA sets whether the certificate should be a CA certificate.
func WithIsCA(isCA bool) SelfSignedOption {
	return func(opts *SelfSignedOptions) {
		opts.isCA = isCA
	}
}

// WithKeyAlgorithm sets the private key algorithm for the self-signed certificate.
// Unknown algorithms are ignored and the default (ECDSA P-256) is kept.
func WithKeyAlgorithm(algorithm KeyAlgorithm) SelfSignedOption {
	return func(opts *SelfSignedOptions) {
		switch algorithm {
		case KeyAlgorithmECDSAP256,
			KeyAlgorithmECDSAP384,
			KeyAlgorithmEd25519,
			KeyAlgorithmRSA2048,
			KeyAlgorithmRSA3072:
			opts.keyAlgorithm = algorithm
		default:
			// Ignore unknown algorithm; default remains in effect.
		}
	}
}

// CSROption is a functional option for configuring CSR Options.
type CSROption func(opts *CSROptions)

// CSROptions holds the configuration for CSR generation.
type CSROptions struct {
	subject            pkix.Name
	dnsNames           []string
	ipAddresses        []net.IP
	emailAddresses     []string
	privateKey         crypto.PrivateKey
	signatureAlgorithm x509.SignatureAlgorithm
}

// NewCSROptions creates CSROptions with defaults.
func NewCSROptions(opts ...CSROption) *CSROptions {
	options := &CSROptions{
		subject:            pkix.Name{},
		dnsNames:           nil,
		ipAddresses:        nil,
		emailAddresses:     nil,
		privateKey:         nil,
		signatureAlgorithm: x509.UnknownSignatureAlgorithm,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithCSRSubject sets the subject for the CSR.
func WithCSRSubject(subject pkix.Name) CSROption {
	return func(opts *CSROptions) {
		opts.subject = subject
	}
}

// WithCSRDNSNames sets the DNS SANs for the CSR.
func WithCSRDNSNames(dnsNames ...string) CSROption {
	return func(opts *CSROptions) {
		if len(dnsNames) > 0 {
			opts.dnsNames = dnsNames
		}
	}
}

// WithCSRIPAddresses sets the IP SANs for the CSR.
func WithCSRIPAddresses(ipAddresses ...net.IP) CSROption {
	return func(opts *CSROptions) {
		if len(ipAddresses) > 0 {
			opts.ipAddresses = ipAddresses
		}
	}
}

// WithCSREmailAddresses sets the email SANs for the CSR.
func WithCSREmailAddresses(emailAddresses ...string) CSROption {
	return func(opts *CSROptions) {
		if len(emailAddresses) > 0 {
			opts.emailAddresses = emailAddresses
		}
	}
}

// WithCSRPrivateKey sets the private key used to sign the CSR.
func WithCSRPrivateKey(privateKey crypto.PrivateKey) CSROption {
	return func(opts *CSROptions) {
		if privateKey != nil {
			opts.privateKey = privateKey
		}
	}
}

// WithCSRSignatureAlgorithm sets the explicit signature algorithm hint for the CSR.
func WithCSRSignatureAlgorithm(algorithm x509.SignatureAlgorithm) CSROption {
	return func(opts *CSROptions) {
		opts.signatureAlgorithm = algorithm
	}
}
