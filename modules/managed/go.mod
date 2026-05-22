module github.com/guidomantilla/yarumo/managed

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../common

require github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000

require (
	github.com/avast/retry-go/v4 v4.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
