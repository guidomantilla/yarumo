package assert

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestNotEmpty(t *testing.T) {
	// non-empty -> should not exit
	NotEmpty("x", "msg")

	// empty -> would exit, disable logging to avoid os.Exit
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(prev)
	NotEmpty("", "msg")
}

func TestNotNil(t *testing.T) {
	// non-nil -> should not exit
	v := 1
	NotNil(&v, "msg")

	// nil -> would exit, disable logging to avoid os.Exit
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(prev)
	var p *int
	NotNil(p, "msg")
}

func TestEqual(t *testing.T) {
	// equal -> should not exit
	Equal(5, 5, "msg")

	// not equal -> would exit, disable logging
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(prev)
	Equal(5, 6, "msg")
}

func TestNotEqual(t *testing.T) {
	// not equal -> should not exit
	NotEqual(5, 6, "msg")

	// equal -> would exit, disable logging
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(prev)
	NotEqual(7, 7, "msg")
}

func TestTrue(t *testing.T) {
	// true -> should not exit
	True(true, "msg")

	// false -> would exit, disable logging
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(prev)
	True(false, "msg")
}

func TestFalse(t *testing.T) {
	// false -> should not exit
	False(false, "msg")

	// true -> would exit, disable logging
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)
	defer zerolog.SetGlobalLevel(prev)
	False(true, "msg")
}
