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

type sample struct{ X int }

func TestCall(t *testing.T) {
	t.Parallel()

	t.Run("success with JSON response", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte(`{"X": 7}`), map[string]string{"Content-Type": applicationJSON}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		res, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Body.X != 7 {
			t.Fatalf("unexpected decoded body: %+v", res.Body)
		}

		if res.Code != 200 {
			t.Fatalf("unexpected status code: %d", res.Code)
		}

		if res.Status != "OK" {
			t.Fatalf("unexpected status: %s", res.Status)
		}

		if res.ContentLength == 0 {
			t.Fatal("expected non-zero content length")
		}

		if res.Duration <= 0 {
			t.Fatal("expected positive duration")
		}
	})

	t.Run("decodes byte slice body", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte("abc"), map[string]string{"Content-Type": "application/octet-stream"}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		res, err := Call[[]byte](context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if string(res.Body) != "abc" {
			t.Fatalf("unexpected body: %q", string(res.Body))
		}
	})

	t.Run("decodes string body", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte("hello"), map[string]string{"Content-Type": "text/plain"}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		res, err := Call[string](context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Body != "hello" {
			t.Fatalf("unexpected body: %q", res.Body)
		}
	})

	t.Run("decodes JSON content-type variants", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte(`{"X":1}`), map[string]string{"Content-Type": "application/ld+json; charset=utf-8"}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		res, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Body.X != 1 {
			t.Fatalf("unexpected body: %+v", res.Body)
		}
	})

	t.Run("returns zero value for 204 No Content", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(204, nil, map[string]string{"Content-Type": applicationJSON}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns zero value for empty body", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, nil, map[string]string{"Content-Type": applicationJSON}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error when context is nil", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](nil, spec) //nolint:staticcheck // testing nil context
		if err == nil {
			t.Fatal("expected error when context is nil")
		}
	})

	t.Run("returns error when spec is nil", func(t *testing.T) {
		t.Parallel()

		_, err := Call[sample](context.Background(), nil)
		if err == nil {
			t.Fatal("expected error when spec is nil")
		}
	})

	t.Run("returns error when DoFn fails", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected error when DoFn fails")
		}
	})

	t.Run("returns error when body read fails", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "OK",
				Header:     http.Header{},
				Body:       errReadCloser{},
			}, nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected error when body read fails")
		}
	})

	t.Run("returns HTTPError for non-2xx status", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(500, []byte("oops"), map[string]string{"Content-Type": "text/plain"}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected error for non-2xx status")
		}

		if !strings.Contains(err.Error(), "unexpected status code 500") {
			t.Fatalf("expected HTTPError, got: %v", err)
		}
	})

	t.Run("returns error for unsupported content-type", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte("<data/>"), map[string]string{"Content-Type": "application/octet-stream"}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected error for unsupported content-type")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte("not-json"), map[string]string{"Content-Type": applicationJSON}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected JSON unmarshal error")
		}
	})

	t.Run("decodes JSON when content-type is empty", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte(`{"X":2}`), map[string]string{}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		res, err := Call[sample](context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Body.X != 2 {
			t.Fatalf("expected X=2, got %d", res.Body.X)
		}
	})

	t.Run("wraps spec build error", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{Method: http.MethodGet, URL: ":bad url"}

		_, err := Call[sample](context.Background(), spec)
		if err == nil {
			t.Fatal("expected error from Call when spec.Build fails")
		}
	})

	t.Run("returns error when response exceeds max size", func(t *testing.T) {
		t.Parallel()

		bigBody := bytes.Repeat([]byte("x"), 100)
		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, bigBody, map[string]string{"Content-Type": applicationJSON}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := Call[sample](context.Background(), spec, WithDoFn(do), WithMaxResponseSize(50))
		if err == nil {
			t.Fatal("expected error when response exceeds max size")
		}

		if !errors.Is(err, ErrResponseTooLarge) {
			t.Fatalf("expected ErrResponseTooLarge, got: %v", err)
		}
	})

	t.Run("succeeds when response equals max size", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"X":1}`)
		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, body, map[string]string{"Content-Type": applicationJSON}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		res, err := Call[sample](context.Background(), spec, WithDoFn(do), WithMaxResponseSize(int64(len(body))))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Body.X != 1 {
			t.Fatalf("unexpected body: %+v", res.Body)
		}
	})
}

func TestIsJSONMediaType(t *testing.T) {
	t.Parallel()

	t.Run("matches application/json", func(t *testing.T) {
		t.Parallel()

		if !isJSONMediaType(applicationJSON) {
			t.Fatal("expected application/json to be recognized as JSON")
		}
	})

	t.Run("matches +json suffix", func(t *testing.T) {
		t.Parallel()

		if !isJSONMediaType("application/ld+json") {
			t.Fatal("expected +json suffix to be recognized as JSON")
		}
	})

	t.Run("rejects non-json type", func(t *testing.T) {
		t.Parallel()

		if isJSONMediaType("text/plain") {
			t.Fatal("unexpected positive for text/plain")
		}
	})
}

func TestParseMediaType(t *testing.T) {
	t.Parallel()

	t.Run("returns empty for empty input", func(t *testing.T) {
		t.Parallel()

		got := parseMediaType("")
		if got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})

	t.Run("extracts media type from content-type with params", func(t *testing.T) {
		t.Parallel()

		got := parseMediaType("application/json; charset=utf-8")
		if got != applicationJSON {
			t.Fatalf("expected %s, got %q", applicationJSON, got)
		}
	})

	t.Run("returns trimmed input on parse error", func(t *testing.T) {
		t.Parallel()

		got := parseMediaType("; invalid")
		if got != "; invalid" {
			t.Fatalf("expected raw media type on parse error, got %q", got)
		}
	})
}

func TestDecodeResponseBody(t *testing.T) {
	t.Parallel()

	t.Run("returns zero for unknown status code", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"X":123}`)

		var zero sample

		got, err := decodeResponseBody[sample](body, 0, applicationJSON)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != zero {
			t.Fatalf("expected zero value, got: %+v", got)
		}
	})

	t.Run("returns zero for empty body", func(t *testing.T) {
		t.Parallel()

		var zero sample

		got, err := decodeResponseBody[sample]([]byte{}, 200, applicationJSON)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != zero {
			t.Fatalf("expected zero value, got: %+v", got)
		}
	})
}

