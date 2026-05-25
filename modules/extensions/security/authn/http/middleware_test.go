package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guidomantilla/yarumo/security/authn"
	authnhttp "github.com/guidomantilla/yarumo/extensions/security/authn/http"
)

// fakeAuthenticator is a test double for authn.Authenticator. The
// pluggable function lets each subtest swap behavior without spinning
// up a real JWT method.
type fakeAuthenticator struct {
	validateFn func(ctx context.Context, token string) (*authn.Principal, error)
}

func (a *fakeAuthenticator) Validate(ctx context.Context, token string) (*authn.Principal, error) {
	return a.validateFn(ctx, token)
}

func newOKAuthenticator(p *authn.Principal) authn.Authenticator {
	return &fakeAuthenticator{
		validateFn: func(_ context.Context, _ string) (*authn.Principal, error) {
			return p, nil
		},
	}
}

func newRejectingAuthenticator(err error) authn.Authenticator {
	return &fakeAuthenticator{
		validateFn: func(_ context.Context, _ string) (*authn.Principal, error) {
			return nil, err
		},
	}
}

func TestNewMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("happy path injects principal", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u-1", Name: "Alice"}
		mw := authnhttp.NewMiddleware(newOKAuthenticator(want))

		var got *authn.Principal

		var saw bool

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			got, saw = authn.FromContext(r.Context())
		}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer t-1")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}

		if !saw {
			t.Fatal("FromContext reported no principal")
		}

		if got != want {
			t.Fatalf("FromContext returned %v, want %v", got, want)
		}
	})

	t.Run("missing header returns 401", func(t *testing.T) {
		t.Parallel()

		mw := authnhttp.NewMiddleware(newOKAuthenticator(&authn.Principal{ID: "x"}))

		reached := false

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			reached = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}

		if reached {
			t.Fatal("inner handler was reached despite missing header")
		}
	})

	t.Run("malformed header returns 401", func(t *testing.T) {
		t.Parallel()

		mw := authnhttp.NewMiddleware(newOKAuthenticator(&authn.Principal{ID: "x"}))

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("scheme without token returns 401", func(t *testing.T) {
		t.Parallel()

		mw := authnhttp.NewMiddleware(newOKAuthenticator(&authn.Principal{ID: "x"}))

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("empty bearer token returns 401", func(t *testing.T) {
		t.Parallel()

		mw := authnhttp.NewMiddleware(newOKAuthenticator(&authn.Principal{ID: "x"}))

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer    ")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("scheme is case insensitive", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u-1"}
		mw := authnhttp.NewMiddleware(newOKAuthenticator(want))

		var saw bool

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			_, saw = authn.FromContext(r.Context())
		}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "bearer t-1")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !saw {
			t.Fatalf("status %d, principal not seen with lower-case scheme", rec.Code)
		}
	})

	t.Run("authenticator error returns 401", func(t *testing.T) {
		t.Parallel()

		mw := authnhttp.NewMiddleware(newRejectingAuthenticator(authn.ErrAuthentication(authn.ErrTokenInvalid)))

		reached := false

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			reached = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer bad")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}

		if reached {
			t.Fatal("inner handler reached after authenticator rejection")
		}
	})

	t.Run("nil principal with no error returns 401", func(t *testing.T) {
		t.Parallel()

		mw := authnhttp.NewMiddleware(newOKAuthenticator(nil))

		reached := false

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			reached = true
		}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer t")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}

		if reached {
			t.Fatal("inner handler reached on nil principal")
		}
	})

	t.Run("custom header name", func(t *testing.T) {
		t.Parallel()

		want := &authn.Principal{ID: "u-1"}
		mw := authnhttp.NewMiddleware(newOKAuthenticator(want),
			authnhttp.WithHeaderName("X-Custom-Auth"),
		)

		var saw bool

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			_, saw = authn.FromContext(r.Context())
		}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("X-Custom-Auth", "Bearer t-1")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !saw {
			t.Fatalf("status %d, principal not propagated via custom header", rec.Code)
		}
	})

	t.Run("custom error handler observes cause", func(t *testing.T) {
		t.Parallel()

		var captured error

		mw := authnhttp.NewMiddleware(newRejectingAuthenticator(authn.ErrAuthentication(authn.ErrTokenInvalid)),
			authnhttp.WithErrorHandler(func(w http.ResponseWriter, _ *http.Request, cause error) {
				captured = cause

				w.WriteHeader(http.StatusForbidden)
			}),
		)

		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "Bearer bad")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403 from custom handler", rec.Code)
		}

		if captured == nil {
			t.Fatal("custom error handler did not capture cause")
		}

		if !errors.Is(captured, authn.ErrTokenInvalid) {
			t.Fatalf("captured = %v, want ErrTokenInvalid chain", captured)
		}
	})
}
