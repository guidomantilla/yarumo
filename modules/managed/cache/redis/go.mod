module github.com/guidomantilla/yarumo/managed/cache/redis

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../common

require (
	github.com/alicebob/miniredis/v2 v2.38.0
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.19.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
