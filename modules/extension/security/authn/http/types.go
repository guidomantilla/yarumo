// Copyright 2026 Guido Mauricio Mantilla Tarazona
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package http provides a server-side net/http middleware that
// terminates Bearer authentication.
//
// The middleware exposes the canonical `func(http.Handler) http.Handler`
// shape so it composes with any net/http router or middleware chain
// (stdlib http.ServeMux, gorilla/mux, chi, ...). It reads the bearer
// token from the request's "Authorization" header, delegates
// verification to an authn.Authenticator, and on success injects the
// resulting *Principal into the request ctx via authn.WithPrincipal.
//
// Failure modes:
//   - missing or malformed Authorization header → 401 Unauthorized.
//   - Authenticator.Validate returns an error → 401 Unauthorized.
//   - Authenticator returns a nil *Principal with no error → 401
//     Unauthorized.
//
// The middleware never writes a response body; it sets the status code
// and lets the caller-provided ErrorHandler (default: empty body) shape
// the payload.
package http

import (
	"net/http"

	"github.com/guidomantilla/yarumo/core/security/authn"
)

var (
	_ NewMiddlewareFn = NewMiddleware
)

// NewMiddlewareFn is the function type for NewMiddleware.
type NewMiddlewareFn func(authenticator authn.Authenticator, options ...Option) Middleware

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
