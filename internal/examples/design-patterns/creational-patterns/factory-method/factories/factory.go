package factories

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/creational-patterns/factory-method/products"
)

func GetGun(gunType string) (products.Gun, error) {
	if gunType == "ak47" {
		return products.NewAk47(), nil
	}
	if gunType == "musket" {
		return products.NewMusket(), nil
	}
	return nil, fmt.Errorf("Wrong gun type passed")
}
