package otel

import (
	"testing"
	"time"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.endpoint != "localhost:4317" {
			t.Fatalf("expected endpoint localhost:4317, got %s", opts.endpoint)
		}
		if !opts.secure {
			t.Fatal("expected secure true")
		}
		if opts.resource == nil {
			t.Fatal("expected non-nil resource")
		}
		if opts.tracerPropagators == nil {
			t.Fatal("expected non-nil tracerPropagators")
		}
		if opts.tracerExporterOptions == nil {
			t.Fatal("expected non-nil tracerExporterOptions")
		}
		if opts.tracerProviderOptions == nil {
			t.Fatal("expected non-nil tracerProviderOptions")
		}
		if opts.meterExporterOptions == nil {
			t.Fatal("expected non-nil meterExporterOptions")
		}
		if opts.meterProviderOptions == nil {
			t.Fatal("expected non-nil meterProviderOptions")
		}
		if opts.meterInterval != time.Millisecond*60000 {
			t.Fatalf("expected meterInterval 60000ms, got %v", opts.meterInterval)
		}
		if opts.meterRuntimeMetricsEnabled {
			t.Fatal("expected meterRuntimeMetricsEnabled false")
		}
		if opts.meterRuntimeMetricsInterval != runtimemetrics.DefaultMinimumReadMemStatsInterval {
			t.Fatalf("expected meterRuntimeMetricsInterval %v, got %v", runtimemetrics.DefaultMinimumReadMemStatsInterval, opts.meterRuntimeMetricsInterval)
		}
		if opts.meterRuntimeMetricsOptions == nil {
			t.Fatal("expected non-nil meterRuntimeMetricsOptions")
		}
		if opts.loggerExporterOptions == nil {
			t.Fatal("expected non-nil loggerExporterOptions")
		}
		if opts.loggerProviderOptions == nil {
			t.Fatal("expected non-nil loggerProviderOptions")
		}
	})

	t.Run("single option", func(t *testing.T) {
		t.Parallel()

		endpoint := "collector:4317"
		opts := NewOptions(WithEndpoint(endpoint))

		if opts.endpoint != endpoint {
			t.Fatalf("expected endpoint %s, got %s", endpoint, opts.endpoint)
		}
		if !opts.secure {
			t.Fatal("expected secure true (default unchanged)")
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		t.Parallel()

		endpoint := "otel-collector:4317"
		opts := NewOptions(WithEndpoint(endpoint), WithInsecure(), WithMeterInterval(time.Second*30))

		if opts.endpoint != endpoint {
			t.Fatalf("expected endpoint %s, got %s", endpoint, opts.endpoint)
		}
		if opts.secure {
			t.Fatal("expected secure false")
		}
		if opts.meterInterval != time.Second*30 {
			t.Fatalf("expected meterInterval 30s, got %v", opts.meterInterval)
		}
	})
}

func TestWithEndpoint(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		endpoint := "remote-collector:4317"
		opts := NewOptions(WithEndpoint(endpoint))

		if opts.endpoint != endpoint {
			t.Fatalf("expected endpoint %s, got %s", endpoint, opts.endpoint)
		}
	})

	t.Run("empty ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithEndpoint(""))

		if opts.endpoint != "localhost:4317" {
			t.Fatalf("expected default endpoint localhost:4317, got %s", opts.endpoint)
		}
	})
}

func TestWithInsecure(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithInsecure())

	if opts.secure {
		t.Fatal("expected secure false")
	}
}

func TestWithResource(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		res := &resource.Resource{}
		opts := NewOptions(WithResource(res))

		if opts.resource != res {
			t.Fatal("expected resource to be set")
		}
	})

	t.Run("nil ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithResource(nil))

		if opts.resource == nil {
			t.Fatal("expected default resource to remain non-nil")
		}
	})
}

func TestWithTracerPropagators(t *testing.T) {
	t.Parallel()

	propagators := []propagation.TextMapPropagator{propagation.TraceContext{}}
	opts := NewOptions(WithTracerPropagators(propagators...))

	if len(opts.tracerPropagators) != 1 {
		t.Fatalf("expected 1 propagator, got %d", len(opts.tracerPropagators))
	}
}

func TestWithTracerExporterOptions(t *testing.T) {
	t.Parallel()

	exporterOpts := []otlptracegrpc.Option{otlptracegrpc.WithInsecure()}
	opts := NewOptions(WithTracerExporterOptions(exporterOpts...))

	if len(opts.tracerExporterOptions) != 1 {
		t.Fatalf("expected 1 exporter option, got %d", len(opts.tracerExporterOptions))
	}
}

