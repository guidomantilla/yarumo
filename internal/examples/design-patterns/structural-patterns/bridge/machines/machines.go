package machines

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/structural-patterns/bridge/printers"
)

type Computer interface {
	Print()
	SetPrinter(printers.Printer)
}

//

type Mac struct {
	printer printers.Printer
}

func (m *Mac) Print() {
	fmt.Println("Print request for mac")
	m.printer.PrintFile()
}

func (m *Mac) SetPrinter(p printers.Printer) {
	m.printer = p
}

//

type Windows struct {
	printer printers.Printer
}

func (w *Windows) Print() {
	fmt.Println("Print request for windows")
	w.printer.PrintFile()
}

func (w *Windows) SetPrinter(p printers.Printer) {
	w.printer = p
}
