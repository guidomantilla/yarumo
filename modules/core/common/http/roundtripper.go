package http

import "net/http"

// RoundTripperFn adapts a plain function to http.RoundTripper. It is the
// client-side analogue of http.HandlerFunc on the server side — surprisingly,
// stdlib does not ship this adapter. Use it to assemble ad-hoc transports in
// tests (canned responses, request inspection) and for one-off middleware
// that doesn't warrant a dedicated struct.
//
//	client := &http.Client{
//	    Transport: chttp.RoundTripperFn(func(req *http.Request) (*http.Response, error) {
//	        return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
//	    }),
//	}
type RoundTripperFn func(req *http.Request) (*http.Response, error)

// RoundTrip calls f(req).
func (f RoundTripperFn) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
