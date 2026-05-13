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
		cache.WithBackend(cache.BackendBigcache),
		cache.WithTTL(2*time.Minute),
		cache.WithBigcacheCapacity(64, 10*time.Minute, 5*time.Minute, 16, 2048),
	)
	if err != nil {
		fmt.Println("failed to create bigcache:", err)
		return
	}
	defer func() { _ = c.Stop(ctx) }()

	err = c.Set(ctx, "payload", []byte("hello from bigcache"), 0)
	if err != nil {
		fmt.Println("set failed:", err)
		return
	}

	got, err := c.Get(ctx, "payload")
	if err != nil {
		fmt.Println("get failed:", err)
		return
	}
	fmt.Printf("bigcache: payload = %q\n", string(got))

	_, err = c.Get(ctx, "absent")
	fmt.Println("bigcache: miss error:", err)
}
