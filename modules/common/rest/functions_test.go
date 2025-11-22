package rest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// helper to build a basic http.Response with given status, body and headers
func makeResp(status int, body []byte, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	if body == nil {
		body = []byte{}
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

type errReadCloser struct{}

func (errReadCloser) Read(_ []byte) (int, error) { return 0, errors.New("read-error") }

func (errReadCloser) Close() error { return nil }

// ---- functions.go ----

type sample struct{ X int }

func TestCall_SuccessJSON(t *testing.T) {
	do := func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		body := []byte(`{"X": 7}`)
		return makeResp(200, body, map[string]string{"Content-Type": "application/json"}), nil
	}
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	res, err := Call[sample](context.Background(), spec, WithDoFn(do))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Body.X != 7 {
		t.Fatalf("unexpected decoded body: %+v", res.Body)
	}
	if res.Code != 200 || res.Status != "OK" || res.ContentLength == 0 || res.Duration <= 0 {
		t.Fatalf("unexpected response spec: %+v", res)
	}
}

func TestCall_SupportsByteSliceAndString(t *testing.T) {
	// []byte
	do1 := func(_ *http.Request) (*http.Response, error) {
		return makeResp(200, []byte("abc"), map[string]string{"Content-Type": "application/octet-stream"}), nil
	}
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	res1, err := Call[[]byte](context.Background(), spec, WithDoFn(do1))
	if err != nil || string(res1.Body) != "abc" {
		t.Fatalf("unexpected ([]byte]) result: %v %q", err, string(res1.Body))
	}

	// string
	do2 := func(_ *http.Request) (*http.Response, error) {
		return makeResp(200, []byte("hello"), map[string]string{"Content-Type": "text/plain"}), nil
	}
	res2, err := Call[string](context.Background(), spec, WithDoFn(do2))
	if err != nil || res2.Body != "hello" {
		t.Fatalf("unexpected (string) result: %v %q", err, res2.Body)
	}
}

func TestCall_JSONContentTypeVariants(t *testing.T) {
	do := func(_ *http.Request) (*http.Response, error) {
		return makeResp(200, []byte(`{"X":1}`), map[string]string{"Content-Type": "application/ld+json; charset=utf-8"}), nil
	}
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	res, err := Call[sample](context.Background(), spec, WithDoFn(do))
	if err != nil || res.Body.X != 1 {
		t.Fatalf("unexpected result: %v %+v", err, res)
	}
}

func TestCall_NoContentAndEmptyBody(t *testing.T) {
	// 204 should return zero value without error
	do204 := func(_ *http.Request) (*http.Response, error) {
		return makeResp(204, nil, map[string]string{"Content-Type": "application/json"}), nil
	}
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(do204)); err != nil {
		t.Fatalf("unexpected error for 204: %v", err)
	}

	// empty body
	doEmpty := func(_ *http.Request) (*http.Response, error) {
		return makeResp(200, nil, map[string]string{"Content-Type": "application/json"}), nil
	}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(doEmpty)); err != nil {
		t.Fatalf("unexpected error for empty body: %v", err)
	}
}

func TestCall_DoFnErrorAndReadErrorAndHTTPStatusError(t *testing.T) {
	// DoFn error
	doErr := func(_ *http.Request) (*http.Response, error) { return nil, errors.New("boom") }
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(doErr)); err == nil {
		t.Fatalf("expected wrapped error when DoFn fails")
	}

	// Read body error
	doReadErr := func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Status: "OK", Body: errReadCloser{}}, nil
	}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(doReadErr)); err == nil {
		t.Fatalf("expected error when reading body fails")
	}

	// Non-2xx
	do500 := func(_ *http.Request) (*http.Response, error) {
		return makeResp(500, []byte("oops"), map[string]string{"Content-Type": "text/plain"}), nil
	}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(do500)); err == nil || !strings.Contains(err.Error(), "unexpected status code 500") {
		t.Fatalf("expected HTTPError wrapped, got: %v", err)
	}
}

func TestCall_ContentTypeNotSupported_Error(t *testing.T) {
	do := func(_ *http.Request) (*http.Response, error) {
		return makeResp(200, []byte("<xml></xml>"), map[string]string{"Content-Type": "application/xml"}), nil
	}
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(do)); err == nil {
		t.Fatalf("expected content type not supported error")
	}
}

func TestCall_JSONUnmarshalError(t *testing.T) {
	do := func(_ *http.Request) (*http.Response, error) {
		return makeResp(200, []byte("not-json"), map[string]string{"Content-Type": "application/json"}), nil
	}
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(do)); err == nil {
		t.Fatalf("expected json unmarshal error")
	}
}

func TestCall_JSONWithoutContentType_Decodes(t *testing.T) {
	do := func(_ *http.Request) (*http.Response, error) {
		// No Content-Type header
		return makeResp(200, []byte(`{"X":2}`), map[string]string{}), nil
	}
	spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
	res, err := Call[sample](context.Background(), spec, WithDoFn(do))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Body.X != 2 {
		t.Fatalf("expected decoded value 2, got %v", res.Body.X)
	}
}

func TestCall_ContextNilAndSpecNil(t *testing.T) {
    // ctx nil
    spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}
    if _, err := Call[sample](nil, spec, WithDoFn(func(r *http.Request) (*http.Response, error) {
        t.Fatalf("doFn should not be called when ctx is nil")
        return nil, nil
    })); err == nil {
        t.Fatalf("expected error when context is nil")
    }

    // spec nil
    if _, err := Call[sample](context.Background(), nil, WithDoFn(func(r *http.Request) (*http.Response, error) {
        t.Fatalf("doFn should not be called when spec is nil")
        return nil, nil
    })); err == nil {
        t.Fatalf("expected error when spec is nil")
    }
}

func TestDecodeResponseBody_UnknownStatusOrEmptyBody(t *testing.T) {
    // Unknown status code (no http.StatusText), should return zero value without error
    body := []byte(`{"X":123}`)
    var zero sample
    got, err := decodeResponseBody[sample](body, 0, "application/json")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got != zero {
        t.Fatalf("expected zero value when status text is empty, got: %+v", got)
    }

    // Empty body should also yield zero without error
    got2, err := decodeResponseBody[sample]([]byte{}, 200, "application/json")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got2 != zero {
        t.Fatalf("expected zero value with empty body, got: %+v", got2)
    }
}

func TestCall_SpecBuildErrorIsWrapped(t *testing.T) {
	// Invalid URL on spec should cause Build to fail and Call to wrap error
	spec := &RequestSpec{Method: http.MethodGet, URL: ":bad url"}
	if _, err := Call[sample](context.Background(), spec, WithDoFn(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("doFn should not be called when build fails")
		return nil, nil
	})); err == nil {
		t.Fatalf("expected error from Call when spec.Build fails")
	}
}

func TestIsJSONMediaType(t *testing.T) {
	if !isJSONMediaType("application/json") {
		t.Fatalf("expected application/json to be json")
	}
	if !isJSONMediaType("application/ld+json") {
		t.Fatalf("expected +json suffix to be json")
	}
	if isJSONMediaType("text/plain") {
		t.Fatalf("unexpected positive for text/plain")
	}
}
