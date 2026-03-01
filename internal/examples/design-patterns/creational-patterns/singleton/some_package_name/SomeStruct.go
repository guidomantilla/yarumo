package some_package_name

import (
	"fmt"
	"sync"
)

var _ Single = (*defaultSingle)(nil)

var once sync.Once

type Single interface {
}

type defaultSingle struct {
	Single
}

var singleInstance *defaultSingle

func GetInstance() Single {
	if singleInstance == nil {
		once.Do(
			func() {
				fmt.Println("Creating single instance now.")
				singleInstance = &defaultSingle{}
			})
	} else {
		fmt.Println("Single instance already created.")
	}

	return singleInstance
}
