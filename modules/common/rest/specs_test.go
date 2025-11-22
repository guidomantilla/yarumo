package rest

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestRequestSpec_Build_SetsDefaultsAndBody(t *testing.T) {
	spec := &RequestSpec{
		Method:  http.MethodPost,
		URL:     "http://example.com/api",
		Path:    "v1/resource",
		Headers: map[string]string{"X-Trace": "1"},
		QueryParams: map[string][]string{
			"a": {"1", "2"},
		},
		Body: map[string]any{"x": 1},
	}

	req, err := spec.Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept header not set, got %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type header not set, got %q", got)
	}
	if got := req.Header.Get("Content-Length"); got == "" {
		t.Fatalf("Content-Length not set")
	}
	if !strings.Contains(req.URL.String(), "/api/v1/resource") {
		t.Fatalf("path not joined correctly: %s", req.URL.String())
	}
	if !strings.Contains(req.URL.RawQuery, "a=1") || !strings.Contains(req.URL.RawQuery, "a=2") {
		t.Fatalf("query params not encoded: %s", req.URL.RawQuery)
	}
}

func TestRequestSpec_Build_InvalidURL(t *testing.T) {
	spec := &RequestSpec{Method: http.MethodGet, URL: ":bad url"}
	if _, err := spec.Build(context.Background()); err == nil {
		t.Fatalf("expected error for invalid URL")
	}
}

func TestRequestSpec_Build_JSONMarshalError(t *testing.T) {
	ch := make(chan int)
	spec := &RequestSpec{Method: http.MethodPost, URL: "http://example.com", Body: ch}
	if _, err := spec.Build(context.Background()); err == nil {
		t.Fatalf("expected json marshal error")
	}
}

func TestRequestSpec_Build_NewRequestWithContextError(t *testing.T) {
	// Method with space should be rejected by http.NewRequestWithContext
	spec := &RequestSpec{Method: "BAD METHOD", URL: "http://example.com"}
	if _, err := spec.Build(context.Background()); err == nil {
		t.Fatalf("expected error from http.NewRequestWithContext due to invalid method")
	}
}

func TestRequestSpec_Build_NoBodyAndAcceptPreset(t *testing.T) {
	spec := &RequestSpec{
		Method:  http.MethodGet,
		URL:     "http://example.com",
		Headers: map[string]string{"Accept": "application/json"},
	}
	req, err := spec.Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("accept should remain preset, got %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "" {
		t.Fatalf("content-type should be empty when no body, got %q", got)
	}
	if req.URL.Path != "" && req.URL.Path != "/" { // path.Join can normalize to empty or '/'
		t.Fatalf("unexpected path: %q", req.URL.Path)
	}
	if req.URL.RawQuery != "" {
		t.Fatalf("unexpected query: %q", req.URL.RawQuery)
	}
}

func TestRequestSpec_Build_HeadersNilSetsAccept(t *testing.T) {
	spec := &RequestSpec{
		Method: http.MethodGet,
		URL:    "http://example.com",
		// Headers intentionally nil
	}
	req, err := spec.Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("expected default Accept header, got %q", got)
	}
}

func TestRequestSpec_Build_BodyWithPreSetContentTypeNotOverwritten(t *testing.T) {
	spec := &RequestSpec{
		Method:  http.MethodPost,
		URL:     "http://example.com",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    map[string]any{"x": 1},
	}
	req, err := spec.Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got := req.Header.Get("Content-Type"); got != "text/plain" {
		t.Fatalf("content-type should not be overwritten, got %q", got)
	}
	if got := req.Header.Get("Content-Length"); got == "" {
		t.Fatalf("content-length must be set")
	}
}

func TestRequestSpec_Build_HeadersNilWithBody_SetsDefaults(t *testing.T) {
	spec := &RequestSpec{
		Method: http.MethodPost,
		URL:    "http://example.com",
		Body:   map[string]any{"k": "v"},
		// Headers nil triggers creation and default Accept/Content-Type
	}
	req, err := spec.Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if req.Header.Get("Accept") != "application/json" {
		t.Fatalf("missing default Accept header")
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("missing default Content-Type header")
	}
	if req.Header.Get("Content-Length") == "" {
		t.Fatalf("missing Content-Length header")
	}
}

func TestRequestSpec_Build_PathOnlyAddsToURL(t *testing.T) {
	spec := &RequestSpec{
		Method: http.MethodGet,
		URL:    "http://example.com",
		Path:   "v1",
	}
	req, err := spec.Build(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if req.URL.Path != "/v1" {
		t.Fatalf("expected path '/v1', got %q", req.URL.Path)
	}
}

func TestRequestSpec_Build_ContextNil(t *testing.T) {
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	req, err := spec.Build(nil) //nolint:staticcheck
	if err == nil || req != nil {
		t.Fatalf("expected error when context is nil, got req=%v err=%v", req, err)
	}
	if !strings.Contains(err.Error(), "context is nil") {
		t.Fatalf("expected wrapped ErrContextNil, got %v", err)
	}
}
