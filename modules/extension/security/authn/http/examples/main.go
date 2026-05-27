// Demo that exercises the public API of the security/authn/http
// middleware:
//
//  1. NewMiddleware wraps a downstream handler with Bearer
//     authentication. A fake Authenticator validates the token "ok-token"
//     and returns a non-nil *Principal; every other input fails.
//  2. Request with no Authorization header -> 401.
//  3. Request with malformed Authorization header -> 401.
//  4. Request with a valid Bearer token -> 200 + principal in ctx.
//  5. WithErrorHandler customizes the 401 response body.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/guidomantilla/yarumo/config"
	authnhttp "github.com/guidomantilla/yarumo/extension/security/authn/http"
	"github.com/guidomantilla/yarumo/core/security/authn"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/security/authn/http/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Missing Authorization header -> 401", demoMissingHeader},
		{"Malformed header -> 401", demoMalformedHeader},
		{"Valid bearer -> 200 + principal", demoValidBearer},
		{"WithErrorHandler customizes 401 body", demoCustomErrorHandler},
	}

	for _, d := range demos {
		fmt.Printf("=== Demo: %s ===\n", d.title)
		err := d.fn(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", d.title, err)
		}
		fmt.Println()
	}

	return nil
}

// fakeAuthenticator implements authn.Authenticator. It accepts the
// token "ok-token" and rejects everything else with ErrTokenInvalid.
type fakeAuthenticator struct{}

func (fakeAuthenticator) Validate(_ context.Context, token string) (*authn.Principal, error) {
	if token != "ok-token" {
		return nil, authn.ErrAuthentication(authn.ErrTokenInvalid)
	}
	return &authn.Principal{
		ID:         "u-42",
		Name:       "Alice",
		Roles:      []string{"admin", "ops"},
		Attributes: map[string]any{"tenant": "acme"},
	}, nil
}

// downstreamHandler reads the principal from ctx (if any) and renders
// it in the response.
func downstreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := authn.FromContext(r.Context())
		if !ok {
			http.Error(w, "no principal in ctx", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "hello %s (id=%s, roles=%v)", principal.Name, principal.ID, principal.Roles)
	})
}

// buildServer wires the middleware in front of the downstream handler
// and starts an httptest.Server.
func buildServer(opts ...authnhttp.Option) *httptest.Server {
	mw := authnhttp.NewMiddleware(fakeAuthenticator{}, opts...)
	return httptest.NewServer(mw(downstreamHandler()))
}

// demoMissingHeader fires a request with no Authorization header and
// expects 401.
func demoMissingHeader(ctx context.Context) error {
	server := buildServer()
	defer server.Close()

	resp, body, err := fetch(ctx, server.URL, "")
	if err != nil {
		return err
	}
	fmt.Printf("  %d %q\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401, got %d", resp.StatusCode)
	}
	return nil
}

// demoMalformedHeader sends a Basic credential instead of Bearer.
func demoMalformedHeader(ctx context.Context) error {
	server := buildServer()
	defer server.Close()

	resp, body, err := fetch(ctx, server.URL, "Basic dXNlcjpwYXNz")
	if err != nil {
		return err
	}
	fmt.Printf("  %d %q\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401, got %d", resp.StatusCode)
	}
	return nil
}

// demoValidBearer sends the magic token; expects 200 and a body that
// echoes the principal.
func demoValidBearer(ctx context.Context) error {
	server := buildServer()
	defer server.Close()

	resp, body, err := fetch(ctx, server.URL, "Bearer ok-token")
	if err != nil {
		return err
	}
	fmt.Printf("  %d %q\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200, got %d", resp.StatusCode)
	}
	return nil
}

// demoCustomErrorHandler installs a custom ErrorHandler that writes a
// JSON-shaped 401 body so callers can see how to customize.
func demoCustomErrorHandler(ctx context.Context) error {
	server := buildServer(authnhttp.WithErrorHandler(func(w http.ResponseWriter, _ *http.Request, cause error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"error":"unauthorized","cause":%q}`, cause.Error())
	}))
	defer server.Close()

	resp, body, err := fetch(ctx, server.URL, "Bearer wrong-token")
	if err != nil {
		return err
	}
	fmt.Printf("  %d %q\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401, got %d", resp.StatusCode)
	}
	return nil
}

// fetch issues a GET to url with optional Authorization header.
func fetch(ctx context.Context, url, auth string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return resp, body, nil
}
