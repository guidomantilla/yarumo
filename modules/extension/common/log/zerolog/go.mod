module github.com/guidomantilla/yarumo/extension/common/log/zerolog

go 1.25.5

replace github.com/guidomantilla/yarumo/core/common => ../../../../core/common

require (
	github.com/guidomantilla/yarumo/core/common v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.34.0
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.42.0 // indirect
)
