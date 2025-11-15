package http

import (
	"errors"
	"net"
	"net/http"

	retry "github.com/avast/retry-go/v4"
)

var (
	_ retry.RetryIfFunc   = NoopRetryIf
	_ retry.OnRetryFunc   = NoopRetryHook
	_ RetryOnResponseFunc = NoopRetryOnResponse

	_ retry.RetryIfFunc   = RetryIfHttpError
	_ RetryOnResponseFunc = RetryOn5xxAnd429Response
)

// Noop

func NoopRetryOnResponse(res *http.Response) bool {
	// no-op: explicitly touch params to generate coverage statements
	_ = res
	return false
}

func NoopRetryIf(err error) bool {
	// no-op: explicitly touch params to generate coverage statements
	_ = err
	return false
}

func NoopRetryHook(n uint, err error) {
	// no-op: explicitly touch params to generate coverage statements
	_ = n
	_ = err
}

// Defaults

type RetryOnResponseFunc func(res *http.Response) bool

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
