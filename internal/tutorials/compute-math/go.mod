module github.com/guidomantilla/yarumo/internal/tutorials/compute-math

go 1.25.5

require github.com/guidomantilla/yarumo/compute/math v0.0.0

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/common v0.0.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/guidomantilla/yarumo/common => ../../../modules/common
	github.com/guidomantilla/yarumo/compute/math => ../../../modules/compute/math
	github.com/guidomantilla/yarumo/log => ../../../modules/log
)
