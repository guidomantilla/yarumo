package main

import (
	"fmt"
)

func main() {
	ch := make(chan int)

	go func(a, b int) {
		c := a + b
		ch <- c
	}(1, 2)
	// get the value computed from goroutine
	c := <-ch // This blocks until the computed value is written into the channel
	fmt.Printf("computed value %v\n", c)
}
