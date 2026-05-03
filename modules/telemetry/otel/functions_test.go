package otel

import (
	"context"
	"errors"
	"testing"
	"time"
)

func Test_noopStop(t *testing.T) {
	t.Parallel()

	noopStop(context.Background(), time.Second)
}

func TestResources(t *testing.T) {
	t.Parallel()

	res, err := Resources(context.Background(), "test-service", "1.0.0", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestProfiler(t *testing.T) {
	t.Parallel()

	stopFn, err := Profiler(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stopFn(context.Background(), time.Second)
}

func TestTracer(t *testing.T) {
	t.Run("insecure", func(t *testing.T) {
		stopFn, err := Tracer(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("secure", func(t *testing.T) {
		stopFn, err := Tracer(context.Background(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("stop with cancelled context", func(t *testing.T) {
		stopFn, err := Tracer(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Millisecond)
	})

}

func TestMeter(t *testing.T) {
	t.Run("insecure", func(t *testing.T) {
		stopFn, err := Meter(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("secure", func(t *testing.T) {
		stopFn, err := Meter(context.Background(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("runtime metrics enabled", func(t *testing.T) {
		stopFn, err := Meter(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"), WithMeterRuntimeMetricsEnabled(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("runtime metrics disabled", func(t *testing.T) {
		stopFn, err := Meter(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"), WithMeterRuntimeMetricsEnabled(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("stop with cancelled context", func(t *testing.T) {
		stopFn, err := Meter(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Millisecond)
	})

}

func TestLogger(t *testing.T) {
	t.Run("insecure", func(t *testing.T) {
		stopFn, err := Logger(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("secure", func(t *testing.T) {
		stopFn, err := Logger(context.Background(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("stop with cancelled context", func(t *testing.T) {
		stopFn, err := Logger(context.Background(), WithInsecure(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Millisecond)
	})

}

func TestObserve(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hookFn := func(ctx context.Context) (context.Context, error) {
			return ctx, nil
		}

		ctx, stopFn, err := Observe(context.Background(), "test-service", "1.0.0", "test", hookFn, WithInsecure(), WithEndpoint("localhost:4317"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}

		stopFn(context.Background(), time.Second)
	})

	t.Run("hook failure", func(t *testing.T) {
		hookFn := func(_ context.Context) (context.Context, error) {
			return nil, errors.New("hook failed")
		}

		_, _, err := Observe(context.Background(), "test-service", "1.0.0", "test", hookFn, WithInsecure(), WithEndpoint("localhost:4317"))
		if err == nil {
			t.Fatal("expected error from hook failure")
		}
	})
}
