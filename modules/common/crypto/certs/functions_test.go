package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testCerts struct {
	caCertFile     string
	caCertPEM      []byte
	clientCertFile string
	clientCertPEM  []byte
	clientKeyFile  string
	clientKeyPEM   []byte
}

func generateTestCerts(t *testing.T) testCerts {
	t.Helper()

	dir := t.TempDir()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate CA key: %v", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create CA certificate: %v", err)
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	caCertFile := filepath.Join(dir, "ca.pem")
	err = os.WriteFile(caCertFile, caCertPEM, 0o600)
	if err != nil {
		t.Fatalf("failed to write CA cert: %v", err)
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate client key: %v", err)
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test Client"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		t.Fatalf("failed to parse CA certificate: %v", err)
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create client certificate: %v", err)
	}

	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})

	clientCertFile := filepath.Join(dir, "client.pem")
	err = os.WriteFile(clientCertFile, clientCertPEM, 0o600)
	if err != nil {
		t.Fatalf("failed to write client cert: %v", err)
	}

	clientKeyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		t.Fatalf("failed to marshal client key: %v", err)
	}

	clientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER})

	clientKeyFile := filepath.Join(dir, "client-key.pem")
	err = os.WriteFile(clientKeyFile, clientKeyPEM, 0o600)
	if err != nil {
		t.Fatalf("failed to write client key: %v", err)
	}

	return testCerts{
		caCertFile:     caCertFile,
		caCertPEM:      caCertPEM,
		clientCertFile: clientCertFile,
		clientCertPEM:  clientCertPEM,
		clientKeyFile:  clientKeyFile,
		clientKeyPEM:   clientKeyPEM,
	}
}

func TestClientTls(t *testing.T) {
	t.Parallel()

	t.Run("returns error when server name is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTls("", "ca.pem", "client.pem", "client-key.pem", false)
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrServerNameEmpty) {
			t.Fatalf("expected ErrServerNameEmpty, got %v", err)
		}
	})

	t.Run("returns error when ca certificate path is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTls("localhost", "", "client.pem", "client-key.pem", false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrCaCertificateEmpty) {
			t.Fatalf("expected ErrCaCertificateEmpty, got %v", err)
		}
	})

	t.Run("returns error when client certificate path is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTls("localhost", "ca.pem", "", "client-key.pem", false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientCertificateEmpty) {
			t.Fatalf("expected ErrClientCertificateEmpty, got %v", err)
		}
	})

	t.Run("returns error when client key path is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTls("localhost", "ca.pem", "client.pem", "", false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientKeyEmpty) {
			t.Fatalf("expected ErrClientKeyEmpty, got %v", err)
		}
	})

	t.Run("returns error when ca certificate file not found", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTls("localhost", "/nonexistent/ca.pem", "client.pem", "client-key.pem", false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrCaCertificateReadFailed) {
			t.Fatalf("expected ErrCaCertificateReadFailed, got %v", err)
		}
	})

	t.Run("returns error when ca certificate PEM is invalid", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		invalidPEM := filepath.Join(dir, "invalid-ca.pem")
		err := os.WriteFile(invalidPEM, []byte("not a valid PEM"), 0o600)
		if err != nil {
			t.Fatalf("failed to write invalid PEM: %v", err)
		}

		tc := generateTestCerts(t)

		_, err = ClientTls("localhost", invalidPEM, tc.clientCertFile, tc.clientKeyFile, false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrCaCertificateParseFailed) {
			t.Fatalf("expected ErrCaCertificateParseFailed, got %v", err)
		}
	})

	t.Run("returns error when client certificate keypair is invalid", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		_, err := ClientTls("localhost", tc.caCertFile, tc.caCertFile, tc.clientKeyFile, false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientCertificateLoadFailed) {
			t.Fatalf("expected ErrClientCertificateLoadFailed, got %v", err)
		}
	})

	t.Run("returns valid tls config on success", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ClientTls("test.example.com", tc.caCertFile, tc.clientCertFile, tc.clientKeyFile, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ServerName != "test.example.com" {
			t.Fatalf("expected ServerName 'test.example.com', got %q", cfg.ServerName)
		}

		if cfg.MinVersion != tls.VersionTLS12 {
			t.Fatalf("expected MinVersion TLS 1.2, got %d", cfg.MinVersion)
		}

		if !cfg.InsecureSkipVerify {
			t.Fatal("expected InsecureSkipVerify to be true")
		}

		if cfg.RootCAs == nil {
			t.Fatal("expected non-nil RootCAs")
		}

		if len(cfg.Certificates) != 1 {
			t.Fatalf("expected 1 certificate, got %d", len(cfg.Certificates))
		}
	})

	t.Run("returns valid tls config with insecure skip verify false", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ClientTls("secure.example.com", tc.caCertFile, tc.clientCertFile, tc.clientKeyFile, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.InsecureSkipVerify {
			t.Fatal("expected InsecureSkipVerify to be false")
		}
	})
}

