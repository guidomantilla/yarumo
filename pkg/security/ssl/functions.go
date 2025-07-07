package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func Tls(serverName string, caCertificate string, clientCertificate string, clientKey string, insecureSkipVerify bool) (*tls.Config, error) {

	if serverName == "" {
		return nil, fmt.Errorf("ssl - error setting up tls: serverName is empty")
	}

	if caCertificate == "" {
		return nil, fmt.Errorf("ssl - error setting up tls: caCertificate is empty")
	}

	if clientCertificate == "" {
		return nil, fmt.Errorf("ssl - error setting up tls: clientCertificate is empty")
	}

	if clientKey == "" {
		return nil, fmt.Errorf("ssl - error setting up tls: clientKey is empty")
	}

	caCert, err := os.ReadFile(caCertificate)
	if err != nil {
		return nil, fmt.Errorf("ca certificate: %s", err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(clientCertificate, clientKey)
	if err != nil {
		return nil, fmt.Errorf("client certificate: %s", err.Error())
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
