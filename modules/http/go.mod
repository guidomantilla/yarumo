module github.com/guidomantilla/yarumo/http

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../common

replace github.com/guidomantilla/yarumo/log => ../log

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/log v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
