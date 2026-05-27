module github.com/guidomantilla/yarumo/extension/common/cache/redis/examples

go 1.25.5

replace github.com/guidomantilla/yarumo/extension/common/cache/redis => ..

replace github.com/guidomantilla/yarumo/core/common => ../../../../../core/common

require (
	github.com/alicebob/miniredis/v2 v2.38.0
	github.com/guidomantilla/yarumo/core/common v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/extension/common/cache/redis v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/redis/go-redis/v9 v9.19.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
