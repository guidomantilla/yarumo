package token

import (
	"github.com/guidomantilla/yarumo/security/authn"
)

var (
	_ authn.Authenticator = (*tokenAuthenticator)(nil)
)
