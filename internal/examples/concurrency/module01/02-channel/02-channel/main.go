package main

import "fmt"

func main() {
	ch := make(chan int)
	go func() {
		for i := 0; i < 6; i++ {
			// send iterator over channel
			fmt.Printf("Sending: %d\n", i)
			ch <- i
		}
		close(ch)
	}()
	// range over channel to recv values
	// This blocks until the computed value is written into the channel
	// The for is break when close() is called
	for i := range ch {
		fmt.Printf("Received: %v\n", i)
	}
}
