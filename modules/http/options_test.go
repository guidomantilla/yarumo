package http

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies defaults", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()
		if o == nil {
			t.Fatalf("NewOptions returned nil")
		}

		if o.readHeaderTimeout != 5*time.Second {
			t.Fatalf("readHeaderTimeout = %v, want 5s", o.readHeaderTimeout)
		}

		if o.readTimeout != 15*time.Second {
			t.Fatalf("readTimeout = %v, want 15s", o.readTimeout)
		}

		if o.writeTimeout != 15*time.Second {
			t.Fatalf("writeTimeout = %v, want 15s", o.writeTimeout)
		}

		if o.idleTimeout != 60*time.Second {
			t.Fatalf("idleTimeout = %v, want 60s", o.idleTimeout)
		}

		if o.maxHeaderBytes != 1<<20 {
			t.Fatalf("maxHeaderBytes = %d, want %d", o.maxHeaderBytes, 1<<20)
		}

		if o.tlsConfig != nil {
			t.Fatalf("tlsConfig = %v, want nil", o.tlsConfig)
		}
	})
}

func TestWithReadHeaderTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithReadHeaderTimeout(10 * time.Second))
		if o.readHeaderTimeout != 10*time.Second {
			t.Fatalf("readHeaderTimeout = %v, want 10s", o.readHeaderTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithReadHeaderTimeout(0))
		if o.readHeaderTimeout != 5*time.Second {
			t.Fatalf("WithReadHeaderTimeout(0) should keep default; got %v", o.readHeaderTimeout)
		}
	})
}

func TestWithReadTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithReadTimeout(30 * time.Second))
		if o.readTimeout != 30*time.Second {
			t.Fatalf("readTimeout = %v, want 30s", o.readTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithReadTimeout(0))
		if o.readTimeout != 15*time.Second {
			t.Fatalf("WithReadTimeout(0) should keep default; got %v", o.readTimeout)
		}
	})
}

func TestWithWriteTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithWriteTimeout(45 * time.Second))
		if o.writeTimeout != 45*time.Second {
			t.Fatalf("writeTimeout = %v, want 45s", o.writeTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithWriteTimeout(0))
		if o.writeTimeout != 15*time.Second {
			t.Fatalf("WithWriteTimeout(0) should keep default; got %v", o.writeTimeout)
		}
	})
}

func TestWithIdleTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithIdleTimeout(120 * time.Second))
		if o.idleTimeout != 120*time.Second {
			t.Fatalf("idleTimeout = %v, want 120s", o.idleTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithIdleTimeout(0))
		if o.idleTimeout != 60*time.Second {
			t.Fatalf("WithIdleTimeout(0) should keep default; got %v", o.idleTimeout)
		}
	})
}

func TestWithMaxHeaderBytes(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithMaxHeaderBytes(2 << 20))
		if o.maxHeaderBytes != 2<<20 {
			t.Fatalf("maxHeaderBytes = %d, want %d", o.maxHeaderBytes, 2<<20)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithMaxHeaderBytes(0))
		if o.maxHeaderBytes != 1<<20 {
			t.Fatalf("WithMaxHeaderBytes(0) should keep default; got %d", o.maxHeaderBytes)
		}
	})
}

func TestWithTLSConfig(t *testing.T) {
	t.Parallel()

	t.Run("applies config", func(t *testing.T) {
		t.Parallel()

		cfg := &tls.Config{MinVersion: tls.VersionTLS13}

		o := NewOptions(WithTLSConfig(cfg))
		if o.tlsConfig != cfg {
			t.Fatalf("tlsConfig not set")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithTLSConfig(nil))
		if o.tlsConfig != nil {
			t.Fatalf("WithTLSConfig(nil) should keep default nil; got %v", o.tlsConfig)
		}
	})
}
