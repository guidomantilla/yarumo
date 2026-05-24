module github.com/guidomantilla/yarumo/extensions/common/http/limiter

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../../common

replace github.com/guidomantilla/yarumo/extensions/common/resilience/limiter => ../../resilience/limiter

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/extensions/common/resilience/limiter v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	golang.org/x/time v0.15.0 // indirect
)
