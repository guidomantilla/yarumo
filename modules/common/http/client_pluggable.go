package http

import (
	"net/http"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// PluggableClient is a Client implementation with pluggable function fields for testing and composition.
type PluggableClient struct {
	DoFn             DoFn
	LimiterEnabledFn LimiterEnabledFn
	RetrierEnabledFn RetrierEnabledFn
}

// Do executes the request by delegating to the configured DoFn.
func (c *PluggableClient) Do(req *http.Request) (*http.Response, error) {
	cassert.NotNil(c, "client is nil")
	cassert.NotNil(c.DoFn, "DoFn is nil")

	if cutils.Nil(req) {
		return nil, ErrDo(ErrHttpRequestNil)
	}

	return c.DoFn(req)
}

// LimiterEnabled reports whether rate limiting is enabled by delegating to LimiterEnabledFn.
func (c *PluggableClient) LimiterEnabled() bool {
	cassert.NotNil(c, "client is nil")

	if c.LimiterEnabledFn == nil {
		return false
	}

	return c.LimiterEnabledFn()
}

// RetrierEnabled reports whether retries are enabled by delegating to RetrierEnabledFn.
func (c *PluggableClient) RetrierEnabled() bool {
	cassert.NotNil(c, "client is nil")

	if c.RetrierEnabledFn == nil {
		return false
	}

	return c.RetrierEnabledFn()
}
