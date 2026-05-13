package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guidomantilla/yarumo/cache"
)

func main() {
	ctx := context.Background()

	c, stopFn, err := cache.BuildCache[string, []byte](ctx, "go-cache-example",
		cache.WithBackend(cache.BackendGoCache),
		cache.WithGoCacheCapacity(time.Minute, 2*time.Minute),
		cache.WithSlog(),
	)
	if err != nil {
		fmt.Println("failed to build go-cache:", err)
		return
	}
	defer stopFn(ctx, 5*time.Second)

	err = c.Set(ctx, "user:42", []byte("alice"), 30*time.Second)
	if err != nil {
		fmt.Println("set failed:", err)
		return
	}

	got, err := c.Get(ctx, "user:42")
	if err != nil {
		fmt.Println("get failed:", err)
		return
	}
	fmt.Printf("go-cache: user:42 = %q\n", string(got))

	if c.Has(ctx, "user:42") {
		fmt.Println("go-cache: Has(user:42)=true")
	}

	err = c.Delete(ctx, "user:42")
	if err != nil {
		fmt.Println("delete failed:", err)
		return
	}

	_, err = c.Get(ctx, "user:42")
	fmt.Println("go-cache: miss after delete:", err)
}
