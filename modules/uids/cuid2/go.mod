module github.com/guidomantilla/yarumo/uids/cuid2

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../../common

replace github.com/guidomantilla/yarumo/uids => ../

require (
	github.com/akshayvadher/cuid2 v0.0.0-20241212114603-8aba656b70dc
	github.com/guidomantilla/yarumo/uids v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
