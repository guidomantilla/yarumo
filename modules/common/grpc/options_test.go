package grpc

import (
	"testing"

	"google.golang.org/grpc"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns empty options when no arguments", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts == nil {
			t.Fatal("expected non-nil options")
		}

		if len(opts.services) != 0 {
			t.Fatalf("expected 0 services, got %d", len(opts.services))
		}

		if len(opts.serverOptions) != 0 {
			t.Fatalf("expected 0 server options, got %d", len(opts.serverOptions))
		}
	})

	t.Run("applies multiple options", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithService(&testServiceImpl{}, newTestDesc()),
			WithServerOption(grpc.MaxRecvMsgSize(1024)),
		)

		if len(opts.services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(opts.services))
		}

		if len(opts.serverOptions) != 1 {
			t.Fatalf("expected 1 server option, got %d", len(opts.serverOptions))
		}
	})
}

func TestWithService(t *testing.T) {
	t.Parallel()

	t.Run("ignores nil service", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithService(nil, newTestDesc()))

		if len(opts.services) != 0 {
			t.Fatalf("expected 0 services, got %d", len(opts.services))
		}
	})

	t.Run("ignores nil descriptor", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithService(&testServiceImpl{}, nil))

		if len(opts.services) != 0 {
			t.Fatalf("expected 0 services, got %d", len(opts.services))
		}
	})

	t.Run("registers valid service", func(t *testing.T) {
		t.Parallel()

		desc := newTestDesc()

		opts := NewOptions(WithService(&testServiceImpl{}, desc))

		if len(opts.services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(opts.services))
		}

		if opts.services[0].descriptor != desc {
			t.Fatal("descriptor mismatch")
		}
	})

	t.Run("registers multiple services", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithService(&testServiceImpl{}, &grpc.ServiceDesc{
				ServiceName: "test.A",
				HandlerType: (*testService)(nil),
			}),
			WithService(&testServiceImpl{}, &grpc.ServiceDesc{
				ServiceName: "test.B",
				HandlerType: (*testService)(nil),
			}),
		)

		if len(opts.services) != 2 {
			t.Fatalf("expected 2 services, got %d", len(opts.services))
		}
	})
}

func TestWithServerOption(t *testing.T) {
	t.Parallel()

	t.Run("adds server options", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithServerOption(grpc.MaxRecvMsgSize(1024)))

		if len(opts.serverOptions) != 1 {
			t.Fatalf("expected 1 server option, got %d", len(opts.serverOptions))
		}
	})

	t.Run("ignores nil individual options", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithServerOption(nil, grpc.MaxRecvMsgSize(1024), nil))

		if len(opts.serverOptions) != 1 {
			t.Fatalf("expected 1 server option, got %d", len(opts.serverOptions))
		}
	})

	t.Run("handles multiple calls", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithServerOption(grpc.MaxRecvMsgSize(1024)),
			WithServerOption(grpc.MaxSendMsgSize(2048)),
		)

		if len(opts.serverOptions) != 2 {
			t.Fatalf("expected 2 server options, got %d", len(opts.serverOptions))
		}
	})
}
