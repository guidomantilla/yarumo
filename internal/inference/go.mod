module github.com/guidomantilla/yarumo/internal/inference

go 1.25.5

require (
	github.com/guidomantilla/yarumo/inference v0.0.0
	github.com/guidomantilla/yarumo/maths v0.0.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/common v0.0.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/guidomantilla/yarumo/common => ../../modules/common
	github.com/guidomantilla/yarumo/inference => ../../modules/inference
	github.com/guidomantilla/yarumo/maths => ../../modules/maths
)
