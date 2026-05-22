module github.com/guidomantilla/yarumo/decisions/core

go 1.25.5

require (
	github.com/guidomantilla/yarumo/common v0.0.0
	github.com/guidomantilla/yarumo/compute/engine v0.0.0
	github.com/guidomantilla/yarumo/compute/math v0.0.0
	github.com/guidomantilla/yarumo/uids/uuid v0.0.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/guidomantilla/yarumo/uids v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/guidomantilla/yarumo/common => ../../../modules/common
	github.com/guidomantilla/yarumo/compute/engine => ../../../modules/compute/engine
	github.com/guidomantilla/yarumo/compute/math => ../../../modules/compute/math
	github.com/guidomantilla/yarumo/uids => ../../../modules/uids
	github.com/guidomantilla/yarumo/uids/uuid => ../../../modules/uids/uuid
)
