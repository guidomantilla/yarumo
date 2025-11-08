package certs

import "crypto/tls"

var (
	_ TlsFn = Tls
)

type TlsFn func(serverName string, caCertificate string, clientCertificate string, clientKey string, insecureSkipVerify bool) (*tls.Config, error)
