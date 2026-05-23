package certs

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
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

// PEM block type constants.
const (
	pemTypeCertificate = "CERTIFICATE"
	pemTypeCSR         = "CERTIFICATE REQUEST"
)

// SelfSigned generates a self-signed certificate and private key as PEM-encoded bytes.
func SelfSigned(options ...SelfSignedOption) ([]byte, []byte, error) {

	opts := NewSelfSignedOptions(options...)

	privateKey, publicKey, err := generateKey(opts.keyAlgorithm)
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
		NotBefore:          now,
		NotAfter:           now.Add(opts.validity),
		DNSNames:           opts.dnsNames,
		IPAddresses:        opts.ipAddresses,
		SignatureAlgorithm: signatureAlgorithmFor(opts.keyAlgorithm),
	}

	if opts.isCA {
		template.IsCA = true
		template.BasicConstraintsValid = true
		template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	} else {
		template.KeyUsage = x509.KeyUsageDigitalSignature
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return nil, nil, ErrSelfSigned(ErrCertificateCreationFailed, err)
	}

	keyPEM, err := marshalPrivateKeyPEM(privateKey, opts.keyAlgorithm)
	if err != nil {
		return nil, nil, ErrSelfSigned(ErrKeyMarshalFailed, err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: pemTypeCertificate, Bytes: certDER})

	return certPEM, keyPEM, nil
}

// generateKey returns a freshly generated private+public key pair for the given algorithm.
func generateKey(algorithm KeyAlgorithm) (crypto.PrivateKey, crypto.PublicKey, error) {
	switch algorithm {
	case KeyAlgorithmECDSAP256:
		return generateECDSAKey(elliptic.P256())
	case KeyAlgorithmECDSAP384:
		return generateECDSAKey(elliptic.P384())
	case KeyAlgorithmEd25519:
		return generateEd25519Key()
	case KeyAlgorithmRSA2048:
		return generateRSAKey(2048)
	case KeyAlgorithmRSA3072:
		return generateRSAKey(3072)
	default:
		return nil, nil, ErrUnsupportedKeyAlgorithm
	}
}

func generateECDSAKey(curve elliptic.Curve) (crypto.PrivateKey, crypto.PublicKey, error) {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return key, &key.PublicKey, nil
}

func generateEd25519Key() (crypto.PrivateKey, crypto.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, pub, nil
}

func generateRSAKey(bits int) (crypto.PrivateKey, crypto.PublicKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return key, &key.PublicKey, nil
}

// signatureAlgorithmFor returns the matching x509.SignatureAlgorithm for the given key algorithm.
func signatureAlgorithmFor(algorithm KeyAlgorithm) x509.SignatureAlgorithm {
	switch algorithm {
	case KeyAlgorithmECDSAP256:
		return x509.ECDSAWithSHA256
	case KeyAlgorithmECDSAP384:
		return x509.ECDSAWithSHA384
	case KeyAlgorithmEd25519:
		return x509.PureEd25519
	case KeyAlgorithmRSA2048, KeyAlgorithmRSA3072:
		return x509.SHA256WithRSA
	default:
		return x509.UnknownSignatureAlgorithm
	}
}

// marshalPrivateKeyPEM serialises the private key to PEM. ECDSA keys keep the legacy
// "EC PRIVATE KEY" block for backward compatibility; RSA and Ed25519 use PKCS#8.
func marshalPrivateKeyPEM(privateKey crypto.PrivateKey, algorithm KeyAlgorithm) ([]byte, error) {
	switch algorithm {
	case KeyAlgorithmECDSAP256, KeyAlgorithmECDSAP384:
		ecKey, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, ErrUnsupportedKeyAlgorithm
		}
		der, err := x509.MarshalECPrivateKey(ecKey)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), nil
	case KeyAlgorithmEd25519, KeyAlgorithmRSA2048, KeyAlgorithmRSA3072:
		der, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
	default:
		return nil, ErrUnsupportedKeyAlgorithm
	}
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

