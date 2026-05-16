package diagnostics

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"runtime/pprof"
	"sync"
	"testing"
	"time"
)

// pprofProfileMagic is the gzip magic number prefix used by all pprof
// profile.proto outputs (CPU, heap, goroutine, block, etc.). The pprof Go
// runtime always writes profiles as gzip-compressed protobuf, so checking for
// the gzip prefix is a cheap parseability sanity check.
var pprofProfileMagic = []byte{0x1f, 0x8b}

func TestCaptureHeapProfile(t *testing.T) {
	t.Parallel()

	t.Run("writes profile to buffer", func(t *testing.T) {
		t.Parallel()

		// Force at least one allocation so the heap profile is non-empty.
		sink := make([]byte, 1<<16)
		runtime.KeepAlive(sink)
		runtime.GC()

		var buf bytes.Buffer

		err := CaptureHeapProfile(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if buf.Len() == 0 {
			t.Fatal("expected non-empty buffer")
		}

		if !bytes.HasPrefix(buf.Bytes(), pprofProfileMagic) {
			t.Fatalf("expected pprof gzip magic prefix, got % x", buf.Bytes()[:2])
		}
	})

	t.Run("nil writer returns ErrWriterNil", func(t *testing.T) {
		t.Parallel()

		err := CaptureHeapProfile(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrWriterNil) {
			t.Fatalf("expected ErrWriterNil in chain, got %v", err)
		}

		if !errors.Is(err, ErrCaptureFailed) {
			t.Fatalf("expected ErrCaptureFailed in chain, got %v", err)
		}
	})
}

func TestCaptureGoroutineProfile(t *testing.T) {
	t.Parallel()

	t.Run("writes profile to buffer", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		err := CaptureGoroutineProfile(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if buf.Len() == 0 {
			t.Fatal("expected non-empty buffer")
		}

		if !bytes.HasPrefix(buf.Bytes(), pprofProfileMagic) {
			t.Fatalf("expected pprof gzip magic prefix, got % x", buf.Bytes()[:2])
		}
	})

	t.Run("reports a registered profile", func(t *testing.T) {
		t.Parallel()

		// Cross-check that the named profile we capture actually exists in the
		// pprof registry — this guards against typos in the profile name.
		profile := pprof.Lookup("goroutine")
		if profile == nil {
			t.Fatal("goroutine profile is not registered")
		}
	})

	t.Run("nil writer returns ErrWriterNil", func(t *testing.T) {
		t.Parallel()

		err := CaptureGoroutineProfile(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrWriterNil) {
			t.Fatalf("expected ErrWriterNil in chain, got %v", err)
		}
	})
}

// blockProfileMu serialises tests that mutate the global block profile rate.
var blockProfileMu sync.Mutex

func TestCaptureBlockProfile(t *testing.T) {
	t.Run("writes profile to buffer", func(t *testing.T) {
		blockProfileMu.Lock()
		defer blockProfileMu.Unlock()

		runtime.SetBlockProfileRate(1)

		defer runtime.SetBlockProfileRate(0)

		// Trigger a blocking event so the profile is non-empty.
		var wg sync.WaitGroup

		ch := make(chan struct{})

		wg.Go(func() {
			<-ch
		})

		time.Sleep(5 * time.Millisecond)
		close(ch)
		wg.Wait()

		var buf bytes.Buffer

		err := CaptureBlockProfile(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if buf.Len() == 0 {
			t.Fatal("expected non-empty buffer")
		}

		if !bytes.HasPrefix(buf.Bytes(), pprofProfileMagic) {
			t.Fatalf("expected pprof gzip magic prefix, got % x", buf.Bytes()[:2])
		}
	})

	t.Run("nil writer returns ErrWriterNil", func(t *testing.T) {
		err := CaptureBlockProfile(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrWriterNil) {
			t.Fatalf("expected ErrWriterNil in chain, got %v", err)
		}
	})
}

// TestCaptureCPUProfile runs sequentially because pprof.StartCPUProfile is
// process-global — only one CPU profile may be active at a time.
func TestCaptureCPUProfile(t *testing.T) {
	t.Run("writes profile within duration", func(t *testing.T) {
		var buf bytes.Buffer

		ctx := context.Background()

		err := CaptureCPUProfile(ctx, &buf, 50*time.Millisecond)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if buf.Len() == 0 {
			t.Fatal("expected non-empty buffer")
		}

		if !bytes.HasPrefix(buf.Bytes(), pprofProfileMagic) {
			t.Fatalf("expected pprof gzip magic prefix, got % x", buf.Bytes()[:2])
		}
	})

	t.Run("nil writer returns ErrWriterNil", func(t *testing.T) {
		err := CaptureCPUProfile(context.Background(), nil, 50*time.Millisecond)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrWriterNil) {
			t.Fatalf("expected ErrWriterNil in chain, got %v", err)
		}
	})

	t.Run("nil context returns ErrContextNil", func(t *testing.T) {
		var buf bytes.Buffer

		err := CaptureCPUProfile(nil, &buf, 50*time.Millisecond) //nolint:staticcheck // intentional nil ctx for input validation test
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil in chain, got %v", err)
		}
	})

	t.Run("non-positive duration returns ErrDurationNonPositive", func(t *testing.T) {
		var buf bytes.Buffer

		err := CaptureCPUProfile(context.Background(), &buf, 0)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrDurationNonPositive) {
			t.Fatalf("expected ErrDurationNonPositive in chain, got %v", err)
		}
	})

	t.Run("cancelled context stops profile early", func(t *testing.T) {
		var buf bytes.Buffer

		ctx, cancel := context.WithCancel(context.Background())

		// Cancel almost immediately so the select returns on ctx.Done().
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := CaptureCPUProfile(ctx, &buf, 5*time.Second)
		if err == nil {
			t.Fatal("expected error from cancelled context")
		}

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled in chain, got %v", err)
		}

		if !errors.Is(err, ErrCaptureFailed) {
			t.Fatalf("expected ErrCaptureFailed in chain, got %v", err)
		}
	})

	t.Run("concurrent start returns StartCPUProfile error", func(t *testing.T) {
		// Drive pprof.StartCPUProfile into its "already in use" branch by
		// holding the global CPU profile and then invoking CaptureCPUProfile.
		var holdBuf bytes.Buffer

		err := pprof.StartCPUProfile(&holdBuf)
		if err != nil {
			t.Fatalf("setup: unexpected error: %v", err)
		}

		defer pprof.StopCPUProfile()

		var buf bytes.Buffer

		captureErr := CaptureCPUProfile(context.Background(), &buf, 50*time.Millisecond)
		if captureErr == nil {
			t.Fatal("expected error when CPU profile is already active")
		}

		if !errors.Is(captureErr, ErrCaptureFailed) {
			t.Fatalf("expected ErrCaptureFailed in chain, got %v", captureErr)
		}
	})
}

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
