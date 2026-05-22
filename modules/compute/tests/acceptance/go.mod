module github.com/guidomantilla/yarumo/compute/tests/acceptance

go 1.25.5

require (
	github.com/guidomantilla/yarumo/compute/engine v0.0.0
	github.com/guidomantilla/yarumo/compute/math v0.0.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/common v0.0.0 // indirect
	github.com/guidomantilla/yarumo/log v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/guidomantilla/yarumo/common => ../../../common
	github.com/guidomantilla/yarumo/compute/engine => ../../engine
	github.com/guidomantilla/yarumo/compute/math => ../../math
	github.com/guidomantilla/yarumo/log => ../../../log
)
