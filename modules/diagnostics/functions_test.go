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

	"github.com/guidomantilla/yarumo/common/lifecycle"
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

func TestBuildTraceFlightRecorder(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil recorder and closeFn", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		// Do not actually call Start on the underlying flight recorder
		// (only one can be active per process). Calling closeFn straight
		// away signals Stop and waits for the lifecycle goroutine to exit.
		// We exercise the wire-up here, not the runtime semantics.
		recorder, closeFn, err := BuildTraceFlightRecorder(context.Background(), "build-tracefr-1", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if recorder == nil {
			t.Fatal("expected non-nil recorder")
		}

		if closeFn == nil {
			t.Fatal("expected non-nil closeFn")
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("recorder carries the given name", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		recorder, closeFn, err := BuildTraceFlightRecorder(context.Background(), "build-tracefr-named", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(context.Background(), time.Second)

		if recorder.Name() != "build-tracefr-named" {
			t.Fatalf("expected name %q, got %q", "build-tracefr-named", recorder.Name())
		}
	})

	t.Run("returned closeFn drains the background goroutine", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		recorder, closeFn, err := BuildTraceFlightRecorder(context.Background(), "build-tracefr-drain", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)

		select {
		case <-recorder.Done():
		default:
			t.Fatal("expected recorder Done closed after closeFn returned")
		}
	})

	t.Run("matches the BuildTraceFlightRecorderFn signature", func(t *testing.T) {
		t.Parallel()

		var fn BuildTraceFlightRecorderFn = BuildTraceFlightRecorder

		errChan := make(chan error, 1)

		_, closeFn, err := fn(context.Background(), "build-tracefr-fn", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("errChan accepts startup errors without blocking", func(t *testing.T) {
		t.Parallel()

		// Unbuffered channel: a non-blocking send by lifecycle.Start
		// should fall through the default arm. The build itself must
		// still succeed.
		errChan := make(chan error)

		_, closeFn, err := BuildTraceFlightRecorder(context.Background(), "build-tracefr-errchan", lifecycle.ErrChan(errChan))
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})
}

func TestBuildBlockProfiling(t *testing.T) {
	t.Run("returns non-nil sampler and closeFn", func(t *testing.T) {
		blockProfileMu.Lock()
		defer blockProfileMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		errChan := make(chan error, 1)

		sampler, closeFn, err := BuildBlockProfiling(context.Background(), "build-blockprof-1", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if sampler == nil {
			t.Fatal("expected non-nil sampler")
		}

		if closeFn == nil {
			t.Fatal("expected non-nil closeFn")
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("sampler carries the given name and rate", func(t *testing.T) {
		blockProfileMu.Lock()
		defer blockProfileMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		errChan := make(chan error, 1)

		sampler, closeFn, err := BuildBlockProfiling(context.Background(), "build-blockprof-named", errChan, WithBlockProfileRate(250))
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(context.Background(), time.Second)

		if sampler.Name() != "build-blockprof-named" {
			t.Fatalf("expected name %q, got %q", "build-blockprof-named", sampler.Name())
		}

		if sampler.Rate() != 250 {
			t.Fatalf("expected rate %d, got %d", 250, sampler.Rate())
		}
	})

	t.Run("returned closeFn drains the background goroutine", func(t *testing.T) {
		blockProfileMu.Lock()
		defer blockProfileMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		errChan := make(chan error, 1)

		sampler, closeFn, err := BuildBlockProfiling(context.Background(), "build-blockprof-drain", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)

		select {
		case <-sampler.Done():
		default:
			t.Fatal("expected sampler Done closed after closeFn returned")
		}
	})

	t.Run("matches the BuildBlockProfilingFn signature", func(t *testing.T) {
		blockProfileMu.Lock()
		defer blockProfileMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		var fn BuildBlockProfilingFn = BuildBlockProfiling

		errChan := make(chan error, 1)

		_, closeFn, err := fn(context.Background(), "build-blockprof-fn", errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("errChan accepts startup errors without blocking", func(t *testing.T) {
		blockProfileMu.Lock()
		defer blockProfileMu.Unlock()
		defer runtime.SetBlockProfileRate(0)

		errChan := make(chan error)

		_, closeFn, err := BuildBlockProfiling(context.Background(), "build-blockprof-errchan", lifecycle.ErrChan(errChan))
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})
}
