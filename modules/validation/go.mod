module github.com/guidomantilla/yarumo/validation

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../common

replace github.com/guidomantilla/yarumo/log => ../log

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	go.yaml.in/yaml/v3 v3.0.4
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
