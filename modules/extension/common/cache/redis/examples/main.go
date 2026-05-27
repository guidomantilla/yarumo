package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/extension/common/cache/redis"
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

	c := redis.NewRedisCache[user]("users",
		redis.WithAddr(server.Addr()),
		redis.WithTTL(time.Minute),
	)

	errCh := make(chan error, 1)
	closeFn, buildErr := lifecycle.Build(ctx, c, errCh)
	if buildErr != nil {
		fmt.Println("lifecycle.Build:", buildErr)
		return
	}
	defer closeFn(ctx, 5*time.Second)

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
