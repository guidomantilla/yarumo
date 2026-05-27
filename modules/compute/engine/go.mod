module github.com/guidomantilla/yarumo/compute/engine

go 1.25.5

require (
	github.com/guidomantilla/yarumo/compute/math v0.0.0
	github.com/guidomantilla/yarumo/core/common v0.0.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/guidomantilla/yarumo/compute/math => ../math
	github.com/guidomantilla/yarumo/core/common => ../../core/common
)
