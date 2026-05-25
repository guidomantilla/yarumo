package http

import (
	"net/http"
)

// Middleware is the canonical net/http server middleware shape: a
// function that wraps an http.Handler and returns a new http.Handler.
// Exporting the alias documents the intended composition pattern and
// keeps callers from having to spell out the function type every time.
type Middleware func(next http.Handler) http.Handler

// ErrorHandler is the function type invoked when authentication fails.
// Implementations write the desired response (status code, body,
// headers) for unauthenticated requests. cause carries the wrapped
// authn error so consumers can log the precise reason.
//
// The default ErrorHandler installed by NewMiddleware writes 401
// Unauthorized with an empty body. Custom handlers may emit a JSON
// problem document, redirect to a login page, etc.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, cause error)
