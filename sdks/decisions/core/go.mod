module github.com/guidomantilla/yarumo/decisions/core

go 1.25.5

require (
	github.com/guidomantilla/yarumo/common v0.0.0
	github.com/guidomantilla/yarumo/compute/engine v0.0.0
	github.com/guidomantilla/yarumo/compute/math v0.0.0
)

require (
	github.com/akshayvadher/cuid2 v0.0.0-20241212114603-8aba656b70dc // indirect
	github.com/devmiek/nanoid-go v0.0.0-20241216084707-e17e38258ffc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace (
	github.com/guidomantilla/yarumo/common => ../../../modules/common
	github.com/guidomantilla/yarumo/compute/engine => ../../../modules/compute/engine
	github.com/guidomantilla/yarumo/compute/math => ../../../modules/compute/math
	github.com/guidomantilla/yarumo/log => ../../../modules/log
)
