package lifecycle_test

import (
	"testing"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	lctests "github.com/guidomantilla/yarumo/common/lifecycle/tests"
)

func TestComponent_StopIsIdempotent(t *testing.T) {
	t.Parallel()

	c := lifecycle.NewComponent("idempotent")
	lctests.AssertIdempotentStop(t, c)
}
