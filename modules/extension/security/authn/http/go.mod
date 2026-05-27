module github.com/guidomantilla/yarumo/extension/security/authn/http

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../../common

replace github.com/guidomantilla/yarumo/security/authn => ../../../../security/authn

replace github.com/guidomantilla/yarumo/crypto => ../../../../crypto

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/security/authn v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/crypto v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
