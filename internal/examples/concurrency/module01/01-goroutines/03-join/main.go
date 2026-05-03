package main

import (
	"fmt"
	"sync"
)

func main() {
	// program to print the value as 1
	// deterministically.
	var wg sync.WaitGroup
	var data int

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("before operation")
		data++
		fmt.Println("after operation")
	}()
	wg.Wait()
	fmt.Printf("the value of data is %v\n", data)

	fmt.Println("Done..")
}
