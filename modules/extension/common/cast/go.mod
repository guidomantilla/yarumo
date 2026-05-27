module github.com/guidomantilla/yarumo/extension/common/cast

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../common

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/spf13/cast v1.10.0
)

require github.com/rogpeppe/go-internal v1.14.1 // indirect
