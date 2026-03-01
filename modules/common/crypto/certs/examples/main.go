package main

import (
	"crypto/tls"
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
