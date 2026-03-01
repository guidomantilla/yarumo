package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/abstract-factory/factories"
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/abstract-factory/products"
)

func main() {
	adidasFactory, _ := factories.GetSportsFactory("adidas")
	nikeFactory, _ := factories.GetSportsFactory("nike")

	nikeShoe := nikeFactory.MakeShoe()
	nikeShirt := nikeFactory.MakeShirt()

	adidasShoe := adidasFactory.MakeShoe()
	adidasShirt := adidasFactory.MakeShirt()

	printShoeDetails(nikeShoe)
	printShirtDetails(nikeShirt)

	printShoeDetails(adidasShoe)
	printShirtDetails(adidasShirt)
}

func printShoeDetails(s products.Shoe) {
	fmt.Printf("Logo: %s", s.GetLogo())
	fmt.Println()
	fmt.Printf("Size: %d", s.GetSize())
	fmt.Println()
}

func printShirtDetails(s products.Shirt) {
	fmt.Printf("Logo: %s", s.GetLogo())
	fmt.Println()
	fmt.Printf("Size: %d", s.GetSize())
	fmt.Println()
}
