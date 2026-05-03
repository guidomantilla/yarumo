package http

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
)

// EnabledLimiter returns true, indicating rate limiting is active.
func EnabledLimiter() bool {
	return true
}

// DisabledLimiter returns false, indicating rate limiting is inactive.
func DisabledLimiter() bool {
	return false
}

// EnabledRetrier returns true, indicating retries are active.
func EnabledRetrier() bool {
	return true
}

// DisabledRetrier returns false, indicating retries are inactive.
func DisabledRetrier() bool {
	return false
}

// NoopRetryOnResponse always returns false, disabling response-based retries.
func NoopRetryOnResponse(res *http.Response) bool {
	_ = res
	return false
}

// NoopRetryIf always returns false, disabling error-based retries.
func NoopRetryIf(err error) bool {
	_ = err
	return false
}

// NoopRetryHook is a no-op retry hook.
func NoopRetryHook(n uint, err error) {
	_ = n
	_ = err
}

// NoopDo ignores the request and returns a 204 No Content response.
func NoopDo(req *http.Request) (*http.Response, error) {
	_ = req
	res := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Header:     make(http.Header),
	}

	return res, nil
}

// ErrorDo ignores the request and always returns an error.
func ErrorDo(req *http.Request) (*http.Response, error) {
	_ = req
	return nil, ErrDo()
}

// RetryOn5xxAnd429Response returns true for 429 (Too Many Requests) and 5xx status codes.
func RetryOn5xxAnd429Response(res *http.Response) bool {
	if res == nil {
		return false
	}

	if res.StatusCode == http.StatusTooManyRequests {
		return true
	}

	if res.StatusCode >= 500 && res.StatusCode <= 599 {
		return true
	}

	return false
}

// RetryIfHttpError returns true for *StatusCodeError or net.Error with Timeout().
func RetryIfHttpError(err error) bool {
	if err == nil {
		return false
	}

	var scErr *StatusCodeError
	if errors.As(err, &scErr) {
		return true
	}

	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}

	return false
}

// Do delegates to the DefaultClient.
func Do(req *http.Request) (*http.Response, error) {
	return DefaultClient.Do(req)
}
