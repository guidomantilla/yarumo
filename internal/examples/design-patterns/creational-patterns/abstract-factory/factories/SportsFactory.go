package factories

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/abstract-factory/products"
)

var _ SportsFactory = (*AdidasFactory)(nil)

var _ SportsFactory = (*NikeFactory)(nil)

type SportsFactory interface {
	MakeShoe() products.Shoe
	MakeShirt() products.Shirt
}

func GetSportsFactory(brand string) (SportsFactory, error) {
	if brand == "adidas" {
		return &AdidasFactory{}, nil
	}

	if brand == "nike" {
		return &NikeFactory{}, nil
	}

	return nil, fmt.Errorf("Wrong brand type passed")
}
