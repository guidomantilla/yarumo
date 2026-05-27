// Package main demonstrates common/http: the Client contract, NewClient
// with WithTimeout / WithTransport, and the RoundTripperFn adapter that
// turns a plain function into an http.RoundTripper for middleware-style
// composition. The example never touches the network — all requests go
// through net/http/httptest.NewServer.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	chttp "github.com/guidomantilla/yarumo/common/http"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "hello from %s", req.URL.Path)
	}))
	defer server.Close()

	demoDefaultClient(server.URL)
	demoCustomTimeout(server.URL)
	demoRoundTripperFn()
	demoTracingTransport(server.URL)
}

// demoDefaultClient builds a client with default options (30s timeout,
// http.DefaultTransport) and performs a GET against the test server.
func demoDefaultClient(baseURL string) {
	fmt.Println("=== Default client ===")

	client := chttp.NewClient()

	resp, err := client.Do(mustGet(baseURL + "/hello"))
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  GET /hello -> %d %q\n", resp.StatusCode, string(body))
}

// demoCustomTimeout shows how WithTimeout overrides the default.
func demoCustomTimeout(baseURL string) {
	fmt.Println("=== WithTimeout ===")

	client := chttp.NewClient(chttp.WithTimeout(2 * time.Second))

	resp, _ := client.Do(mustGet(baseURL + "/timeout"))
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  GET /timeout -> %d %q\n", resp.StatusCode, string(body))
}

// demoRoundTripperFn synthesizes a canned response without touching the network.
func demoRoundTripperFn() {
	fmt.Println("=== RoundTripperFn ===")

	canned := chttp.RoundTripperFn(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTeapot,
			Body:       http.NoBody,
			Request:    req,
		}, nil
	})

	client := chttp.NewClient(chttp.WithTransport(canned))

	resp, _ := client.Do(mustGet("http://does-not-resolve/teapot"))
	defer func() { _ = resp.Body.Close() }()

	fmt.Printf("  fake transport -> %d\n", resp.StatusCode)
}

// demoTracingTransport wraps the default transport with a logging adapter
// built from RoundTripperFn — the middleware pattern common/http
// promotes for retry/limiter/auth.
func demoTracingTransport(baseURL string) {
	fmt.Println("=== Tracing middleware ===")

	tracing := chttp.RoundTripperFn(func(req *http.Request) (*http.Response, error) {
		fmt.Printf("  -> %s %s\n", req.Method, req.URL.Path)
		return http.DefaultTransport.RoundTrip(req)
	})

	client := chttp.NewClient(chttp.WithTransport(tracing))

	resp, _ := client.Do(mustGet(baseURL + "/traced"))
	defer func() { _ = resp.Body.Close() }()

	fmt.Printf("  <- %d\n", resp.StatusCode)
}

// mustGet builds a GET request or panics. Demo helper only.
func mustGet(url string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		panic(err)
	}

	return req
}
