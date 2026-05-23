package certs

import "testing"

// FuzzClientTlsFromPEM exercises ClientTlsFromPEM with attacker-controlled PEM
// bytes for the CA, client certificate, and client key to ensure the parser
// never panics. The function delegates to crypto/x509 and crypto/tls for the
// heavy lifting, but the wrapping logic in this package has its own surface
// (length checks, error path construction). YA-0007 previously found a
// cassert.True panic in this layer — this fuzz guards against regressions.
//
// The contract: any input, regardless of how malformed, must produce an error
// and never a panic.
func FuzzClientTlsFromPEM(f *testing.F) {
	f.Add([]byte("-----BEGIN CERTIFICATE-----"), []byte{}, []byte{})
	f.Add([]byte{}, []byte{}, []byte{})
	f.Add([]byte("not a PEM at all"), []byte("not a PEM"), []byte("not a PEM"))

	f.Fuzz(func(t *testing.T, ca, cert, key []byte) {
		t.Parallel()
		_, _ = ClientTlsFromPEM("svc", ca, cert, key, false)
	})
}

// FuzzServerTlsFromPEM exercises the server-side counterpart of
// ClientTlsFromPEM. Same contract: no panic on any input.
func FuzzServerTlsFromPEM(f *testing.F) {
	f.Add([]byte("-----BEGIN CERTIFICATE-----"), []byte{}, []byte{})
	f.Add([]byte{}, []byte{}, []byte{})

	f.Fuzz(func(t *testing.T, cert, key, clientCA []byte) {
		t.Parallel()
		_, _ = ServerTlsFromPEM(cert, key, clientCA)
	})
}

// FuzzParseCertificatePEM exercises ParseCertificatePEM which decodes one
// CERTIFICATE block via pem.Decode and parses it via x509.ParseCertificate.
// Standard-library parsers are battle-tested, but the wrapping branch decisions
// (empty input, block-type check, parse-error path) are exercised here.
func FuzzParseCertificatePEM(f *testing.F) {
	f.Add([]byte("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"))
	f.Add([]byte{})
	f.Add([]byte("not a PEM"))

	f.Fuzz(func(t *testing.T, pemBytes []byte) {
		t.Parallel()
		_, _ = ParseCertificatePEM(pemBytes)
	})
}

// FuzzParsePEMChain exercises the multi-block PEM chain parser. Distinct from
// ParseCertificatePEM in that it loops over pem.Decode until exhausted; the
// loop termination, block-type check inside the loop, and final empty-chain
// check are the surface of interest.
func FuzzParsePEMChain(f *testing.F) {
	f.Add([]byte("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"))
	f.Add([]byte{})
	f.Add([]byte("garbage"))

	f.Fuzz(func(t *testing.T, pemBytes []byte) {
		t.Parallel()
		_, _ = ParsePEMChain(pemBytes)
	})
}

// FuzzParseCSR exercises the CSR parser which decodes a CERTIFICATE REQUEST
// PEM block and verifies its signature. Same no-panic contract.
func FuzzParseCSR(f *testing.F) {
	f.Add([]byte("-----BEGIN CERTIFICATE REQUEST-----\n-----END CERTIFICATE REQUEST-----"))
	f.Add([]byte{})
	f.Add([]byte("not a CSR"))

	f.Fuzz(func(t *testing.T, pemBytes []byte) {
		t.Parallel()
		_, _ = ParseCSR(pemBytes)
	})
}
