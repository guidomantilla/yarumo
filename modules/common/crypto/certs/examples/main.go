package main

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	ccerts "github.com/guidomantilla/yarumo/common/crypto/certs"
)

func main() {
	selfSignedExample()
	selfSignedCAExample()
	newPoolExample()
	tlsFromPEMExample()
	serverTlsFromPEMExample()
	tlsFromFilesExample()
	multiAlgorithmExample()
	loadCertAndKeyExample()
	csrExample()
}

// selfSignedExample demonstrates generating a self-signed certificate with default and custom options.
func selfSignedExample() {
	fmt.Println("=== SelfSigned (default options) ===")

	certPEM, keyPEM, err := ccerts.SelfSigned()
	if err != nil {
		log.Fatalf("SelfSigned failed: %v", err)
	}

	fmt.Printf("Certificate PEM: %d bytes\n", len(certPEM))
	fmt.Printf("Private Key PEM: %d bytes\n\n", len(keyPEM))

	fmt.Println("=== SelfSigned (custom options) ===")

	certPEM, keyPEM, err = ccerts.SelfSigned(
		ccerts.WithOrganization("Yarumo Inc"),
		ccerts.WithValidity(24*time.Hour),
		ccerts.WithDNSNames("example.com", "*.example.com"),
		ccerts.WithIPAddresses(net.IPv4(192, 168, 1, 1)),
	)
	if err != nil {
		log.Fatalf("SelfSigned (custom) failed: %v", err)
	}

	fmt.Printf("Certificate PEM: %d bytes\n", len(certPEM))
	fmt.Printf("Private Key PEM: %d bytes\n\n", len(keyPEM))
}

// selfSignedCAExample demonstrates generating a CA certificate.
func selfSignedCAExample() {
	fmt.Println("=== SelfSigned (CA certificate) ===")

	certPEM, keyPEM, err := ccerts.SelfSigned(
		ccerts.WithOrganization("Yarumo CA"),
		ccerts.WithIsCA(true),
	)
	if err != nil {
		log.Fatalf("SelfSigned CA failed: %v", err)
	}

	fmt.Printf("CA Certificate PEM: %d bytes\n", len(certPEM))
	fmt.Printf("CA Private Key PEM: %d bytes\n\n", len(keyPEM))
}

// newPoolExample demonstrates building a certificate pool from PEM data.
func newPoolExample() {
	fmt.Println("=== NewPool ===")

	cert1, _, err := ccerts.SelfSigned(ccerts.WithOrganization("CA One"))
	if err != nil {
		log.Fatalf("generating cert1 failed: %v", err)
	}

	cert2, _, err := ccerts.SelfSigned(ccerts.WithOrganization("CA Two"))
	if err != nil {
		log.Fatalf("generating cert2 failed: %v", err)
	}

	pool, err := ccerts.NewPool(cert1, cert2)
	if err != nil {
		log.Fatalf("NewPool failed: %v", err)
	}

	fmt.Printf("Pool created successfully (non-nil: %v)\n", pool != nil)
	fmt.Println()
}

// tlsFromPEMExample demonstrates building a client TLS config from PEM-encoded bytes.
func tlsFromPEMExample() {
	fmt.Println("=== ClientTlsFromPEM ===")

	caCert, caKey, err := ccerts.SelfSigned(ccerts.WithIsCA(true))
	if err != nil {
		log.Fatalf("generating CA cert failed: %v", err)
	}

	// For this example, use the CA cert/key as both the CA and the client certificate.
	_ = caKey
	clientCert, clientKey, err := ccerts.SelfSigned()
	if err != nil {
		log.Fatalf("generating client cert failed: %v", err)
	}

	tlsConfig, err := ccerts.ClientTlsFromPEM("localhost", caCert, clientCert, clientKey, true)
	if err != nil {
		log.Fatalf("ClientTlsFromPEM failed: %v", err)
	}

	fmt.Printf("TLS Config: MinVersion=0x%04x, ServerName=%q, InsecureSkipVerify=%v\n",
		tlsConfig.MinVersion, tlsConfig.ServerName, tlsConfig.InsecureSkipVerify)
	fmt.Printf("RootCAs: present=%v, Certificates: count=%d\n\n",
		tlsConfig.RootCAs != nil, len(tlsConfig.Certificates))
}

