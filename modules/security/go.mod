module github.com/guidomantilla/yarumo/security

go 1.25.2

replace github.com/guidomantilla/yarumo/common => ../common

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/guidomantilla/yarumo/common v0.0.0
	golang.org/x/crypto v0.45.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)
