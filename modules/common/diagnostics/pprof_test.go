package diagnostics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPprofHandler(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil handler", func(t *testing.T) {
		t.Parallel()

		h := NewPprofHandler()
		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})

	t.Run("serves pprof index", func(t *testing.T) {
		t.Parallel()

		h := NewPprofHandler()
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("serves cmdline", func(t *testing.T) {
		t.Parallel()

		h := NewPprofHandler()
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/cmdline", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("serves symbol", func(t *testing.T) {
		t.Parallel()

		h := NewPprofHandler()
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/symbol", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
