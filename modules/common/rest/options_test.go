package rest

import (
	"net/http"
	"testing"
)

func TestOptions_DefaultsAndOverride(t *testing.T) {
	// default
	opts := NewOptions()
	if opts.DoFn == nil {
		t.Fatalf("default DoFn should not be nil")
	}
	// override
	called := false
	do := func(req *http.Request) (*http.Response, error) {
		called = true
		return makeResp(200, []byte("{}"), map[string]string{"Content-Type": "application/json"}), nil
	}
	opts = NewOptions(WithDoFn(do))
	_, _ = opts.DoFn((&http.Request{})) // nil-safe route not used, but call anyway

	if !called {
		t.Fatalf("WithDoFn did not override DoFn")
	}
}
