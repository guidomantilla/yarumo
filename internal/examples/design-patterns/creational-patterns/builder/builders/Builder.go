package builders

import (
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/builder/products"
)

var _ Builder = (*IglooBuilder)(nil)

var _ Builder = (*NormalBuilder)(nil)

type Builder interface {
	SetWindowType()
	SetDoorType()
	SetNumFloor()
	GetHouse() *products.House
}

func GetBuilder(builderType string) Builder {
	if builderType == "normal" {
		return NewNormalBuilder()
	}

	if builderType == "igloo" {
		return NewIglooBuilder()
	}
	return nil
}
