// Package main demonstrates the generic Cache[K, V] contract and the
// reference in-memory backend from common/cache.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	ccache "github.com/guidomantilla/yarumo/common/cache"
)

func main() {
	ctx := context.Background()

	demoSetGet(ctx)
	demoHasDelete(ctx)
	demoMiss(ctx)
	demoCodec()
	demoResolveKeyPrefix()
}

// demoSetGet stores a couple of entries and reads one back.
func demoSetGet(ctx context.Context) {
	fmt.Println("=== Set/Get ===")

	cache := ccache.NewMemoryCache[string, int]("counters")
	_ = cache.Set(ctx, "hits", 7, 0)
	_ = cache.Set(ctx, "misses", 2, 0)

	hits, _ := cache.Get(ctx, "hits")
	fmt.Printf("  counters.hits = %d\n", hits)
}

// demoHasDelete verifies presence and removes an entry.
func demoHasDelete(ctx context.Context) {
	fmt.Println("=== Has/Delete ===")

	cache := ccache.NewMemoryCache[string, string]("sessions")
	_ = cache.Set(ctx, "alice", "token-A", 0)

	ok, _ := cache.Has(ctx, "alice")
	fmt.Printf("  Has(alice) = %v\n", ok)

	_ = cache.Delete(ctx, "alice")

	ok, _ = cache.Has(ctx, "alice")
	fmt.Printf("  Has(alice) after Delete = %v\n", ok)
}

// demoMiss shows the ErrCacheMiss sentinel returned on absent keys.
func demoMiss(ctx context.Context) {
	fmt.Println("=== Miss ===")

	cache := ccache.NewMemoryCache[string, int]("empty")

	_, err := cache.Get(ctx, "nope")
	if errors.Is(err, ccache.ErrCacheMiss) {
		fmt.Println("  Get(nope) -> ErrCacheMiss as expected")
	}
}

// demoCodec exercises the default JSONCodec that backends use to (de)serialize values.
func demoCodec() {
	fmt.Println("=== Codec ===")

	codec := ccache.JSONCodec{}
	data, _ := codec.Encode(map[string]int{"a": 1, "b": 2})
	fmt.Printf("  Encode -> %s\n", string(data))

	var decoded map[string]int
	_ = codec.Decode(data, &decoded)
	// Re-encode for deterministic output.
	roundTrip, _ := json.Marshal(decoded)
	fmt.Printf("  Decode -> %s\n", string(roundTrip))
}

// demoResolveKeyPrefix shows the "name:" default vs an explicit override.
func demoResolveKeyPrefix() {
	fmt.Println("=== ResolveKeyPrefix ===")

	fmt.Printf("  default   -> %q\n", ccache.ResolveKeyPrefix("users", ""))
	fmt.Printf("  overridden -> %q\n", ccache.ResolveKeyPrefix("users", "app:v2:"))
}
