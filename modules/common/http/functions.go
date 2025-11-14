package http

func NoopRetryIf(_ error) bool {
	return false
}

func NoopRetryHook(_ uint, _ error) {

}
