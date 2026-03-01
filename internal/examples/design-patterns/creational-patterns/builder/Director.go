package main

import (
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/builder/builders"
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/builder/products"
)

type Director struct {
	builder builders.Builder
}

func NewDirector(b builders.Builder) *Director {
	return &Director{
		builder: b,
	}
}

func (d *Director) setBuilder(b builders.Builder) {
	d.builder = b
}

func (d *Director) buildHouse() *products.House {
	d.builder.SetDoorType()
	d.builder.SetWindowType()
	d.builder.SetNumFloor()
	return d.builder.GetHouse()
}
