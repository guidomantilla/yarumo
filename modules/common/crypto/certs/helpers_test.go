package certs

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// writeTempFile writes data to a fresh file under t.TempDir and returns the path.
func writeTempFile(t *testing.T, name string, data []byte) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)

	err := os.WriteFile(path, data, 0o600)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return path
}

func TestLoadCertificate(t *testing.T) {
	t.Parallel()

	t.Run("returns error when path is empty", func(t *testing.T) {
		t.Parallel()

		_, err := LoadCertificate("")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPathEmpty) {
			t.Fatalf("expected ErrPathEmpty, got %v", err)
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("returns error when file is missing", func(t *testing.T) {
		t.Parallel()

		_, err := LoadCertificate("/nonexistent/cert.pem")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrLoadFileFailed) {
			t.Fatalf("expected ErrLoadFileFailed, got %v", err)
		}
	})

	t.Run("returns error when PEM is malformed", func(t *testing.T) {
		t.Parallel()

		path := writeTempFile(t, "bad.pem", []byte("not pem"))

		_, err := LoadCertificate(path)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns error when PEM block type is wrong", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned()
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		path := writeTempFile(t, "key.pem", keyPEM)

		_, err = LoadCertificate(path)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMBlockTypeUnexpected) {
			t.Fatalf("expected ErrPEMBlockTypeUnexpected, got %v", err)
		}
	})

	t.Run("loads a valid certificate", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned(WithOrganization("Loaders"))
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		path := writeTempFile(t, "cert.pem", certPEM)

		cert, err := LoadCertificate(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cert.Subject.Organization[0] != "Loaders" {
			t.Fatalf("expected organization 'Loaders', got %q", cert.Subject.Organization[0])
		}
	})
}

func TestParseCertificatePEM(t *testing.T) {
	t.Parallel()

	t.Run("returns error when input is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCertificatePEM(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMEmpty) {
			t.Fatalf("expected ErrPEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when PEM is malformed", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCertificatePEM([]byte("not pem"))
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns error when block type is wrong", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned()
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		_, err = ParseCertificatePEM(keyPEM)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMBlockTypeUnexpected) {
			t.Fatalf("expected ErrPEMBlockTypeUnexpected, got %v", err)
		}
	})

	t.Run("returns error when DER is corrupted", func(t *testing.T) {
		t.Parallel()

		corrupted := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{0x01, 0x02, 0x03}})

		_, err := ParseCertificatePEM(corrupted)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrParseCertificateFailed) {
			t.Fatalf("expected ErrParseCertificateFailed, got %v", err)
		}
	})

	t.Run("parses a valid certificate", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned()
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		cert, err := ParseCertificatePEM(certPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cert == nil {
			t.Fatal("expected non-nil certificate")
		}
	})
}

func TestLoadPrivateKey(t *testing.T) {
	t.Parallel()

	t.Run("returns error when path is empty", func(t *testing.T) {
		t.Parallel()

		_, err := LoadPrivateKey("")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPathEmpty) {
			t.Fatalf("expected ErrPathEmpty, got %v", err)
		}
	})

	t.Run("returns error when file is missing", func(t *testing.T) {
		t.Parallel()

		_, err := LoadPrivateKey("/nonexistent/key.pem")
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrLoadFileFailed) {
			t.Fatalf("expected ErrLoadFileFailed, got %v", err)
		}
	})

	t.Run("returns error when PEM is malformed", func(t *testing.T) {
		t.Parallel()

		path := writeTempFile(t, "bad.pem", []byte("not pem"))

		_, err := LoadPrivateKey(path)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns error when key bytes are garbage", func(t *testing.T) {
		t.Parallel()

		garbage := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{0x01, 0x02, 0x03}})

		path := writeTempFile(t, "garbage.pem", garbage)

		_, err := LoadPrivateKey(path)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrParsePrivateKeyFailed) {
			t.Fatalf("expected ErrParsePrivateKeyFailed, got %v", err)
		}
	})

	t.Run("loads SEC1 EC private key", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmECDSAP256))
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		path := writeTempFile(t, "key.pem", keyPEM)

		key, err := LoadPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := key.(*ecdsa.PrivateKey); !ok {
			t.Fatalf("expected *ecdsa.PrivateKey, got %T", key)
		}
	})

	t.Run("loads PKCS#8 Ed25519 key", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmEd25519))
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		path := writeTempFile(t, "key.pem", keyPEM)

		key, err := LoadPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := key.(ed25519.PrivateKey); !ok {
			t.Fatalf("expected ed25519.PrivateKey, got %T", key)
		}
	})

	t.Run("loads PKCS#8 RSA key", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmRSA2048))
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		path := writeTempFile(t, "key.pem", keyPEM)

		key, err := LoadPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := key.(*rsa.PrivateKey); !ok {
			t.Fatalf("expected *rsa.PrivateKey, got %T", key)
		}
	})

	t.Run("loads PKCS#1 RSA key", func(t *testing.T) {
		t.Parallel()

		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("rsa.GenerateKey failed: %v", err)
		}

		der := x509.MarshalPKCS1PrivateKey(rsaKey)
		pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})

		path := writeTempFile(t, "key.pem", pemBytes)

		key, err := LoadPrivateKey(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := key.(*rsa.PrivateKey); !ok {
			t.Fatalf("expected *rsa.PrivateKey, got %T", key)
		}
	})
}

