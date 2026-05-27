// Demo that exercises the public API of the security/authn/grpc
// interceptors using bufconn for an in-process gRPC server.
//
//  1. NewUnaryInterceptor wraps a fake Authenticator. The server-side
//     health.Check RPC reads the principal from ctx.
//  2. Request with no metadata -> Unauthenticated.
//  3. Request with malformed authorization metadata -> Unauthenticated.
//  4. Request with a valid Bearer token -> OK.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/guidomantilla/yarumo/config"
	authngrpc "github.com/guidomantilla/yarumo/extension/security/authn/grpc"
	"github.com/guidomantilla/yarumo/security/authn"
)

const bufSize = 1024 * 1024

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/security/authn/grpc/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	lis := bufconn.Listen(bufSize)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authngrpc.NewUnaryInterceptor(fakeAuthenticator{})),
		grpc.StreamInterceptor(authngrpc.NewStreamInterceptor(fakeAuthenticator{})),
	)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("demo", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthSrv)

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve(lis)
	}()

	defer func() {
		server.GracefulStop()
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	client := healthpb.NewHealthClient(conn)

	demos := []struct {
		title string
		fn    func(context.Context, healthpb.HealthClient) error
	}{
		{"Missing metadata -> Unauthenticated", demoMissingMetadata},
		{"Malformed metadata -> Unauthenticated", demoMalformedMetadata},
		{"Valid bearer -> OK", demoValidBearer},
	}

	for _, d := range demos {
		fmt.Printf("=== Demo: %s ===\n", d.title)
		err := d.fn(ctx, client)
		if err != nil {
			return fmt.Errorf("%s: %w", d.title, err)
		}
		fmt.Println()
	}

	return nil
}

// fakeAuthenticator accepts "ok-token" and rejects everything else.
type fakeAuthenticator struct{}

func (fakeAuthenticator) Validate(_ context.Context, token string) (*authn.Principal, error) {
	if token != "ok-token" {
		return nil, authn.ErrAuthentication(authn.ErrTokenInvalid)
	}
	return &authn.Principal{
		ID:         "u-42",
		Name:       "Alice",
		Roles:      []string{"admin"},
		Attributes: map[string]any{"tenant": "acme"},
	}, nil
}

// demoMissingMetadata invokes Check with no outgoing metadata.
func demoMissingMetadata(ctx context.Context, client healthpb.HealthClient) error {
	_, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: "demo"})
	if err == nil {
		return errors.New("expected Unauthenticated, got nil")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		return fmt.Errorf("expected Unauthenticated, got %v", err)
	}

	fmt.Printf("  rejected: code=%s msg=%q\n", st.Code(), st.Message())
	return nil
}

// demoMalformedMetadata sends a Basic scheme instead of Bearer.
func demoMalformedMetadata(ctx context.Context, client healthpb.HealthClient) error {
	md := metadata.New(map[string]string{"authorization": "Basic dXNlcjpwYXNz"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: "demo"})
	if err == nil {
		return errors.New("expected Unauthenticated, got nil")
	}

	st, _ := status.FromError(err)
	fmt.Printf("  rejected: code=%s msg=%q\n", st.Code(), st.Message())

	if st.Code() != codes.Unauthenticated {
		return fmt.Errorf("expected Unauthenticated, got %s", st.Code())
	}
	return nil
}

// demoValidBearer sends the magic token; the RPC should succeed.
func demoValidBearer(ctx context.Context, client healthpb.HealthClient) error {
	md := metadata.New(map[string]string{"authorization": "Bearer ok-token"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: "demo"})
	if err != nil {
		return fmt.Errorf("Check: %w", err)
	}

	fmt.Printf("  Check -> %s\n", resp.GetStatus())
	return nil
}
