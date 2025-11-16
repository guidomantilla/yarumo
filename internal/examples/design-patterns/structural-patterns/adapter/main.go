package main

import (
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/structural-patterns/adapter/adapters"
	"github.com/guidomantilla/yarumo/internal/examples/design-patterns/structural-patterns/adapter/machines"
)

func main() {

	client := &Client{}
	mac := &machines.Mac{}

	client.InsertLightningConnectorIntoComputer(mac)

	windowsMachine := &machines.Windows{}
	windowsMachineAdapter := &adapters.WindowsAdapter{
		WindowMachine: windowsMachine,
	}

	client.InsertLightningConnectorIntoComputer(windowsMachineAdapter)
}
