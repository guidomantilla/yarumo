// Package main demonstrates common/rest: declarative request/response
// specs, the typed Call helper that decodes JSON into a generic T, the
// CallStream variant for raw bodies, and the HTTPError sentinel returned
// on non-2xx responses. All requests target net/http/httptest servers.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	crest "github.com/guidomantilla/yarumo/common/rest"
)

// Pokemon is the JSON shape returned by the demo server.
type Pokemon struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

func main() {
	server := newDemoServer()
	defer server.Close()

	ctx := context.Background()

	demoCallTyped(ctx, server.URL)
	demoCallStream(ctx, server.URL)
	demoHTTPError(ctx, server.URL)
	demoQueryParams(ctx, server.URL)
}

// demoCallTyped issues a GET that decodes JSON into Pokemon.
func demoCallTyped(ctx context.Context, baseURL string) {
	fmt.Println("=== Call[Pokemon] ===")

	spec := &crest.RequestSpec{
		Method: http.MethodGet,
		URL:    baseURL,
		Path:   "/pokemon/25",
	}

	resp, err := crest.Call[Pokemon](ctx, spec)
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}

	fmt.Printf("  status=%d body=%+v\n", resp.Code, resp.Body)
}

// demoCallStream returns the raw body so the caller can read it
// incrementally — useful for SSE or NDJSON.
func demoCallStream(ctx context.Context, baseURL string) {
	fmt.Println("=== CallStream ===")

	spec := &crest.RequestSpec{
		Method: http.MethodGet,
		URL:    baseURL,
		Path:   "/raw",
	}

	resp, err := crest.CallStream(ctx, spec)
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  status=%d body=%s\n", resp.Code, string(body))
}

// demoHTTPError pulls a 404 and recovers the structured error.
func demoHTTPError(ctx context.Context, baseURL string) {
	fmt.Println("=== HTTPError ===")

	spec := &crest.RequestSpec{
		Method: http.MethodGet,
		URL:    baseURL,
		Path:   "/missing",
	}

	_, err := crest.Call[Pokemon](ctx, spec)

	var httpErr *crest.HTTPError
	if errors.As(err, &httpErr) {
		fmt.Printf("  HTTPError code=%d body=%s\n", httpErr.StatusCode, string(httpErr.Body))
	}
}

// demoQueryParams shows query encoding via RequestSpec.QueryParams.
func demoQueryParams(ctx context.Context, baseURL string) {
	fmt.Println("=== QueryParams ===")

	spec := &crest.RequestSpec{
		Method: http.MethodGet,
		URL:    baseURL,
		Path:   "/echo",
		QueryParams: map[string][]string{
			"q":   {"pikachu"},
			"lim": {"5"},
		},
	}

	resp, err := crest.Call[map[string]any](ctx, spec)
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}

	fmt.Printf("  echoed query: %v\n", resp.Body["query"])
}

// newDemoServer returns an httptest.Server that handles every route used by main.
func newDemoServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/pokemon/25", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Pokemon{ID: 25, Name: "Pikachu", Level: 35})
	})

	mux.HandleFunc("/raw", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello stream"))
	})

	mux.HandleFunc("/missing", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})

	mux.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"query": req.URL.RawQuery})
	})

	return httptest.NewServer(mux)
}
