package machines

import (
	"fmt"
)

var _ Computer = (*Mac)(nil)

type Computer interface {
	InsertIntoLightningPort()
}

//

type Mac struct {
}

func (m *Mac) InsertIntoLightningPort() {
	fmt.Println("Lightning connector is plugged into mac machine.")
}

//

type Windows struct{}

func (w *Windows) InsertIntoUSBPort() {
	fmt.Println("USB connector is plugged into windows machine.")
}