func TestWithTracerProviderOptions(t *testing.T) {
	t.Parallel()

	providerOpts := []sdktrace.TracerProviderOption{sdktrace.WithSampler(sdktrace.AlwaysSample())}
	opts := NewOptions(WithTracerProviderOptions(providerOpts...))

	if len(opts.tracerProviderOptions) != 1 {
		t.Fatalf("expected 1 provider option, got %d", len(opts.tracerProviderOptions))
	}
}

func TestWithMeterExporterOptions(t *testing.T) {
	t.Parallel()

	exporterOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithInsecure()}
	opts := NewOptions(WithMeterExporterOptions(exporterOpts...))

	if len(opts.meterExporterOptions) != 1 {
		t.Fatalf("expected 1 exporter option, got %d", len(opts.meterExporterOptions))
	}
}

func TestWithMeterProviderOptions(t *testing.T) {
	t.Parallel()

	providerOpts := []sdkmetric.Option{sdkmetric.WithResource(&resource.Resource{})}
	opts := NewOptions(WithMeterProviderOptions(providerOpts...))

	if len(opts.meterProviderOptions) != 1 {
		t.Fatalf("expected 1 provider option, got %d", len(opts.meterProviderOptions))
	}
}

func TestWithMeterInterval(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMeterInterval(time.Second * 30))

		if opts.meterInterval != time.Second*30 {
			t.Fatalf("expected meterInterval 30s, got %v", opts.meterInterval)
		}
	})

	t.Run("zero ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMeterInterval(0))

		if opts.meterInterval != time.Millisecond*60000 {
			t.Fatalf("expected default meterInterval 60000ms, got %v", opts.meterInterval)
		}
	})

	t.Run("negative ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMeterInterval(-time.Second))

		if opts.meterInterval != time.Millisecond*60000 {
			t.Fatalf("expected default meterInterval 60000ms, got %v", opts.meterInterval)
		}
	})
}

func TestWithMeterRuntimeMetricsEnabled(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithMeterRuntimeMetricsEnabled(true))

	if !opts.meterRuntimeMetricsEnabled {
		t.Fatal("expected meterRuntimeMetricsEnabled true")
	}
}

func TestWithMeterRuntimeMetricsInterval(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMeterRuntimeMetricsInterval(time.Second * 10))

		if opts.meterRuntimeMetricsInterval != time.Second*10 {
			t.Fatalf("expected meterRuntimeMetricsInterval 10s, got %v", opts.meterRuntimeMetricsInterval)
		}
	})

	t.Run("zero ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMeterRuntimeMetricsInterval(0))

		if opts.meterRuntimeMetricsInterval != runtimemetrics.DefaultMinimumReadMemStatsInterval {
			t.Fatalf("expected default meterRuntimeMetricsInterval, got %v", opts.meterRuntimeMetricsInterval)
		}
	})

	t.Run("negative ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMeterRuntimeMetricsInterval(-time.Second))

		if opts.meterRuntimeMetricsInterval != runtimemetrics.DefaultMinimumReadMemStatsInterval {
			t.Fatalf("expected default meterRuntimeMetricsInterval, got %v", opts.meterRuntimeMetricsInterval)
		}
	})
}

func TestWithMeterRuntimeMetricsOptions(t *testing.T) {
	t.Parallel()

	runtimeOpts := []runtimemetrics.Option{runtimemetrics.WithMinimumReadMemStatsInterval(time.Second * 5)}
	opts := NewOptions(WithMeterRuntimeMetricsOptions(runtimeOpts...))

	if len(opts.meterRuntimeMetricsOptions) != 1 {
		t.Fatalf("expected 1 runtime metrics option, got %d", len(opts.meterRuntimeMetricsOptions))
	}
}

func TestWithLoggerExporterOptions(t *testing.T) {
	t.Parallel()

	exporterOpts := []otlploggrpc.Option{otlploggrpc.WithInsecure()}
	opts := NewOptions(WithLoggerExporterOptions(exporterOpts...))

	if len(opts.loggerExporterOptions) != 1 {
		t.Fatalf("expected 1 exporter option, got %d", len(opts.loggerExporterOptions))
	}
}

func TestWithLoggerProviderOptions(t *testing.T) {
	t.Parallel()

	providerOpts := []sdklog.LoggerProviderOption{sdklog.WithResource(&resource.Resource{})}
	opts := NewOptions(WithLoggerProviderOptions(providerOpts...))

	if len(opts.loggerProviderOptions) != 1 {
		t.Fatalf("expected 1 provider option, got %d", len(opts.loggerProviderOptions))
	}
}
