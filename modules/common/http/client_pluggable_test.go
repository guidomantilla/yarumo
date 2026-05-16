package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"testing"
)

func TestPluggableClient_Do(t *testing.T) {
	t.Parallel()

	t.Run("delegates to DoFn", func(t *testing.T) {
		t.Parallel()

		called := false
		fc := &PluggableClient{
			DoFn: func(req *stdhttp.Request) (*stdhttp.Response, error) {
				called = true
				return &stdhttp.Response{StatusCode: 204, Body: stdhttp.NoBody}, nil
			},
		}

		req := newRequest(t, context.Background())

		res, err := fc.Do(req)
		if err != nil {
			t.Fatalf("Do returned error: %v", err)
		}

		defer func() { _ = res.Body.Close() }()

		if res.StatusCode != stdhttp.StatusNoContent {
			t.Fatalf("unexpected response: %+v", res)
		}

		if !called {
			t.Fatal("DoFn was not invoked")
		}
	})

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn: func(req *stdhttp.Request) (*stdhttp.Response, error) {
				return &stdhttp.Response{StatusCode: 200, Body: stdhttp.NoBody}, nil
			},
		}

		_, err := fc.Do(nil) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		if !errors.Is(err, ErrHttpRequestNil) {
			t.Fatalf("error does not wrap ErrHttpRequestNil: %v", err)
		}
	})
}

func TestPluggableClient_LimiterEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{DoFn: NoopDo}

		if fc.LimiterEnabled() {
			t.Fatal("LimiterEnabled with nil fn should return false")
		}
	})

	t.Run("delegates to fn true", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			LimiterEnabledFn: func() bool { return true },
		}

		if !fc.LimiterEnabled() {
			t.Fatal("LimiterEnabled should return true")
		}
	})

	t.Run("delegates to fn false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			LimiterEnabledFn: func() bool { return false },
		}

		if fc.LimiterEnabled() {
			t.Fatal("LimiterEnabled should return false")
		}
	})
}

func TestPluggableClient_RetrierEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{DoFn: NoopDo}

		if fc.RetrierEnabled() {
			t.Fatal("RetrierEnabled with nil fn should return false")
		}
	})

	t.Run("delegates to fn true", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			RetrierEnabledFn: func() bool { return true },
		}

		if !fc.RetrierEnabled() {
			t.Fatal("RetrierEnabled should return true")
		}
	})

	t.Run("delegates to fn false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			RetrierEnabledFn: func() bool { return false },
		}

		if fc.RetrierEnabled() {
			t.Fatal("RetrierEnabled should return false")
		}
	})
}
