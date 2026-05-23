package tests_test

import (
	"context"
	"testing"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	"github.com/guidomantilla/yarumo/common/lifecycle/tests"
)

func TestAssertIdempotentStop(t *testing.T) {
	t.Parallel()

	t.Run("passes for fresh NewComponent", func(t *testing.T) {
		t.Parallel()

		c := lifecycle.NewComponent("idempotent-1")
		tests.AssertIdempotentStop(t, c)
	})

	t.Run("passes for a worker-style component already Started", func(t *testing.T) {
		t.Parallel()

		c := lifecycle.NewComponent("idempotent-2")

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		tests.AssertIdempotentStop(t, c)
	})

	t.Run("passes for a component already Stopped once", func(t *testing.T) {
		t.Parallel()

		c := lifecycle.NewComponent("idempotent-3")

		err := c.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		tests.AssertIdempotentStop(t, c)
	})
}
