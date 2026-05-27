package retry

import (
	retrygo "github.com/avast/retry-go/v4"

	cretry "github.com/guidomantilla/yarumo/core/common/resilience/retry"
)

// delayTypeFor maps a Backoff to the corresponding retry-go DelayTypeFunc.
// Unknown values fall back to BackOffDelay (exponential) so a corrupted
// configuration cannot disable the retry loop.
func delayTypeFor(b cretry.Backoff) retrygo.DelayTypeFunc {
	switch b {
	case cretry.BackoffFixed:
		return retrygo.FixedDelay
	case cretry.BackoffRandom:
		return retrygo.RandomDelay
	case cretry.BackoffExponential:
		return retrygo.BackOffDelay
	default:
		return retrygo.BackOffDelay
	}
}