func TestParsePEMChain(t *testing.T) {
	t.Parallel()

	t.Run("returns error when input is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ParsePEMChain(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMEmpty) {
			t.Fatalf("expected ErrPEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when no PEM blocks are present", func(t *testing.T) {
		t.Parallel()

		_, err := ParsePEMChain([]byte("not pem"))
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns error when block type is wrong", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned()
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		_, err = ParsePEMChain(keyPEM)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMBlockTypeUnexpected) {
			t.Fatalf("expected ErrPEMBlockTypeUnexpected, got %v", err)
		}
	})

	t.Run("parses a single-cert bundle", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned()
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		chain, err := ParsePEMChain(certPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(chain) != 1 {
			t.Fatalf("expected 1 cert, got %d", len(chain))
		}
	})

	t.Run("parses a multi-cert bundle in order", func(t *testing.T) {
		t.Parallel()

		first, _, err := SelfSigned(WithOrganization("First"))
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		second, _, err := SelfSigned(WithOrganization("Second"))
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		bundle := append([]byte{}, first...)
		bundle = append(bundle, second...)

		chain, err := ParsePEMChain(bundle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(chain) != 2 {
			t.Fatalf("expected 2 certs, got %d", len(chain))
		}

		if chain[0].Subject.Organization[0] != "First" {
			t.Fatalf("expected first organization 'First', got %q", chain[0].Subject.Organization[0])
		}

		if chain[1].Subject.Organization[0] != "Second" {
			t.Fatalf("expected second organization 'Second', got %q", chain[1].Subject.Organization[0])
		}
	})
}

func TestSelfSignedWithKeyAlgorithm(t *testing.T) {
	t.Parallel()

	t.Run("ECDSA P-256 produces EC PRIVATE KEY block (backward compat)", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmECDSAP256))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		block, _ := pem.Decode(keyPEM)
		if block == nil {
			t.Fatal("expected pem block")
		}
		if block.Type != "EC PRIVATE KEY" {
			t.Fatalf("expected EC PRIVATE KEY, got %q", block.Type)
		}
	})

	t.Run("ECDSA P-384 produces matching certificate", func(t *testing.T) {
		t.Parallel()

		certPEM, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmECDSAP384))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cert, err := ParseCertificatePEM(certPEM)
		if err != nil {
			t.Fatalf("ParseCertificatePEM failed: %v", err)
		}

		if cert.SignatureAlgorithm != x509.ECDSAWithSHA384 {
			t.Fatalf("expected ECDSAWithSHA384, got %v", cert.SignatureAlgorithm)
		}

		// Round-trip TLS keypair to assert validity.
		_, pairErr := tls.X509KeyPair(certPEM, keyPEM)
		if pairErr != nil {
			t.Fatalf("tls.X509KeyPair failed: %v", pairErr)
		}
	})

	t.Run("Ed25519 produces PKCS#8 PRIVATE KEY block", func(t *testing.T) {
		t.Parallel()

		certPEM, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmEd25519))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		block, _ := pem.Decode(keyPEM)
		if block == nil {
			t.Fatal("expected pem block")
		}
		if block.Type != "PRIVATE KEY" {
			t.Fatalf("expected PRIVATE KEY, got %q", block.Type)
		}

		cert, err := ParseCertificatePEM(certPEM)
		if err != nil {
			t.Fatalf("ParseCertificatePEM failed: %v", err)
		}

		if cert.SignatureAlgorithm != x509.PureEd25519 {
			t.Fatalf("expected PureEd25519, got %v", cert.SignatureAlgorithm)
		}
	})

	t.Run("RSA 2048 produces SHA256-RSA certificate", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmRSA2048))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cert, err := ParseCertificatePEM(certPEM)
		if err != nil {
			t.Fatalf("ParseCertificatePEM failed: %v", err)
		}

		if cert.SignatureAlgorithm != x509.SHA256WithRSA {
			t.Fatalf("expected SHA256WithRSA, got %v", cert.SignatureAlgorithm)
		}
	})

	t.Run("RSA 3072 produces a valid 3072-bit key", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmRSA3072))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		path := writeTempFile(t, "key.pem", keyPEM)

		key, err := LoadPrivateKey(path)
		if err != nil {
			t.Fatalf("LoadPrivateKey failed: %v", err)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			t.Fatalf("expected *rsa.PrivateKey, got %T", key)
		}

		if bits := rsaKey.N.BitLen(); bits != 3072 {
			t.Fatalf("expected 3072-bit key, got %d", bits)
		}
	})

	t.Run("unknown algorithm falls back to default ECDSA P-256", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithm("garbage")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cert, err := ParseCertificatePEM(certPEM)
		if err != nil {
			t.Fatalf("ParseCertificatePEM failed: %v", err)
		}

		if cert.SignatureAlgorithm != x509.ECDSAWithSHA256 {
			t.Fatalf("expected ECDSAWithSHA256, got %v", cert.SignatureAlgorithm)
		}
	})
}

