package http

import "github.com/avast/retry-go/v4"

var (
	_ retry.RetryIfFunc = NoopRetryIf
	_ retry.OnRetryFunc = NoopRetryHook
)

func NoopRetryIf(_ error) bool {
	return false
}

func NoopRetryHook(_ uint, _ error) {

}
