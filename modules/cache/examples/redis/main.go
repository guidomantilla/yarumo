package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/guidomantilla/yarumo/cache"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	server, err := miniredis.Run()
	if err != nil {
		fmt.Println("failed to start miniredis:", err)
		return
	}
	defer server.Close()

	ctx := context.Background()

	c, err := cache.BuildRedisCache[user]("users",
		cache.WithRedisAddr(server.Addr()),
		cache.WithTTL(time.Minute),
	)
	if err != nil {
		fmt.Println("failed to build redis cache:", err)
		return
	}
	defer func() { _ = c.Stop(ctx) }()

	err = c.Set(ctx, "42", user{ID: 42, Name: "Ana"}, 0)
	if err != nil {
		fmt.Println("set failed:", err)
		return
	}

	got, err := c.Get(ctx, "42")
	if err != nil {
		fmt.Println("get failed:", err)
		return
	}
	fmt.Printf("redis: user(42) = %+v\n", got)

	_, err = c.Get(ctx, "missing")
	fmt.Println("redis: miss error:", err)
}
