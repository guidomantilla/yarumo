module github.com/guidomantilla/yarumo/uids/xid

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../common

replace github.com/guidomantilla/yarumo/uids => ../

require (
	github.com/guidomantilla/yarumo/uids v0.0.0-00010101000000-000000000000
	github.com/rs/xid v1.6.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/text v0.34.0 // indirect
)
