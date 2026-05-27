package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/extension/common/cache/ristretto"
)

func main() {
	ctx := context.Background()

	c := ristretto.NewRistrettoCache[string]("greetings",
		ristretto.WithTTL(time.Minute),
	)

	errCh := make(chan error, 1)
	closeFn, buildErr := lifecycle.Build(ctx, c, errCh)
	if buildErr != nil {
		fmt.Println("lifecycle.Build:", buildErr)
		return
	}
	defer closeFn(ctx, 5*time.Second)

	err := c.Set(ctx, "greeting", "hello from ristretto", 0)
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
