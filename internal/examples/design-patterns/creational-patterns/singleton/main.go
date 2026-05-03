package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/singleton/some_package_name"
)

func main() {

	for i := 0; i < 30; i++ {
		go some_package_name.GetInstance()
	}

	// Scanln is similar to Scan, but stops scanning at a newline and
	// after the final item there must be a newline or EOF.
	fmt.Scanln()
}
