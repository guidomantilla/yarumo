package retry

// AlwaysRetry is the default RetryIfFn: retry on every non-nil error. It
// is the most permissive predicate; callers that want to short-circuit on
// permanent errors should provide their own RetryIfFn via WithRetryIf on
// the concrete Retry implementation.
func AlwaysRetry(_ error) bool {
	return true
}

// NoopOnRetry is the default OnRetryFn: no-op. Callers that want to
// observe each retry attempt provide their own hook via WithOnRetry on
// the concrete Retry implementation.
func NoopOnRetry(_ uint, _ error) {}
