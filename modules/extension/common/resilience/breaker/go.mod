module github.com/guidomantilla/yarumo/extension/common/resilience/breaker

go 1.25.5

replace github.com/guidomantilla/yarumo/core/common => ../../../../core/common

require (
	github.com/guidomantilla/yarumo/core/common v0.0.0-00010101000000-000000000000
	github.com/sony/gobreaker v1.0.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/text v0.34.0 // indirect
)
