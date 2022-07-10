package lwm2m

import (
	"reflect"

	"github.com/sirupsen/logrus"
)

// A Node is the base type of lwm2m message, can be one of
// Object, ObjectInstance, Resource
// One lwm2m message package may contain one or more Node
type Node interface {
	ID() uint16
	String() string
}

func NodeGetAllResources(nodes []Node, parentPath Path) (map[Path]Resource, error) {
	values := make(map[Path]Resource)
	for _, node := range nodes {
		//logrus.Infof("node type %v", reflect.TypeOf(node).String())
		switch reflect.TypeOf(node).String() {
		case "lwm2m.Resource":
			r, okay := node.(Resource)
			if !okay {
				continue
			}
			if !r.path.IsChildOfOrEq(parentPath) {
				continue
			}
			values[r.path] = r
		case "lwm2m.Object":
			//logrus.Debugf("object")
			if n, okay := node.(Object); okay {
				for _, oi := range n.Instances {
					for _, or := range oi.Resources {
						if !or.path.IsChildOfOrEq(parentPath) {
							continue
						}
						values[or.path] = or
					}
				}
				//logrus.Debugf("object (%v)", n.Id)
			}
		case "lwm2m.ObjectInstance":
			//logrus.Debugf("ObjectInstance")
			if n, okay := node.(ObjectInstance); okay {
				for _, or := range n.Resources {
					if !or.path.IsChildOfOrEq(parentPath) {
						continue
					}
					//logrus.Debugf("resource %v", or)
					values[or.path] = or
				}
				//logrus.Debugf("object instance %v", n.Id)
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

func NodeGetResourceByPath(nodes []Node, p Path) (r Resource, err error) {
	if !p.IsResource() {
		err = ErrPathNotMatch
		return
	}
	oid, err := p.ObjectId()
	if err != nil {
		err = ErrPathNotMatch
		return
	}
	iid, err := p.ObjectInstanceId()
	if err != nil {
		err = ErrPathNotMatch
		return
	}
	rid, err := p.ResourceId()
	if err != nil {
		err = ErrPathNotMatch
		return
	}
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "lwm2m.Resource":
			if n, okay := node.(Resource); okay {
				if rid != node.ID() {
					continue
				}
				r = n
				return
			}
		case "lwm2m.Object":
			if n, okay := node.(Object); okay {
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
				r = ri
				return
			}
		case "lwm2m.ObjectInstance":
			if n, okay := node.(ObjectInstance); okay {
				if iid != node.ID() {
					continue
				}
				ri, okay := n.Resources[rid]
				if !okay {
					continue
				}
				r = ri
				return
			}
		default:
			logrus.Warnf("unhandle node type (%v)", reflect.TypeOf(node).String())
		}
	}
	err = ErrNodeNotFound
	return
}
