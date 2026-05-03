package certs

import (
	"net"
	"time"
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
}

// NewSelfSignedOptions creates SelfSignedOptions with defaults.
func NewSelfSignedOptions(opts ...SelfSignedOption) *SelfSignedOptions {
	options := &SelfSignedOptions{
		organization: "Development",
		validity:     8760 * time.Hour,
		dnsNames:     []string{"localhost"},
		ipAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		isCA:         false,
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
