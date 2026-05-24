package otelhttp

import (
	"context"
	"net/http"
	"strings"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/propagation"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults pull globals and package-path scope name", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.meterProvider == nil {
			t.Fatal("expected non-nil default meter provider")
		}
		if opts.tracerProvider == nil {
			t.Fatal("expected non-nil default tracer provider")
		}
		if opts.propagator == nil {
			t.Fatal("expected non-nil default propagator")
		}
		if opts.spanNameFn == nil {
			t.Fatal("expected non-nil default span name fn")
		}
		if !strings.HasSuffix(opts.name, "/extensions/telemetry/otel/http") {
			t.Fatalf("name = %q, want package import path", opts.name)
		}
	})

	t.Run("applies each option in order", func(t *testing.T) {
		t.Parallel()

		meterProvider := sdkmetric.NewMeterProvider()
		tracerProvider := sdktrace.NewTracerProvider()
		propagator := propagation.TraceContext{}

		opts := NewOptions(
			WithName("custom.scope"),
			WithMeterProvider(meterProvider),
			WithTracerProvider(tracerProvider),
			WithPropagator(propagator),
			WithHeaderRedaction("x-api-key"),
			WithSpanNameFn(func(*http.Request) string { return "fixed" }),
		)

		if opts.name != "custom.scope" {
			t.Fatalf("name = %q, want %q", opts.name, "custom.scope")
		}
		if opts.meterProvider != meterProvider {
			t.Fatal("expected meter provider to be overridden")
		}
		if opts.tracerProvider != tracerProvider {
			t.Fatal("expected tracer provider to be overridden")
		}
		if opts.propagator != propagator {
			t.Fatal("expected propagator to be overridden")
		}
		_, ok := opts.redaction["X-Api-Key"]
		if !ok {
			t.Fatalf("expected canonicalized X-Api-Key in redaction set, got %v", opts.redaction)
		}
	})
}

func TestWithName(t *testing.T) {
	t.Parallel()

	t.Run("overrides default when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithName("svc"))
		if opts.name != "svc" {
			t.Fatalf("name = %q, want %q", opts.name, "svc")
		}
	})

	t.Run("ignores empty, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithName(""))
		if opts.name == "" {
			t.Fatal("expected default name to remain")
		}
	})
}

func TestWithMeterProvider(t *testing.T) {
	t.Parallel()

	t.Run("overrides default when non-nil", func(t *testing.T) {
		t.Parallel()

		provider := sdkmetric.NewMeterProvider()
		opts := NewOptions(WithMeterProvider(provider))
		if opts.meterProvider != provider {
			t.Fatal("expected meter provider to be overridden")
		}
	})

	t.Run("ignores nil, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMeterProvider(nil))
		if opts.meterProvider == nil {
			t.Fatal("expected default meter provider to remain")
		}
	})
}

func TestWithTracerProvider(t *testing.T) {
	t.Parallel()

	t.Run("overrides default when non-nil", func(t *testing.T) {
		t.Parallel()

		provider := sdktrace.NewTracerProvider()
		opts := NewOptions(WithTracerProvider(provider))
		if opts.tracerProvider != provider {
			t.Fatal("expected tracer provider to be overridden")
		}
	})

	t.Run("ignores nil, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTracerProvider(nil))
		if opts.tracerProvider == nil {
			t.Fatal("expected default tracer provider to remain")
		}
	})
}

func TestWithPropagator(t *testing.T) {
	t.Parallel()

	t.Run("overrides default when non-nil", func(t *testing.T) {
		t.Parallel()

		propagator := propagation.TraceContext{}
		opts := NewOptions(WithPropagator(propagator))
		if opts.propagator != propagator {
			t.Fatal("expected propagator to be overridden")
		}
	})

	t.Run("ignores nil, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPropagator(nil))
		if opts.propagator == nil {
			t.Fatal("expected default propagator to remain")
		}
	})
}

func TestWithHeaderRedaction(t *testing.T) {
	t.Parallel()

	t.Run("canonicalizes header names and accumulates across calls", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithHeaderRedaction("authorization", "X-API-Key"),
			WithHeaderRedaction("cookie"),
		)
		expected := []string{"Authorization", "X-Api-Key", "Cookie"}
		for _, name := range expected {
			_, ok := opts.redaction[name]
			if !ok {
				t.Fatalf("expected %q in redaction set, got %v", name, opts.redaction)
			}
		}
	})

	t.Run("empty argument list is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithHeaderRedaction())
		if opts.redaction != nil {
			t.Fatalf("expected nil redaction set, got %v", opts.redaction)
		}
	})
}

func TestWithSpanNameFn(t *testing.T) {
	t.Parallel()

	t.Run("overrides default when non-nil", func(t *testing.T) {
		t.Parallel()

		fn := func(*http.Request) string { return "fixed" }
		opts := NewOptions(WithSpanNameFn(fn))
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://x", http.NoBody)
		if opts.spanNameFn(req) != "fixed" {
			t.Fatal("expected span name fn to be overridden")
		}
	})

	t.Run("ignores nil, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithSpanNameFn(nil))
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://x", http.NoBody)
		if opts.spanNameFn(req) != "HTTP GET" {
			t.Fatalf("default span name fn unexpected: %q", opts.spanNameFn(req))
		}
	})
}
