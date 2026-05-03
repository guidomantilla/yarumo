package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// ClientTls builds a client TLS configuration from the given certificate files.
func ClientTls(serverName string, caCertificate string, clientCertificate string, clientKey string, insecureSkipVerify bool) (*tls.Config, error) {

	if cutils.Empty(serverName) {
		return nil, ErrClientTls(ErrServerNameEmpty)
	}

	if cutils.Empty(caCertificate) {
		return nil, ErrClientTls(ErrCaCertificateEmpty)
	}

	if cutils.Empty(clientCertificate) {
		return nil, ErrClientTls(ErrClientCertificateEmpty)
	}

	if cutils.Empty(clientKey) {
		return nil, ErrClientTls(ErrClientKeyEmpty)
	}

	caCert, err := os.ReadFile(caCertificate) //nolint:gosec // G304: file path comes from caller by design
	if err != nil {
		return nil, ErrClientTls(ErrCaCertificateReadFailed, err)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, ErrClientTls(ErrCaCertificateParseFailed)
	}

	cert, err := tls.LoadX509KeyPair(clientCertificate, clientKey)
	if err != nil {
		return nil, ErrClientTls(ErrClientCertificateLoadFailed, err)
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
	}

	return tlsConfig, nil
}

// ServerTls builds a server TLS configuration from the given certificate files.
func ServerTls(certFile string, keyFile string, clientCACertFile string) (*tls.Config, error) {

	if cutils.Empty(certFile) {
		return nil, ErrServerTls(ErrServerCertificateEmpty)
	}

	if cutils.Empty(keyFile) {
		return nil, ErrServerTls(ErrServerKeyEmpty)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, ErrServerTls(ErrServerCertificateLoadFailed, err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if !cutils.Empty(clientCACertFile) {
		caCert, err := os.ReadFile(clientCACertFile) //nolint:gosec // G304: file path comes from caller by design
		if err != nil {
			return nil, ErrServerTls(ErrClientCACertificateReadFailed, err)
		}

		clientCAs := x509.NewCertPool()
		ok := clientCAs.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, ErrServerTls(ErrClientCACertificateParseFailed)
		}

		tlsConfig.ClientCAs = clientCAs
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

// ClientTlsFromPEM builds a client TLS configuration from PEM-encoded bytes.
func ClientTlsFromPEM(serverName string, caCertPEM []byte, clientCertPEM []byte, clientKeyPEM []byte, insecureSkipVerify bool) (*tls.Config, error) {

	if cutils.Empty(serverName) {
		return nil, ErrClientTls(ErrServerNameEmpty)
	}

	if len(caCertPEM) == 0 {
		return nil, ErrClientTls(ErrCaCertificatePEMEmpty)
	}

	if len(clientCertPEM) == 0 {
		return nil, ErrClientTls(ErrClientCertificatePEMEmpty)
	}

	if len(clientKeyPEM) == 0 {
		return nil, ErrClientTls(ErrClientKeyPEMEmpty)
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCertPEM)
	if !ok {
		return nil, ErrClientTls(ErrCaCertificateParseFailed)
	}

	cert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	if err != nil {
		return nil, ErrClientTls(ErrClientKeyPairFailed, err)
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
		ServerName:         serverName,
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
	}

	return tlsConfig, nil
}

// ServerTlsFromPEM builds a server TLS configuration from PEM-encoded bytes.
func ServerTlsFromPEM(certPEM []byte, keyPEM []byte, clientCACertPEM []byte) (*tls.Config, error) {

	if len(certPEM) == 0 {
		return nil, ErrServerTls(ErrServerCertificatePEMEmpty)
	}

	if len(keyPEM) == 0 {
		return nil, ErrServerTls(ErrServerKeyPEMEmpty)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, ErrServerTls(ErrServerKeyPairFailed, err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if len(clientCACertPEM) > 0 {
		clientCAs := x509.NewCertPool()
		ok := clientCAs.AppendCertsFromPEM(clientCACertPEM)
		if !ok {
			return nil, ErrServerTls(ErrClientCACertificateParseFailed)
		}

		tlsConfig.ClientCAs = clientCAs
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

// SelfSigned generates a self-signed certificate and private key as PEM-encoded bytes.
func SelfSigned(options ...SelfSignedOption) ([]byte, []byte, error) {

	opts := NewSelfSignedOptions(options...)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, ErrSelfSigned(ErrKeyGenerationFailed, err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, ErrSelfSigned(ErrKeyGenerationFailed, err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{opts.organization},
		},
		NotBefore:   now,
		NotAfter:    now.Add(opts.validity),
		DNSNames:    opts.dnsNames,
		IPAddresses: opts.ipAddresses,
	}

	if opts.isCA {
		template.IsCA = true
		template.BasicConstraintsValid = true
		template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	} else {
		template.KeyUsage = x509.KeyUsageDigitalSignature
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, ErrSelfSigned(ErrCertificateCreationFailed, err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, ErrSelfSigned(ErrKeyMarshalFailed, err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

// NewPool builds a certificate pool from multiple PEM-encoded certificate bytes.
func NewPool(pemCerts ...[]byte) (*x509.CertPool, error) {

	pool := x509.NewCertPool()

	for _, pemData := range pemCerts {
		if len(pemData) == 0 {
			return nil, ErrPool(ErrPoolCertificateEmpty)
		}

		ok := pool.AppendCertsFromPEM(pemData)
		if !ok {
			return nil, ErrPool(ErrPoolCertificateParseFailed)
		}
	}

	return pool, nil
}
