package core

import (
	"fmt"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/yplam/lwm2m/node"
)

type ObserveFunc func(d *Device, p node.Path, notify []node.Node)
type ObserveObjectFunc func(d *Device, p node.Path, notify *node.Object)
type ObserveResourceFunc func(d *Device, p node.Path, notify *node.Resource)

type Observation struct {
	o  mux.Observation
	cb ObserveFunc
}

func wrapObserveResourceFunc(f ObserveResourceFunc) ObserveFunc {
	return func(d *Device, p node.Path, notify []node.Node) {
		if data, err := node.GetResourceByPath(notify, p); err == nil {
			f(d, p, data)
		}
	}
}

func wrapObserveObjectFunc(f ObserveObjectFunc) ObserveFunc {
	return func(d *Device, p node.Path, notify []node.Node) {
		if data, err := node.GetObjectByPath(notify, p); err == nil {
			f(d, p, data)
		} else {
			fmt.Printf("obj err %v\n", err)
		}
	}
}
