package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/factory-method/factories"
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/factory-method/products"
)

func main() {
	ak47, _ := factories.GetGun("ak47")
	musket, _ := factories.GetGun("musket")

	printDetails(ak47)
	printDetails(musket)
}

func printDetails(g products.Gun) {
	fmt.Printf("Gun: %s", g.GetName())
	fmt.Println()
	fmt.Printf("Power: %d", g.GetPower())
	fmt.Println()
}
