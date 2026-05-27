module github.com/guidomantilla/yarumo/extension/common/http/breaker

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../../common

replace github.com/guidomantilla/yarumo/extension/common/resilience/breaker => ../../resilience/breaker

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/extension/common/resilience/breaker v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
