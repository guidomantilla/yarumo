module github.com/guidomantilla/yarumo/managed/cache/ristretto/examples

go 1.25.5

replace github.com/guidomantilla/yarumo/managed/cache/ristretto => ../

replace github.com/guidomantilla/yarumo/common => ../../../../common

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/managed/cache/ristretto v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgraph-io/ristretto/v2 v2.4.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
