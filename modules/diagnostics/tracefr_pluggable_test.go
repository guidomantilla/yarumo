package diagnostics

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestPluggableTraceFlightRecorder_Name(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns empty", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}
		if p.Name() != "" {
			t.Fatalf("expected empty string, got %q", p.Name())
		}
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{
			NameFn: func() string { return "configured-name" },
		}
		if p.Name() != "configured-name" {
			t.Fatalf("expected %q, got %q", "configured-name", p.Name())
		}
	})
}

func TestPluggableTraceFlightRecorder_Start(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns nil", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}

		err := p.Start(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		called := false
		p := &PluggableTraceFlightRecorder{
			StartFn: func(_ context.Context) error {
				called = true

				return nil
			},
		}

		err := p.Start(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !called {
			t.Fatal("expected StartFn to be called")
		}
	})
}

func TestPluggableTraceFlightRecorder_Stop(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns nil and closes Done", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}

		// Initialise Done lazily.
		_ = p.Done()

		err := p.Stop(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		select {
		case <-p.Done():
		default:
			t.Fatal("expected Done channel closed after Stop")
		}
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		called := false
		p := &PluggableTraceFlightRecorder{
			StopFn: func(_ context.Context) error {
				called = true

				return nil
			},
		}

		err := p.Stop(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !called {
			t.Fatal("expected StopFn to be called")
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}
		_ = p.Done()

		err := p.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		err = p.Stop(context.Background())
		if err != nil {
			t.Fatalf("second Stop returned %v", err)
		}
	})
}

func TestPluggableTraceFlightRecorder_Done(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns internal channel", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}
		if p.Done() == nil {
			t.Fatal("expected non-nil channel")
		}
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		ch := make(chan struct{})
		p := &PluggableTraceFlightRecorder{
			DoneFn: func() <-chan struct{} { return ch },
		}

		if p.Done() != (<-chan struct{})(ch) {
			t.Fatal("expected DoneFn channel")
		}
	})
}

func TestPluggableTraceFlightRecorder_Enabled(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns false", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}
		if p.Enabled() {
			t.Fatal("expected false")
		}
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{
			EnabledFn: func() bool { return true },
		}

		if !p.Enabled() {
			t.Fatal("expected true")
		}
	})
}

func TestPluggableTraceFlightRecorder_WriteTo(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns zero", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}

		n, err := p.WriteTo(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n != 0 {
			t.Fatalf("got %d, want 0", n)
		}
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		p := &PluggableTraceFlightRecorder{
			WriteToFn: func(w io.Writer) (int64, error) {
				n, err := w.Write([]byte("trace"))

				return int64(n), err
			},
		}

		n, err := p.WriteTo(&buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n != 5 {
			t.Fatalf("got %d, want 5", n)
		}
	})
}
