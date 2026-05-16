package diagnostics

import (
	"bytes"
	"io"
	"testing"
)

func TestPluggableTraceFlightRecorder_Start(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns nil", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}

		err := p.Start()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		called := false
		p := &PluggableTraceFlightRecorder{
			StartFn: func() error {
				called = true
				return nil
			},
		}

		err := p.Start()
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

	t.Run("nil fn does not panic", func(t *testing.T) {
		t.Parallel()

		p := &PluggableTraceFlightRecorder{}
		p.Stop()
	})

	t.Run("delegates to fn", func(t *testing.T) {
		t.Parallel()

		called := false
		p := &PluggableTraceFlightRecorder{
			StopFn: func() { called = true },
		}

		p.Stop()

		if !called {
			t.Fatal("expected StopFn to be called")
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
