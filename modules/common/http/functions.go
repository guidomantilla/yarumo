package http

import retry "github.com/avast/retry-go/v4"

var (
	_ retry.RetryIfFunc = NoopRetryIf
	_ retry.OnRetryFunc = NoopRetryHook
)

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
