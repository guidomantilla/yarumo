// Package main demonstrates common/uids: the abstract UID contract plus
// the Register / Lookup / Supported registry. The package carries NO
// concrete generators — the canonical UUID/ULID/NanoID/etc. implementations
// live in modules/extension/common/uids/. This demo wires a couple of
// tiny in-memory generators to exercise the registry end-to-end without
// pulling external dependencies.
package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"

	cuids "github.com/guidomantilla/yarumo/core/common/uids"
)

// counterGenerator returns monotonically increasing numeric strings prefixed
// with a label. It is intentionally simple — the demo's focus is the contract,
// not entropy quality.
func counterGenerator(prefix string) cuids.UIDFn {
	var counter atomic.Uint64

	return func() (string, error) {
		next := counter.Add(1)
		return prefix + "-" + strconv.FormatUint(next, 10), nil
	}
}

// failingGenerator always returns ErrGenerationFailed — useful for showing
// how callers can detect generator failures with errors.Is.
func failingGenerator() (string, error) {
	return "", cuids.ErrGenerationFailed
}

func main() {
	demoNewUID()
	demoRegistry()
	demoLookupMiss()
	demoFailure()
}

// demoNewUID wraps a UIDFn into a UID and calls Generate directly.
func demoNewUID() {
	fmt.Println("=== NewUID ===")

	gen := cuids.NewUID("counter-a", counterGenerator("a"))

	for i := 0; i < 3; i++ {
		id, _ := gen.Generate()
		fmt.Printf("  Generate() -> %s\n", id)
	}
}

// demoRegistry registers a couple of generators and lists them.
func demoRegistry() {
	fmt.Println("=== Register / Supported ===")

	cuids.Register(cuids.NewUID("counter-b", counterGenerator("b")))
	cuids.Register(cuids.NewUID("counter-c", counterGenerator("c")))

	for _, u := range cuids.Supported() {
		fmt.Printf("  registered: %s\n", u.Name())
	}
}

// demoLookupMiss queries an unregistered name and inspects the typed error.
func demoLookupMiss() {
	fmt.Println("=== Lookup miss ===")

	_, err := cuids.Lookup("does-not-exist")

	var domainErr *cuids.Error
	if errors.As(err, &domainErr) {
		fmt.Printf("  domain error type=%q msg=%s\n", domainErr.Type, domainErr.Err)
	}
}

// demoFailure shows how callers detect entropy failures via errors.Is.
func demoFailure() {
	fmt.Println("=== Generation failure ===")

	gen := cuids.NewUID("broken", failingGenerator)
	_, err := gen.Generate()

	if errors.Is(err, cuids.ErrGenerationFailed) {
		fmt.Printf("  Generate() -> ErrGenerationFailed as expected\n")
	}
}
