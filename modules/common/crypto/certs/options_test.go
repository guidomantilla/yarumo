package certs

import (
	"net"
	"testing"
	"time"
)

const (
	defaultOrganization = "Development"
	defaultDNSName      = "localhost"
)

func TestNewSelfSignedOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions()

		if opts.organization != defaultOrganization {
			t.Fatalf("expected organization 'Development', got %q", opts.organization)
		}
		if opts.validity != 8760*time.Hour {
			t.Fatalf("expected validity 8760h, got %v", opts.validity)
		}
		if len(opts.dnsNames) != 1 || opts.dnsNames[0] != defaultDNSName {
			t.Fatalf("expected dnsNames [localhost], got %v", opts.dnsNames)
		}
		if len(opts.ipAddresses) != 2 {
			t.Fatalf("expected 2 ip addresses, got %d", len(opts.ipAddresses))
		}
		if !opts.ipAddresses[0].Equal(net.IPv4(127, 0, 0, 1)) {
			t.Fatalf("expected 127.0.0.1, got %v", opts.ipAddresses[0])
		}
		if !opts.ipAddresses[1].Equal(net.IPv6loopback) {
			t.Fatalf("expected ::1, got %v", opts.ipAddresses[1])
		}
		if opts.isCA {
			t.Fatal("expected isCA false by default")
		}
	})

	t.Run("applies options", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(
			WithOrganization("TestOrg"),
			WithValidity(24*time.Hour),
			WithDNSNames("example.com"),
			WithIPAddresses(net.IPv4(10, 0, 0, 1)),
			WithIsCA(true),
		)

		if opts.organization != "TestOrg" {
			t.Fatalf("expected organization 'TestOrg', got %q", opts.organization)
		}
		if opts.validity != 24*time.Hour {
			t.Fatalf("expected validity 24h, got %v", opts.validity)
		}
		if len(opts.dnsNames) != 1 || opts.dnsNames[0] != "example.com" {
			t.Fatalf("expected dnsNames [example.com], got %v", opts.dnsNames)
		}
		if len(opts.ipAddresses) != 1 || !opts.ipAddresses[0].Equal(net.IPv4(10, 0, 0, 1)) {
			t.Fatalf("expected ip [10.0.0.1], got %v", opts.ipAddresses)
		}
		if !opts.isCA {
			t.Fatal("expected isCA true")
		}
	})
}

func TestWithOrganization(t *testing.T) {
	t.Parallel()

	t.Run("sets organization", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithOrganization("MyOrg"))

		if opts.organization != "MyOrg" {
			t.Fatalf("expected 'MyOrg', got %q", opts.organization)
		}
	})

	t.Run("ignores empty string", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithOrganization(""))

		if opts.organization != defaultOrganization {
			t.Fatalf("expected default 'Development', got %q", opts.organization)
		}
	})
}

func TestWithValidity(t *testing.T) {
	t.Parallel()

	t.Run("sets validity", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithValidity(48 * time.Hour))

		if opts.validity != 48*time.Hour {
			t.Fatalf("expected 48h, got %v", opts.validity)
		}
	})

	t.Run("sets minimum validity of one hour", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithValidity(time.Hour))

		if opts.validity != time.Hour {
			t.Fatalf("expected 1h, got %v", opts.validity)
		}
	})

	t.Run("ignores validity less than one hour", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithValidity(30 * time.Minute))

		if opts.validity != 8760*time.Hour {
			t.Fatalf("expected default 8760h, got %v", opts.validity)
		}
	})

	t.Run("ignores zero validity", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithValidity(0))

		if opts.validity != 8760*time.Hour {
			t.Fatalf("expected default 8760h, got %v", opts.validity)
		}
	})

	t.Run("ignores negative validity", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithValidity(-time.Hour))

		if opts.validity != 8760*time.Hour {
			t.Fatalf("expected default 8760h, got %v", opts.validity)
		}
	})
}

func TestWithDNSNames(t *testing.T) {
	t.Parallel()

	t.Run("sets dns names", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithDNSNames("a.com", "b.com"))

		if len(opts.dnsNames) != 2 {
			t.Fatalf("expected 2 dns names, got %d", len(opts.dnsNames))
		}
		if opts.dnsNames[0] != "a.com" || opts.dnsNames[1] != "b.com" {
			t.Fatalf("expected [a.com, b.com], got %v", opts.dnsNames)
		}
	})

	t.Run("ignores empty list", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithDNSNames())

		if len(opts.dnsNames) != 1 || opts.dnsNames[0] != defaultDNSName {
			t.Fatalf("expected default [localhost], got %v", opts.dnsNames)
		}
	})
}

func TestWithIPAddresses(t *testing.T) {
	t.Parallel()

	t.Run("sets ip addresses", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithIPAddresses(net.IPv4(192, 168, 1, 1)))

		if len(opts.ipAddresses) != 1 {
			t.Fatalf("expected 1 ip address, got %d", len(opts.ipAddresses))
		}
		if !opts.ipAddresses[0].Equal(net.IPv4(192, 168, 1, 1)) {
			t.Fatalf("expected 192.168.1.1, got %v", opts.ipAddresses[0])
		}
	})

	t.Run("ignores empty list", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithIPAddresses())

		if len(opts.ipAddresses) != 2 {
			t.Fatalf("expected default 2 ip addresses, got %d", len(opts.ipAddresses))
		}
	})
}

func TestWithIsCA(t *testing.T) {
	t.Parallel()

	t.Run("sets is ca true", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithIsCA(true))

		if !opts.isCA {
			t.Fatal("expected isCA true")
		}
	})

	t.Run("sets is ca false", func(t *testing.T) {
		t.Parallel()

		opts := NewSelfSignedOptions(WithIsCA(false))

		if opts.isCA {
			t.Fatal("expected isCA false")
		}
	})
}
