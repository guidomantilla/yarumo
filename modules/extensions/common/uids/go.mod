module github.com/guidomantilla/yarumo/extensions/common/uids

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../../common

require (
	github.com/akshayvadher/cuid2 v0.0.0-20241212114603-8aba656b70dc
	github.com/devmiek/nanoid-go v0.0.0-20241216084707-e17e38258ffc
	github.com/google/uuid v1.6.0
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/oklog/ulid/v2 v2.1.1
	github.com/rs/xid v1.6.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
