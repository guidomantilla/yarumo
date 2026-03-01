package factories

import (
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/abstract-factory/products"
)

type AdidasFactory struct {
}

func (n *AdidasFactory) MakeShoe() products.Shoe {
	return &products.AdidasShoe{
		Shoe: &products.AbstractShoe{
			Logo: "nike",
			Size: 14,
		},
	}
}

func (n *AdidasFactory) MakeShirt() products.Shirt {
	return &products.AdidasShirt{
		Shirt: &products.AbstractShirt{
			Logo: "nike",
			Size: 14,
		},
	}
}
