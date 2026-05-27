// Demo that exercises the public API of the extension/common/uids
// package:
//
//  1. The six preconfigured singletons (UuidV4, NanoID, Cuid2, UuidV7,
//     Ulid, XId) — each Generate() produces an ID and IsXxx confirms it.
//  2. The common/uids registry resolves them by name (Lookup) and
//     enumerates them (Supported).
//  3. A custom UID constructed via cuids.NewUID and added to the registry.
package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	cuids "github.com/guidomantilla/yarumo/common/uids"
	"github.com/guidomantilla/yarumo/config"
	uids "github.com/guidomantilla/yarumo/extension/common/uids"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/common/uids/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Generate from each preconfigured UID", demoSingletons},
		{"Validate generated IDs with IsXxx", demoValidators},
		{"Registry: Lookup + Supported", demoRegistry},
		{"Custom UID via NewUID", demoCustomUID},
	}

	for _, d := range demos {
		fmt.Printf("=== Demo: %s ===\n", d.title)
		err := d.fn(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", d.title, err)
		}
		fmt.Println()
	}

	return nil
}

// demoSingletons calls Generate on each shipped UID singleton.
func demoSingletons(_ context.Context) error {
	entries := []cuids.UID{
		uids.UuidV4, uids.UuidV7, uids.Ulid, uids.NanoID, uids.Cuid2, uids.XId,
	}

	for _, u := range entries {
		id, err := u.Generate()
		if err != nil {
			return fmt.Errorf("%s.Generate: %w", u.Name(), err)
		}
		fmt.Printf("  %-7s -> %s\n", u.Name(), id)
	}

	return nil
}

// demoValidators round-trips each generator through its matching IsXxx
// predicate to confirm the canonical format.
func demoValidators(_ context.Context) error {
	pairs := []struct {
		name string
		gen  cuids.UIDFn
		is   cuids.IsUIDFn
	}{
		{"UUIDv4", uids.UUIDv4, uids.IsUUID},
		{"UUIDv7", uids.UUIDv7, uids.IsUUID},
		{"ULID", uids.ULID, uids.IsULID},
		{"NanoID", uids.NANOID, uids.IsNanoID},
		{"CUID2", uids.CUID2, uids.IsCUID2},
		{"XID", uids.XID, uids.IsXID},
	}

	for _, p := range pairs {
		id, err := p.gen()
		if err != nil {
			return fmt.Errorf("%s.gen: %w", p.name, err)
		}
		ok := p.is(id)
		fmt.Printf("  %-7s -> %s -> Is%s=%v\n", p.name, id, p.name, ok)
		if !ok {
			return fmt.Errorf("%s validator rejected its own output", p.name)
		}
	}

	return nil
}

// demoRegistry exercises Lookup and Supported on the package-level
// registry populated by extension/common/uids' init().
func demoRegistry(_ context.Context) error {
	names := make([]string, 0)
	for _, u := range cuids.Supported() {
		names = append(names, u.Name())
	}
	sort.Strings(names)
	fmt.Printf("  registry has %d entries: %v\n", len(names), names)

	u, err := cuids.Lookup("UUIDv4")
	if err != nil {
		return fmt.Errorf("Lookup(UUIDv4): %w", err)
	}

	id, err := u.Generate()
	if err != nil {
		return fmt.Errorf("Generate: %w", err)
	}
	fmt.Printf("  Lookup(\"UUIDv4\").Generate() -> %s\n", id)

	_, err = cuids.Lookup("MD5")
	if err == nil {
		return fmt.Errorf("Lookup(MD5) should have failed")
	}
	fmt.Printf("  Lookup(\"MD5\") -> %v\n", err)

	return nil
}

// demoCustomUID builds a custom UID with NewUID + a trivial generator
// and registers it.
func demoCustomUID(_ context.Context) error {
	counter := 0
	myFn := func() (string, error) {
		counter++
		return fmt.Sprintf("local-%04d", counter), nil
	}

	myUID := cuids.NewUID("LOCAL", myFn)
	cuids.Register(myUID)

	for i := 0; i < 3; i++ {
		id, err := myUID.Generate()
		if err != nil {
			return fmt.Errorf("Generate: %w", err)
		}
		fmt.Printf("  LOCAL -> %s\n", id)
	}

	resolved, err := cuids.Lookup("LOCAL")
	if err != nil {
		return fmt.Errorf("Lookup: %w", err)
	}
	fmt.Printf("  Lookup(\"LOCAL\") found: name=%s\n", resolved.Name())

	return nil
}
