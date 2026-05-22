package http

import (
	"context"
	stdlog "log"
	"net"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns zero-valued options when no arguments", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()
		if o == nil {
			t.Fatal("expected non-nil options")
		}

		if o.readTimeout != 0 {
			t.Fatalf("expected zero readTimeout, got %v", o.readTimeout)
		}

		if o.tlsEnabled {
			t.Fatal("expected tlsEnabled to be false")
		}
	})

	t.Run("applies multiple options", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(
			WithReadTimeout(5*time.Second),
			WithWriteTimeout(10*time.Second),
		)

		if o.readTimeout != 5*time.Second {
			t.Fatalf("expected 5s readTimeout, got %v", o.readTimeout)
		}

		if o.writeTimeout != 10*time.Second {
			t.Fatalf("expected 10s writeTimeout, got %v", o.writeTimeout)
		}
	})
}

func TestWithReadTimeout(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithReadTimeout(7 * time.Second))
	if o.readTimeout != 7*time.Second {
		t.Fatalf("expected 7s, got %v", o.readTimeout)
	}
}

func TestWithWriteTimeout(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithWriteTimeout(3 * time.Second))
	if o.writeTimeout != 3*time.Second {
		t.Fatalf("expected 3s, got %v", o.writeTimeout)
	}
}

func TestWithIdleTimeout(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithIdleTimeout(time.Minute))
	if o.idleTimeout != time.Minute {
		t.Fatalf("expected 1m, got %v", o.idleTimeout)
	}
}

func TestWithReadHeaderTimeout(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithReadHeaderTimeout(2 * time.Second))
	if o.readHeaderTimeout != 2*time.Second {
		t.Fatalf("expected 2s, got %v", o.readHeaderTimeout)
	}
}

func TestWithMaxHeaderBytes(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithMaxHeaderBytes(8192))
	if o.maxHeaderBytes != 8192 {
		t.Fatalf("expected 8192, got %d", o.maxHeaderBytes)
	}
}

func TestWithErrorLog(t *testing.T) {
	t.Parallel()

	logger := stdlog.New(stdlog.Writer(), "test", 0)

	o := NewOptions(WithErrorLog(logger))
	if o.errorLog != logger {
		t.Fatal("expected errorLog to match the provided logger")
	}
}

func TestWithBaseContext(t *testing.T) {
	t.Parallel()

	fn := func(_ net.Listener) context.Context { return context.Background() }

	o := NewOptions(WithBaseContext(fn))
	if o.baseContext == nil {
		t.Fatal("expected baseContext to be set")
	}
}

func TestWithTLS(t *testing.T) {
	t.Parallel()

	t.Run("enables TLS with valid cert and key", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithTLS("cert.pem", "key.pem"))
		if !o.tlsEnabled {
			t.Fatal("expected tlsEnabled true")
		}

		if o.tlsCertFile != "cert.pem" || o.tlsKeyFile != "key.pem" {
			t.Fatalf("got cert=%q key=%q", o.tlsCertFile, o.tlsKeyFile)
		}
	})

	t.Run("ignores empty cert", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithTLS("", "key.pem"))
		if o.tlsEnabled {
			t.Fatal("expected tlsEnabled false")
		}
	})

	t.Run("ignores empty key", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithTLS("cert.pem", ""))
		if o.tlsEnabled {
			t.Fatal("expected tlsEnabled false")
		}
	})
}
