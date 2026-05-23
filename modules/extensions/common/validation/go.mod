module github.com/guidomantilla/yarumo/extensions/common/validation

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../common

require (
	github.com/google/uuid v1.6.0
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/oklog/ulid/v2 v2.1.1
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