func TestSelfSignedTLSHandshakeRoundtrip(t *testing.T) {
	t.Parallel()

	algos := []KeyAlgorithm{
		KeyAlgorithmEd25519,
		KeyAlgorithmRSA2048,
		KeyAlgorithmECDSAP384,
	}

	for _, algo := range algos {
		t.Run(string(algo), func(t *testing.T) {
			t.Parallel()

			certPEM, keyPEM, err := SelfSigned(
				WithKeyAlgorithm(algo),
				WithOrganization("Handshake "+string(algo)),
				WithDNSNames("example.test"),
			)
			if err != nil {
				t.Fatalf("SelfSigned failed: %v", err)
			}

			certPath := writeTempFile(t, "cert.pem", certPEM)
			keyPath := writeTempFile(t, "key.pem", keyPEM)

			loadedCert, err := LoadCertificate(certPath)
			if err != nil {
				t.Fatalf("LoadCertificate failed: %v", err)
			}

			loadedKey, err := LoadPrivateKey(keyPath)
			if err != nil {
				t.Fatalf("LoadPrivateKey failed: %v", err)
			}

			pool := x509.NewCertPool()
			pool.AddCert(loadedCert)

			tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
			if err != nil {
				t.Fatalf("tls.X509KeyPair failed: %v", err)
			}

			// Sanity check the loaded key is a crypto.Signer.
			if _, ok := loadedKey.(crypto.Signer); !ok {
				t.Fatalf("loaded key is not a crypto.Signer: %T", loadedKey)
			}

			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("ok"))
			}))
			server.TLS = &tls.Config{
				Certificates: []tls.Certificate{tlsCert},
				MinVersion:   tls.VersionTLS12,
			}
			server.StartTLS()
			defer server.Close()

			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs:    pool,
						ServerName: "example.test",
						MinVersion: tls.VersionTLS12,
					},
				},
			}

			req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
			if reqErr != nil {
				t.Fatalf("http.NewRequestWithContext failed: %v", reqErr)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("client.Do failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body failed: %v", err)
			}

			if string(body) != "ok" {
				t.Fatalf("expected body 'ok', got %q", string(body))
			}
		})
	}
}

