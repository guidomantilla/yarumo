module github.com/guidomantilla/yarumo/core/crypto

go 1.25.5

replace github.com/guidomantilla/yarumo/core/common => ../common

require (
	github.com/cloudflare/circl v1.6.3
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/guidomantilla/yarumo/core/common v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.48.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
