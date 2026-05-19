package http

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

// spanRequest is the minimal projection of an *http.Request passed to the
// tracing span-name callback. Decoupled from net/http so the callback
// signature stays stable across stdlib changes.
type spanRequest struct {
	method string
	host   string
	path   string
}

// defaultSpanName returns "HTTP <method>" as the span name when no custom
// builder is configured.
func defaultSpanName(r *spanRequest) string {
	return "HTTP " + r.method
}

// canonicalHeader returns name in canonical MIME header form, suitable for
// case-insensitive lookup against http.Request.Header.
func canonicalHeader(name string) string {
	return http.CanonicalHeaderKey(name)
}

// hostOrEmpty returns req.URL.Host when URL is non-nil, the empty string
// otherwise. Used by the tracing decorator to record the http.host
// attribute without panicking on URL-less requests (rare but possible in
// tests).
func hostOrEmpty(req *http.Request) string {
	if req.URL == nil {
		return ""
	}
	return req.URL.Host
}

// pathOrEmpty mirrors hostOrEmpty for req.URL.Path.
func pathOrEmpty(req *http.Request) string {
	if req.URL == nil {
		return ""
	}
	return req.URL.Path
}

// collectHeaderAttributes walks every header in h and emits one
// attribute.KeyValue per name. Values are masked with "<redacted>" when
// the header's canonical form appears in the redaction set. Multiple
// values for the same header are joined with comma.
func collectHeaderAttributes(h http.Header, redacted map[string]struct{}) []attribute.KeyValue {
	if len(h) == 0 {
		return nil
	}
	out := make([]attribute.KeyValue, 0, len(h))
	for name, values := range h {
		key := "http.request.header." + name
		if _, ok := redacted[name]; ok {
			out = append(out, attribute.String(key, "<redacted>"))
			continue
		}
		out = append(out, attribute.StringSlice(key, values))
	}
	return out
}
