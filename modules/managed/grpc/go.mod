module github.com/guidomantilla/yarumo/managed/grpc

go 1.25.5

replace github.com/guidomantilla/yarumo/core/common => ../../core/common

require (
	github.com/guidomantilla/yarumo/core/common v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.1
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
