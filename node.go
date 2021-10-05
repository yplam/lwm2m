package lwm2m

import (
	"errors"
	"reflect"

	"github.com/sirupsen/logrus"
)

// A Node is the base type of lwm2m message, can be one of
// Object, ObjectInstance, Resource
// One lwm2m message package may contain one or more Node
type Node interface {
	ID() uint16
}

var (
	ErrPathNotMatch = errors.New("wrong path type")
	ErrNodeNotFound = errors.New("node not found")
)

func NodeGetAllResources(nodes []Node, parentPath Path) (map[Path]*Resource, error) {
	values := make(map[Path]*Resource)
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*lwm2m.Resource":
			logrus.Debug("resource")
			if !parentPath.IsResourceInstance() {
				continue
			}
			if n, okay := node.(*Resource); okay {
				logrus.Infof("resource %v", n.Id)
				values[parentPath] = n
			}
		case "*lwm2m.Object":
			logrus.Debug("object")
			if !parentPath.IsRoot() {
				continue
			}
			if n, okay := node.(*Object); okay {
				for _, oi := range n.Instances {
					for _, or := range oi.Resources {
						p := NewResourcePath(n.Id, oi.Id, or.Id)
						values[p] = or
					}
				}
				logrus.Infof("object (%v)", n.Id)
			}
		case "*lwm2m.ObjectInstance":
			logrus.Debug("ObjectInstance")
			if !parentPath.IsObject() {
				continue
			}
			objID, err := parentPath.ObjectId()
			if err != nil {
				continue
			}
			if n, okay := node.(*ObjectInstance); okay {
				for _, or := range n.Resources {
					logrus.Infof("object instance resource %v", or.Id)
					p := NewResourcePath(objID, n.Id, or.Id)
					values[p] = or
				}
				logrus.Infof("object instance %v", n.Id)
			}
		default:
			logrus.Warnf("unhandle node type (%v)", reflect.TypeOf(node).String())
		}
	}
	if len(values) == 0 {
		return nil, ErrNodeNotFound
	}
	return values, nil
}

func NodeGetResourceByPath(nodes []Node, p Path) (*Resource, error) {
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
		case "*lwm2m.Resource":
			if n, okay := node.(*Resource); okay {
				if rid != node.ID() {
					continue
				}
				return n, nil
			}
		case "*lwm2m.Object":
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
				return ri, nil
			}
		case "*lwm2m.ObjectInstance":
			if n, okay := node.(*ObjectInstance); okay {
				if iid != node.ID() {
					continue
				}
				ri, okay := n.Resources[rid]
				if !okay {
					continue
				}
				return ri, nil
			}
		default:
			logrus.Warnf("unhandle node type (%v)", reflect.TypeOf(node).String())
		}
	}
	return nil, ErrNodeNotFound
}
