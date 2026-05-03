module github.com/guidomantilla/yarumo/internal/deprecated/servers

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../modules/common

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
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
