module github.com/guidomantilla/yarumo/servers

go 1.25.2

replace github.com/guidomantilla/yarumo/common => ../common

require (
	github.com/guidomantilla/yarumo/common v0.0.0
	github.com/qmdx00/lifecycle v1.1.1
	github.com/rs/zerolog v1.34.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
)
