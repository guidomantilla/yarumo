package retry

import (
	retrygo "github.com/avast/retry-go/v4"
)

// delayTypeFor maps a Backoff to the corresponding retry-go DelayTypeFunc.
// Unknown values fall back to BackOffDelay (exponential) so a corrupted
// configuration cannot disable the retry loop.
func delayTypeFor(b Backoff) retrygo.DelayTypeFunc {
	switch b {
	case BackoffFixed:
		return retrygo.FixedDelay
	case BackoffRandom:
		return retrygo.RandomDelay
	case BackoffExponential:
		return retrygo.BackOffDelay
	default:
		return retrygo.BackOffDelay
	}
}
