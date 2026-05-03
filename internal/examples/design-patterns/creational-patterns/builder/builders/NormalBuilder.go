package builders

import (
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/builder/products"
)

type NormalBuilder struct {
	windowType string
	doorType   string
	floor      int
}

func NewNormalBuilder() *NormalBuilder {
	return &NormalBuilder{}
}

func (b *NormalBuilder) SetWindowType() {
	b.windowType = "Wooden Window"
}

func (b *NormalBuilder) SetDoorType() {
	b.doorType = "Wooden Door"
}

func (b *NormalBuilder) SetNumFloor() {
	b.floor = 2
}

func (b *NormalBuilder) GetHouse() *products.House {
	return &products.House{
		DoorType:   b.doorType,
		WindowType: b.windowType,
		Floor:      b.floor,
	}
}