// serverTlsFromPEMExample demonstrates building a server TLS config from PEM-encoded bytes.
func serverTlsFromPEMExample() {
	fmt.Println("=== ServerTlsFromPEM ===")

	certPEM, keyPEM, err := ccerts.SelfSigned()
	if err != nil {
		log.Fatalf("generating server cert failed: %v", err)
	}

	// Server TLS without mTLS
	serverConfig, err := ccerts.ServerTlsFromPEM(certPEM, keyPEM, nil)
	if err != nil {
		log.Fatalf("ServerTlsFromPEM failed: %v", err)
	}

	fmt.Printf("Server TLS Config: MinVersion=0x%04x, Certificates=%d, ClientAuth=%v\n",
		serverConfig.MinVersion, len(serverConfig.Certificates), serverConfig.ClientAuth)

	// Server TLS with mTLS (client CA verification)
	caCert, _, err := ccerts.SelfSigned(ccerts.WithIsCA(true))
	if err != nil {
		log.Fatalf("generating CA cert failed: %v", err)
	}

	mtlsConfig, err := ccerts.ServerTlsFromPEM(certPEM, keyPEM, caCert)
	if err != nil {
		log.Fatalf("ServerTlsFromPEM (mTLS) failed: %v", err)
	}

	fmt.Printf("Server mTLS Config: ClientAuth=%v, ClientCAs present=%v\n\n",
		mtlsConfig.ClientAuth == tls.RequireAndVerifyClientCert, mtlsConfig.ClientCAs != nil)
}

// tlsFromFilesExample demonstrates building client and server TLS configs from certificate files.
func tlsFromFilesExample() {
	fmt.Println("=== ClientTls & ServerTls (from files) ===")

	tmpDir, err := os.MkdirTemp("", "certs-example-*")
	if err != nil {
		log.Fatalf("creating temp dir failed: %v", err)
	}

	certPEM, keyPEM, err := ccerts.SelfSigned(ccerts.WithIsCA(true))
	if err != nil {
		log.Fatalf("generating cert failed: %v", err)
	}

	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")

	err = os.WriteFile(certFile, certPEM, 0o600)
	if err != nil {
		log.Fatalf("writing cert file failed: %v", err)
	}

	err = os.WriteFile(keyFile, keyPEM, 0o600)
	if err != nil {
		log.Fatalf("writing key file failed: %v", err)
	}

	// Server TLS config
	serverConfig, err := ccerts.ServerTls(certFile, keyFile, "")
	if err != nil {
		log.Fatalf("ServerTls failed: %v", err)
	}

	fmt.Printf("Server TLS Config: MinVersion=0x%04x, Certificates=%d, ClientAuth=%v\n",
		serverConfig.MinVersion, len(serverConfig.Certificates), serverConfig.ClientAuth)

	// Client TLS config
	clientConfig, err := ccerts.ClientTls("localhost", certFile, certFile, keyFile, true)
	if err != nil {
		log.Fatalf("ClientTls failed: %v", err)
	}

	fmt.Printf("Client TLS Config: MinVersion=0x%04x, ServerName=%q, InsecureSkipVerify=%v\n",
		clientConfig.MinVersion, clientConfig.ServerName, clientConfig.InsecureSkipVerify)

	// Server TLS with mTLS (client CA verification)
	mtlsConfig, err := ccerts.ServerTls(certFile, keyFile, certFile)
	if err != nil {
		log.Fatalf("ServerTls (mTLS) failed: %v", err)
	}

	fmt.Printf("Server mTLS Config: ClientAuth=%v, ClientCAs present=%v\n",
		mtlsConfig.ClientAuth == tls.RequireAndVerifyClientCert, mtlsConfig.ClientCAs != nil)

	_ = os.RemoveAll(tmpDir)
	fmt.Println()
}

// multiAlgorithmExample demonstrates generating self-signed certs with different key algorithms.
func multiAlgorithmExample() {
	fmt.Println("=== SelfSigned (multi-algorithm) ===")

	algos := []ccerts.KeyAlgorithm{
		ccerts.KeyAlgorithmECDSAP256,
		ccerts.KeyAlgorithmECDSAP384,
		ccerts.KeyAlgorithmEd25519,
		ccerts.KeyAlgorithmRSA2048,
		ccerts.KeyAlgorithmRSA3072,
	}

	for _, algo := range algos {
		certPEM, keyPEM, err := ccerts.SelfSigned(
			ccerts.WithKeyAlgorithm(algo),
			ccerts.WithValidity(time.Hour),
		)
		if err != nil {
			log.Fatalf("SelfSigned (%s) failed: %v", algo, err)
		}

		fmt.Printf("  %-12s cert=%d bytes, key=%d bytes\n", algo, len(certPEM), len(keyPEM))
	}

	fmt.Println()
}

