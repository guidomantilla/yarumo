package jwt

import (
	"github.com/guidomantilla/yarumo/security/authn"
)

var (
	_ authn.Authenticator = (*jwtAuthenticator)(nil)
)
