module github.com/guidomantilla/yarumo/extension/common/http/retry

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../../common

replace github.com/guidomantilla/yarumo/extension/common/resilience/retry => ../../resilience/retry

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/extension/common/resilience/retry v0.0.0-00010101000000-000000000000
)

require (
	github.com/avast/retry-go/v4 v4.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
