package model

import (
	"errors"
	"lwm2m/path"
	"reflect"
)

// A Node is the base type of lwm2m message, can be one of Object, ObjectInstance, Resource
// One lwm2m message package may contain one or more Node
type Node interface {
	ID() uint16
}

var (
	ErrPathNotMatch = errors.New("wrong path type")
	ErrNodeNotFound = errors.New("node not found")
)

func NodeGetSingleResourceValue(nodes []Node, p *path.Path) ([]byte, error) {
	if !p.IsResource() {
		return nil, ErrPathNotMatch
	}
	oid, err := p.ObjectId()
	if err != nil {
		return nil, ErrPathNotMatch
	}
	iid, err := p.ObjectInstanceId()
	if err != nil {
		return nil, ErrPathNotMatch
	}
	rid, err := p.ResourceId()
	if err != nil {
		return nil, ErrPathNotMatch
	}
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*model.Resource":
			if n, okay := node.(*Resource); okay {
				if rid != node.ID() {
					continue
				}
				if v, okay := n.Values[0]; okay {
					return v, nil
				}
			}
		case "*model.Object":
			if n, okay := node.(*Object); okay {
				if oid != node.ID() {
					continue
				}
				oi, okay := n.Instances[iid]
				if !okay {
					continue
				}
				ri, okay := oi.Resources[rid]
				if !okay {
					continue
				}
				if v, okay := ri.Values[0]; okay {
					return v, nil
				}
			}
		case "*model.ObjectInstance":
			if n, okay := node.(*ObjectInstance); okay {
				if iid != node.ID() {
					continue
				}
				ri, okay := n.Resources[rid]
				if !okay {
					continue
				}
				if v, okay := ri.Values[0]; okay {
					return v, nil
				}
			}
		}
	}
	return nil, ErrNodeNotFound
}
