module github.com/guidomantilla/yarumo/managed

go 1.25.5

replace github.com/guidomantilla/yarumo/common => ../common

require (
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.34.0
)

require (
	github.com/avast/retry-go/v4 v4.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