func TestServerTls(t *testing.T) {
	t.Parallel()

	t.Run("returns error when cert file is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTls("", "key.pem", "")
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrServerCertificateEmpty) {
			t.Fatalf("expected ErrServerCertificateEmpty, got %v", err)
		}
	})

	t.Run("returns error when key file is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTls("cert.pem", "", "")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrServerKeyEmpty) {
			t.Fatalf("expected ErrServerKeyEmpty, got %v", err)
		}
	})

	t.Run("returns error when server cert load fails", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTls("/nonexistent/cert.pem", "/nonexistent/key.pem", "")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrServerCertificateLoadFailed) {
			t.Fatalf("expected ErrServerCertificateLoadFailed, got %v", err)
		}
	})

	t.Run("returns error when client ca cert file not found", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		_, err := ServerTls(tc.clientCertFile, tc.clientKeyFile, "/nonexistent/ca.pem")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientCACertificateReadFailed) {
			t.Fatalf("expected ErrClientCACertificateReadFailed, got %v", err)
		}
	})

	t.Run("returns error when client ca cert PEM is invalid", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		dir := t.TempDir()
		invalidPEM := filepath.Join(dir, "invalid-ca.pem")
		err := os.WriteFile(invalidPEM, []byte("not a valid PEM"), 0o600)
		if err != nil {
			t.Fatalf("failed to write invalid PEM: %v", err)
		}

		_, err = ServerTls(tc.clientCertFile, tc.clientKeyFile, invalidPEM)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientCACertificateParseFailed) {
			t.Fatalf("expected ErrClientCACertificateParseFailed, got %v", err)
		}
	})

	t.Run("returns valid config without client ca", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ServerTls(tc.clientCertFile, tc.clientKeyFile, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Certificates) != 1 {
			t.Fatalf("expected 1 certificate, got %d", len(cfg.Certificates))
		}

		if cfg.MinVersion != tls.VersionTLS12 {
			t.Fatalf("expected MinVersion TLS 1.2, got %d", cfg.MinVersion)
		}

		if cfg.ClientCAs != nil {
			t.Fatal("expected nil ClientCAs")
		}

		if cfg.ClientAuth != tls.NoClientCert {
			t.Fatalf("expected NoClientCert, got %v", cfg.ClientAuth)
		}
	})

	t.Run("returns valid config with client ca for mtls", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ServerTls(tc.clientCertFile, tc.clientKeyFile, tc.caCertFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Certificates) != 1 {
			t.Fatalf("expected 1 certificate, got %d", len(cfg.Certificates))
		}

		if cfg.ClientCAs == nil {
			t.Fatal("expected non-nil ClientCAs")
		}

		if cfg.ClientAuth != tls.RequireAndVerifyClientCert {
			t.Fatalf("expected RequireAndVerifyClientCert, got %v", cfg.ClientAuth)
		}
	})
}

func TestClientTlsFromPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns error when server name is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTlsFromPEM("", []byte("ca"), []byte("cert"), []byte("key"), false)
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrServerNameEmpty) {
			t.Fatalf("expected ErrServerNameEmpty, got %v", err)
		}
	})

	t.Run("returns error when ca cert pem is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTlsFromPEM("localhost", nil, []byte("cert"), []byte("key"), false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrCaCertificatePEMEmpty) {
			t.Fatalf("expected ErrCaCertificatePEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when ca cert pem is zero length", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTlsFromPEM("localhost", []byte{}, []byte("cert"), []byte("key"), false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrCaCertificatePEMEmpty) {
			t.Fatalf("expected ErrCaCertificatePEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when client cert pem is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTlsFromPEM("localhost", []byte("ca"), nil, []byte("key"), false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientCertificatePEMEmpty) {
			t.Fatalf("expected ErrClientCertificatePEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when client key pem is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTlsFromPEM("localhost", []byte("ca"), []byte("cert"), nil, false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientKeyPEMEmpty) {
			t.Fatalf("expected ErrClientKeyPEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when ca cert pem parse fails", func(t *testing.T) {
		t.Parallel()

		_, err := ClientTlsFromPEM("localhost", []byte("not valid pem"), []byte("cert"), []byte("key"), false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrCaCertificateParseFailed) {
			t.Fatalf("expected ErrCaCertificateParseFailed, got %v", err)
		}
	})

	t.Run("returns error when client key pair fails", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		_, err := ClientTlsFromPEM("localhost", tc.caCertPEM, tc.caCertPEM, tc.clientKeyPEM, false)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientKeyPairFailed) {
			t.Fatalf("expected ErrClientKeyPairFailed, got %v", err)
		}
	})

	t.Run("returns valid tls config on success", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ClientTlsFromPEM("test.example.com", tc.caCertPEM, tc.clientCertPEM, tc.clientKeyPEM, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ServerName != "test.example.com" {
			t.Fatalf("expected ServerName 'test.example.com', got %q", cfg.ServerName)
		}

		if cfg.MinVersion != tls.VersionTLS12 {
			t.Fatalf("expected MinVersion TLS 1.2, got %d", cfg.MinVersion)
		}

		if !cfg.InsecureSkipVerify {
			t.Fatal("expected InsecureSkipVerify to be true")
		}

		if cfg.RootCAs == nil {
			t.Fatal("expected non-nil RootCAs")
		}

		if len(cfg.Certificates) != 1 {
			t.Fatalf("expected 1 certificate, got %d", len(cfg.Certificates))
		}
	})

	t.Run("returns valid tls config with insecure skip verify false", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ClientTlsFromPEM("secure.example.com", tc.caCertPEM, tc.clientCertPEM, tc.clientKeyPEM, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.InsecureSkipVerify {
			t.Fatal("expected InsecureSkipVerify to be false")
		}
	})
}

func TestServerTlsFromPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns error when cert pem is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTlsFromPEM([]byte{}, []byte("key"), nil)
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrServerCertificatePEMEmpty) {
			t.Fatalf("expected ErrServerCertificatePEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when cert pem is nil", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTlsFromPEM(nil, []byte("key"), nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrServerCertificatePEMEmpty) {
			t.Fatalf("expected ErrServerCertificatePEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when key pem is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTlsFromPEM([]byte("cert"), []byte{}, nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrServerKeyPEMEmpty) {
			t.Fatalf("expected ErrServerKeyPEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when key pem is nil", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTlsFromPEM([]byte("cert"), nil, nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrServerKeyPEMEmpty) {
			t.Fatalf("expected ErrServerKeyPEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when key pair fails", func(t *testing.T) {
		t.Parallel()

		_, err := ServerTlsFromPEM([]byte("not a cert"), []byte("not a key"), nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrServerKeyPairFailed) {
			t.Fatalf("expected ErrServerKeyPairFailed, got %v", err)
		}
	})

	t.Run("returns error when client ca cert pem parse fails", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		_, err := ServerTlsFromPEM(tc.clientCertPEM, tc.clientKeyPEM, []byte("not valid pem"))
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrClientCACertificateParseFailed) {
			t.Fatalf("expected ErrClientCACertificateParseFailed, got %v", err)
		}
	})

	t.Run("returns valid config without client ca", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ServerTlsFromPEM(tc.clientCertPEM, tc.clientKeyPEM, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Certificates) != 1 {
			t.Fatalf("expected 1 certificate, got %d", len(cfg.Certificates))
		}

		if cfg.MinVersion != tls.VersionTLS12 {
			t.Fatalf("expected MinVersion TLS 1.2, got %d", cfg.MinVersion)
		}

		if cfg.ClientCAs != nil {
			t.Fatal("expected nil ClientCAs")
		}

		if cfg.ClientAuth != tls.NoClientCert {
			t.Fatalf("expected NoClientCert, got %v", cfg.ClientAuth)
		}
	})

	t.Run("returns valid config with empty client ca", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ServerTlsFromPEM(tc.clientCertPEM, tc.clientKeyPEM, []byte{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ClientCAs != nil {
			t.Fatal("expected nil ClientCAs")
		}

		if cfg.ClientAuth != tls.NoClientCert {
			t.Fatalf("expected NoClientCert, got %v", cfg.ClientAuth)
		}
	})

	t.Run("returns valid config with client ca for mtls", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		cfg, err := ServerTlsFromPEM(tc.clientCertPEM, tc.clientKeyPEM, tc.caCertPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Certificates) != 1 {
			t.Fatalf("expected 1 certificate, got %d", len(cfg.Certificates))
		}

		if cfg.ClientCAs == nil {
			t.Fatal("expected non-nil ClientCAs")
		}

		if cfg.ClientAuth != tls.RequireAndVerifyClientCert {
			t.Fatalf("expected RequireAndVerifyClientCert, got %v", cfg.ClientAuth)
		}
	})
}

func TestSelfSigned(t *testing.T) {
	t.Parallel()

	t.Run("generates valid cert and key with defaults", func(t *testing.T) {
		t.Parallel()

		certPEM, keyPEM, err := SelfSigned()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(certPEM) == 0 {
			t.Fatal("expected non-empty certPEM")
		}

		if len(keyPEM) == 0 {
			t.Fatal("expected non-empty keyPEM")
		}

		block, _ := pem.Decode(certPEM)
		if block == nil {
			t.Fatal("expected valid PEM block for cert")
		}
		if block.Type != "CERTIFICATE" {
			t.Fatalf("expected CERTIFICATE block type, got %q", block.Type)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("failed to parse certificate: %v", err)
		}

		if cert.Subject.Organization[0] != "Development" {
			t.Fatalf("expected organization 'Development', got %q", cert.Subject.Organization[0])
		}

		if cert.IsCA {
			t.Fatal("expected non-CA certificate")
		}

		if len(cert.DNSNames) != 1 || cert.DNSNames[0] != "localhost" {
			t.Fatalf("expected DNSNames [localhost], got %v", cert.DNSNames)
		}

		foundIPv4 := false
		foundIPv6 := false
		for _, ip := range cert.IPAddresses {
			if ip.Equal(net.IPv4(127, 0, 0, 1)) {
				foundIPv4 = true
			}
			if ip.Equal(net.IPv6loopback) {
				foundIPv6 = true
			}
		}
		if !foundIPv4 {
			t.Fatal("expected 127.0.0.1 in IPAddresses")
		}
		if !foundIPv6 {
			t.Fatal("expected ::1 in IPAddresses")
		}

		keyBlock, _ := pem.Decode(keyPEM)
		if keyBlock == nil {
			t.Fatal("expected valid PEM block for key")
		}
		if keyBlock.Type != "EC PRIVATE KEY" {
			t.Fatalf("expected EC PRIVATE KEY block type, got %q", keyBlock.Type)
		}

		_, err = x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			t.Fatalf("failed to parse EC private key: %v", err)
		}
	})

	t.Run("generates CA certificate when isCA is true", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned(WithIsCA(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		block, _ := pem.Decode(certPEM)
		if block == nil {
			t.Fatal("expected valid PEM block")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("failed to parse certificate: %v", err)
		}

		if !cert.IsCA {
			t.Fatal("expected CA certificate")
		}

		if !cert.BasicConstraintsValid {
			t.Fatal("expected BasicConstraintsValid to be true")
		}

		if cert.KeyUsage&x509.KeyUsageCertSign == 0 {
			t.Fatal("expected KeyUsageCertSign")
		}
	})

	t.Run("applies custom options", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned(
			WithOrganization("TestOrg"),
			WithValidity(24*time.Hour),
			WithDNSNames("example.com", "*.example.com"),
			WithIPAddresses(net.IPv4(10, 0, 0, 1)),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		block, _ := pem.Decode(certPEM)
		if block == nil {
			t.Fatal("expected valid PEM block")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("failed to parse certificate: %v", err)
		}

		if cert.Subject.Organization[0] != "TestOrg" {
			t.Fatalf("expected organization 'TestOrg', got %q", cert.Subject.Organization[0])
		}

		if len(cert.DNSNames) != 2 {
			t.Fatalf("expected 2 DNS names, got %d", len(cert.DNSNames))
		}

		if len(cert.IPAddresses) != 1 || !cert.IPAddresses[0].Equal(net.IPv4(10, 0, 0, 1)) {
			t.Fatalf("expected [10.0.0.1], got %v", cert.IPAddresses)
		}

		expectedNotAfter := time.Now().Add(24 * time.Hour)
		diff := cert.NotAfter.Sub(expectedNotAfter)
		if diff < -time.Minute || diff > time.Minute {
			t.Fatalf("expected NotAfter close to 24h from now, got %v", cert.NotAfter)
		}
	})

	t.Run("generates usable keypair for tls", func(t *testing.T) {
		t.Parallel()

		certPEM, keyPEM, err := SelfSigned()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			t.Fatalf("generated cert/key pair is not usable for TLS: %v", err)
		}
	})
}

func TestNewPool(t *testing.T) {
	t.Parallel()

	t.Run("returns error when cert is empty", func(t *testing.T) {
		t.Parallel()

		_, err := NewPool([]byte{})
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrPoolCertificateEmpty) {
			t.Fatalf("expected ErrPoolCertificateEmpty, got %v", err)
		}
	})

	t.Run("returns error when cert is nil", func(t *testing.T) {
		t.Parallel()

		_, err := NewPool(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPoolCertificateEmpty) {
			t.Fatalf("expected ErrPoolCertificateEmpty, got %v", err)
		}
	})

	t.Run("returns error when cert PEM is invalid", func(t *testing.T) {
		t.Parallel()

		_, err := NewPool([]byte("not valid pem"))
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPoolCertificateParseFailed) {
			t.Fatalf("expected ErrPoolCertificateParseFailed, got %v", err)
		}
	})

	t.Run("returns pool with single cert", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		pool, err := NewPool(tc.caCertPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pool == nil {
			t.Fatal("expected non-nil pool")
		}
	})

	t.Run("returns pool with multiple certs", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		pool, err := NewPool(tc.caCertPEM, tc.clientCertPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pool == nil {
			t.Fatal("expected non-nil pool")
		}
	})

	t.Run("returns empty pool with no args", func(t *testing.T) {
		t.Parallel()

		pool, err := NewPool()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pool == nil {
			t.Fatal("expected non-nil pool")
		}
	})

	t.Run("returns error when second cert is invalid", func(t *testing.T) {
		t.Parallel()

		tc := generateTestCerts(t)

		_, err := NewPool(tc.caCertPEM, []byte("bad pem"))
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPoolCertificateParseFailed) {
			t.Fatalf("expected ErrPoolCertificateParseFailed, got %v", err)
		}
	})
}
