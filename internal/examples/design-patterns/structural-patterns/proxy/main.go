package main

import (
	"fmt"
)

// To use proxy and to object they must implement same methods
type Object interface {
	ObjDo(action string)
}

// TargetObject represents real objects which proxy will delegate data
type TargetObject struct {
	action string
}

// ObjDo implements IObject interface and handel's all logic
func (obj *TargetObject) ObjDo(action string) {
	// Action behavior
	fmt.Printf("I can, %s", action)
}

// ProxyObject represents proxy object with intercepts actions
type ProxyObject struct {
	targetObject *TargetObject
}

// ObjDo are implemented IObject and intercept action before send in real Object
func (p *ProxyObject) ObjDo(action string) {
	if p.targetObject == nil {
		p.targetObject = &TargetObject{}
	}
	if action == "Run" {
		p.targetObject.ObjDo(action) // Prints: I can, Run
	}
}

func main() {

	po := &ProxyObject{}
	po.ObjDo("Run")
}
