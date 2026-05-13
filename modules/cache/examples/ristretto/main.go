package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guidomantilla/yarumo/cache"
)

func main() {
	ctx := context.Background()

	c, err := cache.NewCache[string, []byte](
		cache.WithBackend(cache.BackendRistretto),
		cache.WithTTL(time.Minute),
	)
	if err != nil {
		fmt.Println("failed to create ristretto cache:", err)
		return
	}
	defer func() { _ = c.Stop(ctx) }()

	err = c.Set(ctx, "greeting", []byte("hello from ristretto"), 0)
	if err != nil {
		fmt.Println("set failed:", err)
		return
	}

	// Ristretto buffers writes asynchronously by default; the WithSynchronousSet
	// option used by this module ensures Set is observable on the next Get.
	got, err := c.Get(ctx, "greeting")
	if err != nil {
		fmt.Println("get failed:", err)
		return
	}
	fmt.Printf("ristretto: greeting = %q\n", string(got))

	_, err = c.Get(ctx, "missing-key")
	fmt.Println("ristretto: miss error:", err)
}
