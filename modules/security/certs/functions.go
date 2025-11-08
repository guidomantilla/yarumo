package certs

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/guidomantilla/yarumo/common/utils"
)

func Tls(serverName string, caCertificate string, clientCertificate string, clientKey string, insecureSkipVerify bool) (*tls.Config, error) {
	if utils.Empty(serverName) {
		return nil, ErrCertificateTLS(fmt.Errorf("serverName cannot be empty"))
	}
	if utils.Empty(caCertificate) {
		return nil, ErrCertificateTLS(fmt.Errorf("caCertificate is empty"))
	}
	if utils.Empty(clientCertificate) {
		return nil, ErrCertificateTLS(fmt.Errorf("clientCertificate is empty"))
	}
	if utils.Empty(clientKey) {
		return nil, ErrCertificateTLS(fmt.Errorf("clientKey is empty"))
	}

	caCert, err := os.ReadFile(caCertificate)
	if err != nil {
		return nil, ErrCertificateCaTLS(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(clientCertificate, clientKey)
	if err != nil {
		return nil, ErrCertificateClientTLS(err)
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
