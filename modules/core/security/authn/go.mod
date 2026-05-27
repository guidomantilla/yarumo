module github.com/guidomantilla/yarumo/core/security/authn

go 1.25.5

replace github.com/guidomantilla/yarumo/core/common => ../../common

replace github.com/guidomantilla/yarumo/core/crypto => ../../crypto

require (
	github.com/guidomantilla/yarumo/core/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/core/crypto v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
