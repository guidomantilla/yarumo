module github.com/guidomantilla/yarumo/messaging/examples

go 1.25.5

replace github.com/guidomantilla/yarumo/messaging => ../

replace github.com/guidomantilla/yarumo/common => ../../common

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/messaging v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
