package http

import (
	"errors"
	"net/http"
	"testing"
)

func TestNoopRetryIf_AlwaysFalse(t *testing.T) {
	// nil error
	if got := NoopRetryIf(nil); got {
		t.Fatalf("NoopRetryIf(nil) = %v, want false", got)
	}
	// non-nil error
	if got := NoopRetryIf(errors.New("boom")); got {
		t.Fatalf("NoopRetryIf(non-nil) = %v, want false", got)
	}
}

func TestNoopRetryHook_NoOp(t *testing.T) {
	// Should not panic or have any side effects; just invoke with a few values
	NoopRetryHook(0, nil)
	NoopRetryHook(3, errors.New("x"))
}

func TestNoopRetryOnResponse_AlwaysFalse(t *testing.T) {
	if got := NoopRetryOnResponse(nil); got {
		t.Fatalf("NoopRetryOnResponse(nil) = %v, want false", got)
	}
	if got := NoopRetryOnResponse(&http.Response{StatusCode: 200}); got {
		t.Fatalf("NoopRetryOnResponse(res) = %v, want false", got)
	}
}

func TestRetryOn5xxAnd429Response(t *testing.T) {
	cases := []struct {
		name string
		res  *http.Response
		want bool
	}{
		{"nil", nil, false},
		{"200", &http.Response{StatusCode: 200}, false},
		{"400", &http.Response{StatusCode: 400}, false},
		{"499", &http.Response{StatusCode: 499}, false},
		{"429", &http.Response{StatusCode: http.StatusTooManyRequests}, true},
		{"500", &http.Response{StatusCode: 500}, true},
		{"503", &http.Response{StatusCode: 503}, true},
		{"599", &http.Response{StatusCode: 599}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RetryOn5xxAnd429Response(tc.res); got != tc.want {
				t.Fatalf("RetryOn5xxAnd429Response(%v) = %v, want %v", tc.res, got, tc.want)
			}
		})
	}
}

// custom net.Error implementations for testing
type testNetErr struct{ timeout bool }

func (e testNetErr) Error() string { return "net" }

func (e testNetErr) Timeout() bool { return e.timeout }

func (e testNetErr) Temporary() bool { return false }

func TestRetryIfHttpError(t *testing.T) {
	// nil -> false
	if got := RetryIfHttpError(nil); got {
		t.Fatalf("RetryIfHttpError(nil) = %v, want false", got)
	}

	// StatusCodeError -> true
	if got := RetryIfHttpError(&StatusCodeError{StatusCode: 503}); !got {
		t.Fatalf("RetryIfHttpError(StatusCodeError) = %v, want true", got)
	}

	// net.Error timeout true -> true
	if got := RetryIfHttpError(testNetErr{timeout: true}); !got {
		t.Fatalf("RetryIfHttpError(net.Error timeout) = %v, want true", got)
	}

	// net.Error timeout false -> false
	if got := RetryIfHttpError(testNetErr{timeout: false}); got {
		t.Fatalf("RetryIfHttpError(net.Error no-timeout) = %v, want false", got)
	}

	// other error (non net.Error, non StatusCodeError) -> false
	if got := RetryIfHttpError(errors.New("plain")); got {
		t.Fatalf("RetryIfHttpError(plain error) = %v, want false", got)
	}
}

func TestNoopDo_ReturnsErrDoRequestFailed(t *testing.T) {
	// NoopDo ignores the request and always returns ErrDo(ErrHttpRequestFailed)
	res, err := NoopDo(nil)
	if res != nil {
		t.Fatalf("NoopDo returned non-nil response: %+v", res)
	}
	if err == nil || !errors.Is(err, ErrHttpRequestFailed) {
		t.Fatalf("NoopDo error = %v, want wrapping ErrHttpRequestFailed", err)
	}
}

func TestDo_UsesDefaultClientTransport(t *testing.T) {
	// Swap the default client's transport with a success round tripper and ensure Do delegates.
	oldTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	http.DefaultTransport = successRT{body: ""}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	res, err := Do(req) // nolint:gosec
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if res == nil || res.StatusCode != 200 {
		t.Fatalf("unexpected response: %+v", res)
	}
	_ = res.Body.Close()
}