func TestDecodeHTTPError(t *testing.T) {
	t.Parallel()

	type errBody struct {
		Code string `json:"code"`
	}

	t.Run("decodes JSON body from HTTPError", func(t *testing.T) {
		t.Parallel()

		httpErr := &HTTPError{StatusCode: 400, Status: "Bad Request", Body: []byte(`{"code":"invalid"}`)}
		err := ErrCall(httpErr)

		decoded, ok := DecodeHTTPError[errBody](err)
		if !ok {
			t.Fatal("expected DecodeHTTPError to succeed")
		}

		if decoded.Code != "invalid" {
			t.Fatalf("expected Code=invalid, got %q", decoded.Code)
		}
	})

	t.Run("returns false for non-HTTPError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("plain error")

		_, ok := DecodeHTTPError[errBody](err)
		if ok {
			t.Fatal("expected DecodeHTTPError to return false for non-HTTPError")
		}
	})

	t.Run("returns false for empty body", func(t *testing.T) {
		t.Parallel()

		httpErr := &HTTPError{StatusCode: 500, Status: "Internal Server Error", Body: nil}
		err := ErrCall(httpErr)

		_, ok := DecodeHTTPError[errBody](err)
		if ok {
			t.Fatal("expected DecodeHTTPError to return false for empty body")
		}
	})

	t.Run("returns false for invalid JSON body", func(t *testing.T) {
		t.Parallel()

		httpErr := &HTTPError{StatusCode: 422, Status: "Unprocessable Entity", Body: []byte("not-json")}
		err := ErrCall(httpErr)

		_, ok := DecodeHTTPError[errBody](err)
		if ok {
			t.Fatal("expected DecodeHTTPError to return false for invalid JSON")
		}
	})
}

func TestCallStream(t *testing.T) {
	t.Parallel()

	t.Run("returns streaming body on success", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte("streaming-data"), map[string]string{"Content-Type": "text/event-stream"}), nil
		}
		spec := &RequestSpec{Method: http.MethodPost, URL: "http://example.com"}

		res, err := CallStream(context.Background(), spec, WithDoFn(do))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		defer func() { _ = res.Body.Close() }()

		data, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			t.Fatalf("unexpected read error: %v", readErr)
		}

		if string(data) != "streaming-data" {
			t.Fatalf("unexpected body: %q", string(data))
		}

		if res.Code != 200 {
			t.Fatalf("unexpected status code: %d", res.Code)
		}

		if res.Status != "OK" {
			t.Fatalf("unexpected status: %s", res.Status)
		}
	})

	t.Run("returns HTTPError for non-2xx status", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(500, []byte("server error"), map[string]string{}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := CallStream(context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected error for non-2xx status")
		}

		if !strings.Contains(err.Error(), "unexpected status code 500") {
			t.Fatalf("expected HTTPError, got: %v", err)
		}
	})

	t.Run("returns error when context is nil", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := CallStream(nil, spec) //nolint:staticcheck // testing nil context
		if err == nil {
			t.Fatal("expected error when context is nil")
		}
	})

	t.Run("returns error when spec is nil", func(t *testing.T) {
		t.Parallel()

		_, err := CallStream(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error when spec is nil")
		}
	})

	t.Run("returns error when DoFn fails", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := CallStream(context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected error when DoFn fails")
		}
	})

	t.Run("closes body when DoFn returns both response and error", func(t *testing.T) {
		t.Parallel()

		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(200, []byte("partial"), map[string]string{}), errors.New("redirect error")
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := CallStream(context.Background(), spec, WithDoFn(do))
		if err == nil {
			t.Fatal("expected error when DoFn returns error")
		}
	})

	t.Run("wraps spec build error", func(t *testing.T) {
		t.Parallel()

		spec := &RequestSpec{Method: http.MethodGet, URL: ":bad url"}

		_, err := CallStream(context.Background(), spec)
		if err == nil {
			t.Fatal("expected error from CallStream when spec.Build fails")
		}
	})

	t.Run("returns error when error response exceeds max size", func(t *testing.T) {
		t.Parallel()

		bigBody := bytes.Repeat([]byte("x"), 100)
		do := func(_ *http.Request) (*http.Response, error) {
			return makeResp(500, bigBody, map[string]string{}), nil
		}
		spec := &RequestSpec{Method: http.MethodGet, URL: "http://example.com"}

		_, err := CallStream(context.Background(), spec, WithDoFn(do), WithMaxResponseSize(50))
		if err == nil {
			t.Fatal("expected error when error response exceeds max size")
		}

		if !errors.Is(err, ErrResponseTooLarge) {
			t.Fatalf("expected ErrResponseTooLarge, got: %v", err)
		}
	})
}