func TestGenerateCSR(t *testing.T) {
	t.Parallel()

	t.Run("returns error when private key is missing", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateCSR()
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPrivateKeyNil) {
			t.Fatalf("expected ErrPrivateKeyNil, got %v", err)
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("returns error when key is not a signer", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateCSR(WithCSRPrivateKey("not a key"))
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrUnsupportedKeyAlgorithm) {
			t.Fatalf("expected ErrUnsupportedKeyAlgorithm, got %v", err)
		}
	})

	t.Run("generates a CSR with ECDSA key", func(t *testing.T) {
		t.Parallel()

		_, keyPEM, err := SelfSigned(WithKeyAlgorithm(KeyAlgorithmECDSAP256))
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		path := writeTempFile(t, "key.pem", keyPEM)
		key, err := LoadPrivateKey(path)
		if err != nil {
			t.Fatalf("LoadPrivateKey failed: %v", err)
		}

		csrPEM, err := GenerateCSR(
			WithCSRPrivateKey(key),
			WithCSRSubject(pkix.Name{
				CommonName:   "csr.example.com",
				Organization: []string{"CSR Tests"},
			}),
			WithCSRDNSNames("csr.example.com", "alt.example.com"),
			WithCSRIPAddresses(net.IPv4(10, 0, 0, 5)),
			WithCSREmailAddresses("ops@example.com"),
		)
		if err != nil {
			t.Fatalf("GenerateCSR failed: %v", err)
		}

		block, _ := pem.Decode(csrPEM)
		if block == nil || block.Type != "CERTIFICATE REQUEST" {
			t.Fatalf("expected CERTIFICATE REQUEST block, got %v", block)
		}
	})
}

func TestCSRRoundtrip(t *testing.T) {
	t.Parallel()

	algos := []KeyAlgorithm{
		KeyAlgorithmECDSAP256,
		KeyAlgorithmEd25519,
		KeyAlgorithmRSA2048,
	}

	for _, algo := range algos {
		t.Run(string(algo), func(t *testing.T) {
			t.Parallel()

			_, keyPEM, err := SelfSigned(WithKeyAlgorithm(algo))
			if err != nil {
				t.Fatalf("SelfSigned failed: %v", err)
			}

			keyPath := writeTempFile(t, "key.pem", keyPEM)
			key, err := LoadPrivateKey(keyPath)
			if err != nil {
				t.Fatalf("LoadPrivateKey failed: %v", err)
			}

			csrPEM, err := GenerateCSR(
				WithCSRPrivateKey(key),
				WithCSRSubject(pkix.Name{
					CommonName:   "roundtrip.example.com",
					Organization: []string{"Roundtrip"},
				}),
				WithCSRDNSNames("roundtrip.example.com"),
				WithCSRIPAddresses(net.IPv4(127, 0, 0, 1)),
			)
			if err != nil {
				t.Fatalf("GenerateCSR failed: %v", err)
			}

			csr, err := ParseCSR(csrPEM)
			if err != nil {
				t.Fatalf("ParseCSR failed: %v", err)
			}

			if csr.Subject.CommonName != "roundtrip.example.com" {
				t.Fatalf("expected CN roundtrip.example.com, got %q", csr.Subject.CommonName)
			}

			if len(csr.DNSNames) != 1 || csr.DNSNames[0] != "roundtrip.example.com" {
				t.Fatalf("expected DNS [roundtrip.example.com], got %v", csr.DNSNames)
			}

			if len(csr.IPAddresses) != 1 || !csr.IPAddresses[0].Equal(net.IPv4(127, 0, 0, 1)) {
				t.Fatalf("expected IP 127.0.0.1, got %v", csr.IPAddresses)
			}

			sigErr := csr.CheckSignature()
			if sigErr != nil {
				t.Fatalf("CheckSignature failed: %v", sigErr)
			}
		})
	}
}

func TestParseCSR(t *testing.T) {
	t.Parallel()

	t.Run("returns error when input is empty", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCSR(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMEmpty) {
			t.Fatalf("expected ErrPEMEmpty, got %v", err)
		}
	})

	t.Run("returns error when PEM is malformed", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCSR([]byte("not pem"))
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns error when block type is wrong", func(t *testing.T) {
		t.Parallel()

		certPEM, _, err := SelfSigned()
		if err != nil {
			t.Fatalf("SelfSigned failed: %v", err)
		}

		_, err = ParseCSR(certPEM)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrPEMBlockTypeUnexpected) {
			t.Fatalf("expected ErrPEMBlockTypeUnexpected, got %v", err)
		}
	})

	t.Run("returns error when CSR DER is corrupted", func(t *testing.T) {
		t.Parallel()

		bad := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: []byte{0x01, 0x02}})

		_, err := ParseCSR(bad)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrParseCSRFailed) {
			t.Fatalf("expected ErrParseCSRFailed, got %v", err)
		}
	})
}
