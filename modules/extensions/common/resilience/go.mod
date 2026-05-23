module github.com/guidomantilla/yarumo/extensions/common/resilience

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../common

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/sony/gobreaker v1.0.0
	golang.org/x/time v0.15.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/text v0.34.0 // indirect
)
