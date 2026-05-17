module github.com/guidomantilla/yarumo/cache

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../common

require (
	github.com/dgraph-io/ristretto/v2 v2.4.0
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/alicebob/miniredis/v2 v2.38.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/redis/go-redis/v9 v9.19.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