// writeMTLSKeypair writes a CA + client mTLS keypair to tmpDir and returns the file paths.
func writeMTLSKeypair(tmpDir string) (string, string, string) {
	caCertPEM, _, err := ccerts.SelfSigned(
		ccerts.WithOrganization("Yarumo CA"),
		ccerts.WithIsCA(true),
		ccerts.WithKeyAlgorithm(ccerts.KeyAlgorithmEd25519),
	)
	if err != nil {
		log.Fatalf("generating CA cert failed: %v", err)
	}

	clientCertPEM, clientKeyPEM, err := ccerts.SelfSigned(
		ccerts.WithOrganization("Yarumo Client"),
		ccerts.WithKeyAlgorithm(ccerts.KeyAlgorithmEd25519),
	)
	if err != nil {
		log.Fatalf("generating client cert failed: %v", err)
	}

	caFile := filepath.Join(tmpDir, "ca.pem")
	certFile := filepath.Join(tmpDir, "client.pem")
	keyFile := filepath.Join(tmpDir, "client-key.pem")

	for path, data := range map[string][]byte{
		caFile:   caCertPEM,
		certFile: clientCertPEM,
		keyFile:  clientKeyPEM,
	} {
		writeErr := os.WriteFile(path, data, 0o600)
		if writeErr != nil {
			log.Fatalf("writing %s failed: %v", path, writeErr)
		}
	}

	return caFile, certFile, keyFile
}

// loadCertAndKeyExample demonstrates loading an mTLS keypair from a temp dir and building a TLS config.
func loadCertAndKeyExample() {
	fmt.Println("=== Load mTLS keypair from disk ===")

	tmpDir, err := os.MkdirTemp("", "certs-load-example-*")
	if err != nil {
		log.Fatalf("creating temp dir failed: %v", err)
	}

	caFile, certFile, keyFile := writeMTLSKeypair(tmpDir)

	caCert, err := ccerts.LoadCertificate(caFile)
	if err != nil {
		log.Fatalf("LoadCertificate failed: %v", err)
	}

	clientCert, err := ccerts.LoadCertificate(certFile)
	if err != nil {
		log.Fatalf("LoadCertificate (client) failed: %v", err)
	}

	clientKey, err := ccerts.LoadPrivateKey(keyFile)
	if err != nil {
		log.Fatalf("LoadPrivateKey failed: %v", err)
	}

	tlsConfig, err := ccerts.ClientTls("yarumo.test", caFile, certFile, keyFile, true)
	if err != nil {
		log.Fatalf("ClientTls failed: %v", err)
	}

	fmt.Printf("CA subject:         %s\n", caCert.Subject.Organization)
	fmt.Printf("Client subject:     %s\n", clientCert.Subject.Organization)
	fmt.Printf("Client key type:    %T\n", clientKey)
	fmt.Printf("TLS Config:         MinVersion=0x%04x, ServerName=%q, Certificates=%d\n\n",
		tlsConfig.MinVersion, tlsConfig.ServerName, len(tlsConfig.Certificates))

	_ = os.RemoveAll(tmpDir)
}

// csrExample demonstrates generating and parsing a CSR with a caller-supplied key.
func csrExample() {
	fmt.Println("=== GenerateCSR & ParseCSR ===")

	_, keyPEM, err := ccerts.SelfSigned(ccerts.WithKeyAlgorithm(ccerts.KeyAlgorithmECDSAP256))
	if err != nil {
		log.Fatalf("SelfSigned failed: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "certs-csr-example-*")
	if err != nil {
		log.Fatalf("creating temp dir failed: %v", err)
	}

	keyFile := filepath.Join(tmpDir, "key.pem")
	writeErr := os.WriteFile(keyFile, keyPEM, 0o600)
	if writeErr != nil {
		_ = os.RemoveAll(tmpDir)
		log.Fatalf("writing key failed: %v", writeErr)
	}

	key, err := ccerts.LoadPrivateKey(keyFile)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		log.Fatalf("LoadPrivateKey failed: %v", err)
	}

	csrPEM, err := ccerts.GenerateCSR(
		ccerts.WithCSRPrivateKey(key),
		ccerts.WithCSRSubject(pkix.Name{
			CommonName:   "service.yarumo.test",
			Organization: []string{"Yarumo"},
		}),
		ccerts.WithCSRDNSNames("service.yarumo.test"),
		ccerts.WithCSRIPAddresses(net.IPv4(127, 0, 0, 1)),
	)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		log.Fatalf("GenerateCSR failed: %v", err)
	}

	csr, err := ccerts.ParseCSR(csrPEM)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		log.Fatalf("ParseCSR failed: %v", err)
	}

	fmt.Printf("CSR CN:       %s\n", csr.Subject.CommonName)
	fmt.Printf("CSR DNS SANs: %v\n", csr.DNSNames)
	fmt.Printf("CSR IP SANs:  %v\n", csr.IPAddresses)
	fmt.Println()

	_ = os.RemoveAll(tmpDir)
}
