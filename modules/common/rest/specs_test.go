package rest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRequestSpec_Build(t *testing.T) {
	t.Parallel()

	t.Run("sets default Accept header", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodGet,
			URL:    "http://example.com",
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Accept"); got != applicationJSON {
			t.Fatalf("expected default Accept header, got %q", got)
		}
	})

	t.Run("preserves preset Accept header", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method:  http.MethodGet,
			URL:     "http://example.com",
			Headers: map[string]string{"Accept": "text/html"},
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Accept"); got != "text/html" {
			t.Fatalf("expected preserved Accept header, got %q", got)
		}
	})

	t.Run("initializes nil headers with defaults", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodGet,
			URL:    "http://example.com",
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Accept"); got != applicationJSON {
			t.Fatalf("expected Accept header, got %q", got)
		}
	})

	t.Run("marshals body and sets Content-Type and Content-Length", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodPost,
			URL:    "http://example.com",
			Body:   map[string]any{"x": 1},
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Content-Type"); got != applicationJSON {
			t.Fatalf("expected Content-Type header, got %q", got)
		}

		if got := req.Header.Get("Content-Length"); got == "" {
			t.Fatal("expected Content-Length header to be set")
		}
	})

	t.Run("preserves preset Content-Type", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method:  http.MethodPost,
			URL:     "http://example.com",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    map[string]any{"x": 1},
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Content-Type"); got != "text/plain" {
			t.Fatalf("content-type should not be overwritten, got %q", got)
		}
	})

	t.Run("does not set Content-Type when no body", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodGet,
			URL:    "http://example.com",
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Content-Type"); got != "" {
			t.Fatalf("content-type should be empty when no body, got %q", got)
		}
	})

	t.Run("joins path to URL", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodGet,
			URL:    "http://example.com/api",
			Path:   "v1/resource",
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(req.URL.String(), "/api/v1/resource") {
			t.Fatalf("path not joined correctly: %s", req.URL.String())
		}
	})

	t.Run("encodes query params", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodGet,
			URL:    "http://example.com",
			QueryParams: map[string][]string{
				"a": {"1", "2"},
			},
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(req.URL.RawQuery, "a=1") || !strings.Contains(req.URL.RawQuery, "a=2") {
			t.Fatalf("query params not encoded: %s", req.URL.RawQuery)
		}
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{Method: http.MethodGet, URL: ":bad url"}

		_, err := spec.Build(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})

	t.Run("returns error for unmarshalable body", func(t *testing.T) {
		t.Parallel()

		ch := make(chan int)
		spec := &RequestSpec{Method: http.MethodPost, URL: "http://example.com", Body: ch}

		_, err := spec.Build(context.Background())
		if err == nil {
			t.Fatal("expected json marshal error")
		}
	})

	t.Run("returns error for invalid method", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{Method: "BAD METHOD", URL: "http://example.com"}

		_, err := spec.Build(context.Background())
		if err == nil {
			t.Fatal("expected error from http.NewRequestWithContext due to invalid method")
		}
	})

	t.Run("returns error when context is nil", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := spec.Build(nil) //nolint:staticcheck // testing nil context
		if err == nil {
			t.Fatal("expected error when context is nil")
		}

		if !strings.Contains(err.Error(), "context is nil") {
			t.Fatalf("expected wrapped ErrContextNil, got %v", err)
		}
	})

	t.Run("does not mutate spec headers", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodGet,
			URL:    "http://example.com",
		}

		_, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if spec.Headers != nil {
			t.Fatal("expected spec.Headers to remain nil after Build")
		}
	})

	t.Run("does not mutate spec RawBody", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodPost,
			URL:    "http://example.com",
			Body:   map[string]any{"k": "v"},
		}

		_, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if spec.RawBody != nil {
			t.Fatal("expected spec.RawBody to remain nil after Build")
		}
	})

	t.Run("marshals io.Reader body without default Content-Type", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodPost,
			URL:    "http://example.com",
			Body:   strings.NewReader("raw-data"),
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Content-Type"); got != "" {
			t.Fatalf("expected empty Content-Type for io.Reader, got %q", got)
		}
	})

	t.Run("marshals byte slice body with Content-Length", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodPost,
			URL:    "http://example.com",
			Body:   []byte("raw-bytes"),
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Content-Length"); got == "" {
			t.Fatal("expected Content-Length header for byte slice body")
		}

		if got := req.Header.Get("Content-Type"); got != "" {
			t.Fatalf("expected empty Content-Type for []byte, got %q", got)
		}
	})

	t.Run("marshals string body with Content-Length", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodPost,
			URL:    "http://example.com",
			Body:   "string-body",
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Content-Length"); got == "" {
			t.Fatal("expected Content-Length header for string body")
		}

		if got := req.Header.Get("Content-Type"); got != "" {
			t.Fatalf("expected empty Content-Type for string, got %q", got)
		}
	})

	t.Run("marshals url.Values body with form content-type", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{
			Method: http.MethodPost,
			URL:    "http://example.com",
			Body:   url.Values{"key": {"val"}},
		}

		req, err := spec.Build(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := req.Header.Get("Content-Type"); got != applicationFormURLEncoded {
			t.Fatalf("expected %s, got %q", applicationFormURLEncoded, got)
		}

		if got := req.Header.Get("Content-Length"); got == "" {
			t.Fatal("expected Content-Length header for url.Values body")
		}
	})
}
