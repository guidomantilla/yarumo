package main

import (
	"fmt"
)

var _ Node = (*File)(nil)

var _ Node = (*Folder)(nil)

type Node interface {
	print(string)
	clone() Node
}

//

type File struct {
	name string
}

func (f *File) print(indentation string) {
	fmt.Println(indentation + f.name)
}

func (f *File) clone() Node {
	return &File{name: f.name + "_clone"}
}

//

type Folder struct {
	children []Node
	name     string
}

func (f *Folder) print(indentation string) {
	fmt.Println(indentation + f.name)
	for _, i := range f.children {
		i.print(indentation + indentation)
	}
}

func (f *Folder) clone() Node {
	cloneFolder := &Folder{name: f.name + "_clone"}
	var tempChildren []Node
	for _, i := range f.children {
		copy := i.clone()
		tempChildren = append(tempChildren, copy)
	}
	cloneFolder.children = tempChildren
	return cloneFolder
}
