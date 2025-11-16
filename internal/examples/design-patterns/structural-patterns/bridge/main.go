package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/structural-patterns/bridge/machines"
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/structural-patterns/bridge/printers"
)

func main() {

	hpPrinter := &printers.Hp{}
	epsonPrinter := &printers.Epson{}

	macComputer := &machines.Mac{}

	macComputer.SetPrinter(hpPrinter)
	macComputer.Print()
	fmt.Println()

	macComputer.SetPrinter(epsonPrinter)
	macComputer.Print()
	fmt.Println()

	winComputer := &machines.Windows{}

	winComputer.SetPrinter(hpPrinter)
	winComputer.Print()
	fmt.Println()

	winComputer.SetPrinter(epsonPrinter)
	winComputer.Print()
	fmt.Println()
}
