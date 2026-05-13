package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	ccerts "github.com/guidomantilla/yarumo/common/crypto/certs"
)

// TestRoundtrip_Certs exercises the cross-process flow for the certs package:
// self-sign a server certificate, persist the cert and key PEMs to disk, hand
// the on-disk paths to an `httptest.NewUnstartedServer` configured with the
// reloaded TLS material, then drive a real HTTPS request through it using a
// client TLS config also built from the same on-disk files. The successful
// round-trip proves both the server-side load (cert + key) and the client-side
// trust path (CA bundle) work end-to-end without any in-memory shortcut.
//
// Encoding choices:
//   - Certificate: PEM (CERTIFICATE block). PEM is the canonical TLS interchange
//     format and is what `ccerts.SelfSigned` emits; rewriting it to disk is the
//     realistic deployment path (mounted secrets, Kubernetes Secret keys,
//     /etc/ssl/...).
//   - Private key: PEM (PKCS#8 PRIVATE KEY block). Same rationale as the cert.
//     The `ccerts.LoadPrivateKey` helper expects PEM, so writing PEM is the
//     only encoding that closes the loop with the production load API.
func TestRoundtrip_Certs(t *testing.T) {
	t.Parallel()

	t.Run("HTTPS_server_from_disk", func(t *testing.T) {
		t.Parallel()
		runCertHttpsRoundtrip(t)
	})

	t.Run("MutualTLS_from_disk", func(t *testing.T) {
		t.Parallel()
		runCertMutualTLSRoundtrip(t)
	})
}

// runCertHttpsRoundtrip self-signs a localhost certificate, writes the cert
// and key PEM to disk, loads them back via `ccerts.ServerTls` (file-based) and
// `ccerts.ClientTls` (file-based), boots an `httptest` TLS server with the
// reloaded credentials, and verifies that a client trusting the on-disk CA
// can complete an HTTPS round-trip.
func runCertHttpsRoundtrip(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.pem")
	keyPath := filepath.Join(dir, "server-key.pem")

	// Self-sign a server cert. The self-signed cert is also its own CA in this
	// test — the client trusts it explicitly via the CA bundle path.
	certPEM, keyPEM, err := ccerts.SelfSigned(
		ccerts.WithOrganization("Yarumo Roundtrip"),
		ccerts.WithDNSNames("localhost", "127.0.0.1"),
		ccerts.WithValidity(1*time.Hour),
	)
	if err != nil {
		t.Fatalf("SelfSigned: %v", err)
	}

	writeErr := os.WriteFile(certPath, certPEM, 0o600)
	if writeErr != nil {
		t.Fatalf("WriteFile cert: %v", writeErr)
	}

	writeErr = os.WriteFile(keyPath, keyPEM, 0o600)
	if writeErr != nil {
		t.Fatalf("WriteFile key: %v", writeErr)
	}

	// Server side: reload cert + key from disk via ServerTls (no mTLS).
	serverTLS, err := ccerts.ServerTls(certPath, keyPath, "")
	if err != nil {
		t.Fatalf("ServerTls: %v", err)
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "yarumo-roundtrip-ok")
	}))
	server.TLS = serverTLS
	server.StartTLS()
	t.Cleanup(server.Close)

	// Client side: reload the same cert as a CA bundle from disk; treat the
	// (cert, key) pair as the optional client cert. InsecureSkipVerify=false
	// makes this a real trust check — the client verifies the server using only
	// what it loaded from disk.
	clientTLS, err := ccerts.ClientTls("localhost", certPath, certPath, keyPath, false)
	if err != nil {
		t.Fatalf("ClientTls: %v", err)
	}

	resp := doGet(t, server.URL, clientTLS)
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("ReadAll: %v", readErr)
	}

	if string(body) != "yarumo-roundtrip-ok" {
		t.Fatalf("unexpected body: got %q", string(body))
	}

	if resp.TLS == nil {
		t.Fatal("response missing TLS state")
	}

	if len(resp.TLS.PeerCertificates) == 0 {
		t.Fatal("client did not see any peer certificates")
	}
}

// runCertMutualTLSRoundtrip writes a CA, a server cert+key, and a client
// cert+key to disk; reloads them via the file-based ServerTls/ClientTls
// helpers; then asserts that a mutual-TLS HTTPS request completes
// successfully through an `httptest` server.
func runCertMutualTLSRoundtrip(t *testing.T) {
	t.Helper()

	dir := t.TempDir()

	// Self-signed CA acts as the trust anchor for both directions. For this
	// example, the CA's own cert is reused as both the server cert and the
	// client cert (they are pinned by exact-match against the CA bundle path).
	caCertPEM, caKeyPEM, err := ccerts.SelfSigned(
		ccerts.WithOrganization("Yarumo Roundtrip CA"),
		ccerts.WithDNSNames("localhost", "127.0.0.1"),
		ccerts.WithIsCA(true),
		ccerts.WithValidity(1*time.Hour),
	)
	if err != nil {
		t.Fatalf("SelfSigned CA: %v", err)
	}

	caPath := filepath.Join(dir, "ca.pem")
	caKeyPath := filepath.Join(dir, "ca-key.pem")

	writeErr := os.WriteFile(caPath, caCertPEM, 0o600)
	if writeErr != nil {
		t.Fatalf("WriteFile ca: %v", writeErr)
	}

	writeErr = os.WriteFile(caKeyPath, caKeyPEM, 0o600)
	if writeErr != nil {
		t.Fatalf("WriteFile ca-key: %v", writeErr)
	}

	// Server side: load (cert, key) from disk; require client certs signed by
	// the CA bundle path. This is the full mTLS configuration.
	serverTLS, err := ccerts.ServerTls(caPath, caKeyPath, caPath)
	if err != nil {
		t.Fatalf("ServerTls (mTLS): %v", err)
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "missing client cert", http.StatusUnauthorized)
			return
		}

		_, _ = io.WriteString(w, "yarumo-mtls-ok")
	}))
	server.TLS = serverTLS
	server.StartTLS()
	t.Cleanup(server.Close)

	// Client side: full mTLS — trust the CA, and present the same (cert, key)
	// as the client certificate. Everything is reloaded from disk.
	clientTLS, err := ccerts.ClientTls("localhost", caPath, caPath, caKeyPath, false)
	if err != nil {
		t.Fatalf("ClientTls (mTLS): %v", err)
	}

	resp := doGet(t, server.URL, clientTLS)
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("ReadAll: %v", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", resp.StatusCode, string(body))
	}

	if string(body) != "yarumo-mtls-ok" {
		t.Fatalf("unexpected body: got %q", string(body))
	}
}

// doGet performs an HTTPS GET with the supplied TLS config and fails the test
// on transport error.
func doGet(t *testing.T, url string, tlsConfig *tls.Config) *http.Response {
	t.Helper()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do: %v", err)
	}

	return resp
}
