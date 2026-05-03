package adapters

import (
	"fmt"

	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/structural-patterns/adapter/machines"
)

var _ machines.Computer = (*WindowsAdapter)(nil)

type WindowsAdapter struct {
	WindowMachine *machines.Windows
}

func (w *WindowsAdapter) InsertIntoLightningPort() {
	fmt.Println("Adapter converts Lightning signal to USB.")
	w.WindowMachine.InsertIntoUSBPort()
}
