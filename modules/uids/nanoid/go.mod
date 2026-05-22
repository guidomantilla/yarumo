module github.com/guidomantilla/yarumo/uids/nanoid

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../common

replace github.com/guidomantilla/yarumo/uids => ../

require (
	github.com/devmiek/nanoid-go v0.0.0-20241216084707-e17e38258ffc
	github.com/guidomantilla/yarumo/uids v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/text v0.34.0 // indirect
)
