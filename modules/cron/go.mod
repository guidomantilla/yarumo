module github.com/guidomantilla/yarumo/cron

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../common

replace github.com/guidomantilla/yarumo/log => ../log

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/robfig/cron/v3 v3.0.1
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
