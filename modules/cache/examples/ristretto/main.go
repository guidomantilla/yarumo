package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guidomantilla/yarumo/cache"
)

func main() {
	ctx := context.Background()

	c, err := cache.BuildRistrettoCache[string]("greetings",
		cache.WithTTL(time.Minute),
	)
	if err != nil {
		fmt.Println("failed to build ristretto cache:", err)
		return
	}
	defer func() { _ = c.Stop(ctx) }()

	err = c.Set(ctx, "greeting", "hello from ristretto", 0)
	if err != nil {
		fmt.Println("set failed:", err)
		return
	}

	got, err := c.Get(ctx, "greeting")
	if err != nil {
		fmt.Println("get failed:", err)
		return
	}
	fmt.Printf("ristretto: greeting = %q\n", got)

	_, err = c.Get(ctx, "missing-key")
	fmt.Println("ristretto: miss error:", err)
}
