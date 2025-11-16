package printers

import (
	"fmt"
)

var _ Printer = (*Epson)(nil)

var _ Printer = (*Hp)(nil)

type Printer interface {
	PrintFile()
}

//

type Epson struct {
}

func (p *Epson) PrintFile() {
	fmt.Println("Printing by a EPSON Printer")
}

//

type Hp struct {
}

func (p *Hp) PrintFile() {
	fmt.Println("Printing by a HP Printer")
}