// LoadCertificate reads a PEM-encoded certificate file from disk and returns the parsed certificate.
func LoadCertificate(path string) (*x509.Certificate, error) {

	if cutils.Empty(path) {
		return nil, ErrLoad(ErrPathEmpty)
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: file path comes from caller by design
	if err != nil {
		return nil, ErrLoad(ErrLoadFileFailed, err)
	}

	cert, err := ParseCertificatePEM(data)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// ParseCertificatePEM parses PEM-encoded certificate bytes and returns the certificate.
func ParseCertificatePEM(pemBytes []byte) (*x509.Certificate, error) {

	if len(pemBytes) == 0 {
		return nil, ErrLoad(ErrPEMEmpty)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrLoad(ErrPEMDecodeFailed)
	}

	if block.Type != pemTypeCertificate {
		return nil, ErrLoad(ErrPEMBlockTypeUnexpected)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, ErrLoad(ErrParseCertificateFailed, err)
	}

	return cert, nil
}

// LoadPrivateKey reads a PEM-encoded private key file and returns the parsed key.
// It auto-detects the encoding (PKCS#8, PKCS#1 RSA, or SEC1 EC).
func LoadPrivateKey(path string) (crypto.PrivateKey, error) {

	if cutils.Empty(path) {
		return nil, ErrLoad(ErrPathEmpty)
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: file path comes from caller by design
	if err != nil {
		return nil, ErrLoad(ErrLoadFileFailed, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrLoad(ErrPEMDecodeFailed)
	}

	pkcs8Key, pkcs8Err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if pkcs8Err == nil {
		return pkcs8Key, nil
	}

	rsaKey, rsaErr := x509.ParsePKCS1PrivateKey(block.Bytes)
	if rsaErr == nil {
		return rsaKey, nil
	}

	ecKey, ecErr := x509.ParseECPrivateKey(block.Bytes)
	if ecErr == nil {
		return ecKey, nil
	}

	return nil, ErrLoad(ErrParsePrivateKeyFailed)
}

// ParsePEMChain parses a PEM bundle containing one or more CERTIFICATE blocks
// and returns the certificates in the order they appear in the input.
func ParsePEMChain(pemBytes []byte) ([]*x509.Certificate, error) {

	if len(pemBytes) == 0 {
		return nil, ErrLoad(ErrPEMEmpty)
	}

	chain := make([]*x509.Certificate, 0)
	rest := pemBytes
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		if block.Type != pemTypeCertificate {
			return nil, ErrLoad(ErrPEMBlockTypeUnexpected)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, ErrLoad(ErrParseCertificateFailed, err)
		}

		chain = append(chain, cert)
	}

	if len(chain) == 0 {
		return nil, ErrLoad(ErrPEMDecodeFailed)
	}

	return chain, nil
}

// GenerateCSR generates a PEM-encoded certificate signing request.
// The caller must supply a private key via WithCSRPrivateKey so that any algorithm
// from the signers/* packages (or anywhere else) can be used.
func GenerateCSR(options ...CSROption) ([]byte, error) {

	opts := NewCSROptions(options...)

	if opts.privateKey == nil {
		return nil, ErrCSR(ErrPrivateKeyNil)
	}

	signer, ok := opts.privateKey.(crypto.Signer)
	if !ok {
		return nil, ErrCSR(ErrUnsupportedKeyAlgorithm)
	}

	template := &x509.CertificateRequest{
		Subject:            opts.subject,
		DNSNames:           opts.dnsNames,
		IPAddresses:        opts.ipAddresses,
		EmailAddresses:     opts.emailAddresses,
		SignatureAlgorithm: opts.signatureAlgorithm,
	}

	if template.SignatureAlgorithm == x509.UnknownSignatureAlgorithm {
		template.SignatureAlgorithm = signatureAlgorithmForKey(signer)
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, template, opts.privateKey)
	if err != nil {
		return nil, ErrCSR(ErrGenerateCSRFailed, err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: pemTypeCSR, Bytes: csrDER}), nil
}

// ParseCSR parses a PEM-encoded certificate signing request and verifies its signature.
func ParseCSR(pemBytes []byte) (*x509.CertificateRequest, error) {

	if len(pemBytes) == 0 {
		return nil, ErrCSR(ErrPEMEmpty)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrCSR(ErrPEMDecodeFailed)
	}

	if block.Type != pemTypeCSR {
		return nil, ErrCSR(ErrPEMBlockTypeUnexpected)
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, ErrCSR(ErrParseCSRFailed, err)
	}

	sigErr := csr.CheckSignature()
	if sigErr != nil {
		return nil, ErrCSR(ErrCSRSignatureVerifyFailed, sigErr)
	}

	return csr, nil
}

// signatureAlgorithmForKey picks a sane default signature algorithm based on the signer's public key.
func signatureAlgorithmForKey(signer crypto.Signer) x509.SignatureAlgorithm {
	switch pub := signer.Public().(type) {
	case *ecdsa.PublicKey:
		switch pub.Curve {
		case elliptic.P256():
			return x509.ECDSAWithSHA256
		case elliptic.P384():
			return x509.ECDSAWithSHA384
		case elliptic.P521():
			return x509.ECDSAWithSHA512
		default:
			return x509.ECDSAWithSHA256
		}
	case ed25519.PublicKey:
		return x509.PureEd25519
	case *rsa.PublicKey:
		return x509.SHA256WithRSA
	default:
		return x509.UnknownSignatureAlgorithm
	}
}
