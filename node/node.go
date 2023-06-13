package node

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/yplam/lwm2m/encoding"
	"io"
	"reflect"
)

var (
	ErrEmpty                   = errors.New("empty")
	ErrNotFound                = errors.New("not found")
	ErrContentFormatNotSupport = errors.New("content format not support")
	ErrPathNotMatch            = errors.New("wrong path type")
)

// A Node is the base type of lwm2m message, can be one of
// Object, ObjectInstance, Resource
// One lwm2m message package may contain one or more Node
type Node interface {
	ID() uint16
	String() string
}

func DecodeMessage(basePath Path, msg *pool.Message) (nodes []Node, err error) {
	nodes = make([]Node, 0)
	format, err := msg.ContentFormat()
	if err != nil {
		return
	}
	content, err := msg.ReadBody()
	if err != nil {
		return
	}

	switch format {
	case message.AppLwm2mTLV:
		tlvs, errD := encoding.DecodeTlv(content)
		if errD != nil {
			err = errD
			return
		}
		return decodeTLVMessage(basePath, tlvs)
	case message.TextPlain:
		if !basePath.IsResource() {
			return nil, errors.New("only resource can be decoded from textplain content")
		}
		pt := encoding.NewPlainTextRaw(content)
		if n, err := NewResource(basePath, false); err == nil {
			if ri, err := NewResourceInstance(basePath, pt); err == nil {
				n.SetInstance(ri)
			}
			nodes = append(nodes, n)
		}
		return nodes, nil
	case message.AppOctets:
		if !basePath.IsResource() {
			return nil, errors.New("only resource can be decoded from opaque content")
		}
		om, err := encoding.NewOpaqueValue(content)
		if err != nil {
			return nil, err
		}
		if n, err := NewResource(basePath, false); err == nil {
			if ri, err := NewResourceInstance(basePath, om); err == nil {
				n.SetInstance(ri)
			}
			nodes = append(nodes, n)
		}
		return nodes, nil
	default:
		err = ErrContentFormatNotSupport
		return
	}
	return
}

func decodeTLVMessage(p Path, tlvs []*encoding.Tlv) ([]Node, error) {
	//logrus.Debugf("decode path %v", p.String())
	var curObjInstanceId uint16 = 0
	nodes := make([]Node, 0)
	for _, item := range tlvs {
		switch item.Type {
		case encoding.TlvObjectInstance:
			n := NewObjectInstance(item.Identifier)
			if len(item.Children) > 0 {
				p.SetObjectInstanceId(curObjInstanceId)
				curObjInstanceId += 1
				if nn, err := decodeTLVMessage(p, item.Children); err == nil {
					for _, v := range nn {
						if reflect.TypeOf(v).String() == "*node.Resource" {
							n.Resources[v.ID()] = v.(*Resource)
						}
					}

				}
			}
			nodes = append(nodes, n)
		case encoding.TlvSingleResource:
			p.SetResourceId(item.Identifier)
			if n, err := NewResource(p, false); err == nil {
				if ri, err := NewResourceInstance(p, item); err == nil {
					n.SetInstance(ri)
				}
				nodes = append(nodes, n)
			}
		case encoding.TlvMultipleResource:
			p.SetResourceId(item.Identifier)
			if n, err := NewResource(p, true); err == nil {
				if nn, err := decodeTLVMessage(p, item.Children); err == nil {
					for _, v := range nn {
						if reflect.TypeOf(v).String() == "*node.ResourceInstance" {
							n.SetInstance(v.(*ResourceInstance))
						}
					}
				} else {
					fmt.Printf("decode tlv error %v", err)
				}
				nodes = append(nodes, n)
			}
		case encoding.TlvMultipleResourceItem:
			p.resourceInstanceId = int32(item.Identifier)
			if ri, err := NewResourceInstance(p, item); err == nil {
				nodes = append(nodes, ri)
			}
		default:

		}
	}
	return nodes, nil
}

func EncodeMessage(t message.MediaType, node []Node) (io.ReadSeeker, error) {
	if len(node) == 0 {
		return nil, ErrEmpty
	}
	switch t {
	case message.AppLwm2mTLV:
		tlvs, err := encodeTLVMessage(node)
		if err != nil {
			return nil, err
		}
		c := encoding.EncodeTlv(tlvs)
		return bytes.NewReader(c), nil
	case message.TextPlain:
		text, err := encodeTextMessage(node)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(text.Raw()), nil
	case message.AppOctets:
		om, err := encodeOpaqueMessage(node)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(om.Raw()), nil
	}
	return nil, ErrEmpty
}

func encodeTextMessage(node []Node) (*encoding.PlainTextValue, error) {
	if len(node) != 1 {
		return nil, errors.New("cannot encode multiple values as PlainText")
	}
	if rr, ok := node[0].(*Resource); ok {
		if rr.InstanceCount() != 1 {
			return nil, errors.New("only single resource can be encoded as plaintext")
		}
		ri, err := rr.GetInstance(0)
		if err != nil {
			return nil, err
		}
		vv := ri.Value()
		pt, err := encoding.NewPlainTextValue(vv)
		if err != nil {
			return nil, err
		}
		return pt, nil
	}
	return nil, errors.New("only single resource can be encoded as plaintext")
}

func encodeOpaqueMessage(node []Node) (*encoding.OpaqueValue, error) {
	if len(node) != 1 {
		return nil, errors.New("cannot encode multiple values as opaque octet stream")
	}
	if rr, ok := node[0].(*Resource); ok {
		if rr.InstanceCount() != 1 {
			return nil, errors.New("only single resource can be encoded as octet stream")
		}
		ri, err := rr.GetInstance(0)
		if err != nil {
			return nil, err
		}
		om, err := encoding.NewOpaqueValue(ri.Value())
		if err != nil {
			return nil, err
		}
		return om, nil
	}
	return nil, errors.New("only single resource can be encoded as octet stream")
}

func encodeTLVMessage(nodes []Node) ([]*encoding.Tlv, error) {
	tlvs := make([]*encoding.Tlv, 0)
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*node.Resource":
			if n, okay := node.(*Resource); okay {
				if n.isMultiple {
					tlv := encoding.NewTlv(encoding.TlvMultipleResource, n.id, []byte{})
					for _, ri := range n.instances {
						child := encoding.NewTlv(encoding.TlvMultipleResourceItem, ri.id, ri.Data().Raw())
						tlv.Children = append(tlv.Children, child)
					}
					tlvs = append(tlvs, tlv)
				} else {
					tlv := encoding.NewTlv(encoding.TlvSingleResource, n.id, n.Data().Raw())
					tlvs = append(tlvs, tlv)
				}
			}
		case "*node.Object":
			if n, okay := node.(*Object); okay {
				for _, oi := range n.Instances {
					if eoi, err := encodeTLVMessage([]Node{oi}); err == nil {
						tlvs = append(tlvs, eoi...)
					}
				}
			}
		case "*node.ObjectInstance":
			if n, okay := node.(*ObjectInstance); okay {
				tlv := encoding.NewTlv(encoding.TlvObjectInstance, n.Id, []byte{})
				for _, ri := range n.Resources {
					if eri, err := encodeTLVMessage([]Node{ri}); err == nil {
						tlv.Children = append(tlv.Children, eri...)
					}
				}
				tlvs = append(tlvs, tlv)
			}
		case "*node.ResourceInstance":
			if n, okay := node.(*ResourceInstance); okay {
				tlv := encoding.NewTlv(encoding.TlvMultipleResourceItem, n.id, n.Data().Raw())
				tlvs = append(tlvs, tlv)
			}
		}

	}
	return tlvs, nil
}

func GetAllResources(nodes []Node, parentPath Path) (map[Path]*Resource, error) {
	values := make(map[Path]*Resource)
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*node.Resource":
			r, okay := node.(*Resource)
			if !okay {
				continue
			}
			if !r.path.IsChildOfOrEq(parentPath) {
				continue
			}
			values[r.path] = r
		case "*node.Object":
			if n, okay := node.(*Object); okay {
				for _, oi := range n.Instances {
					for _, or := range oi.Resources {
						if !or.path.IsChildOfOrEq(parentPath) {
							continue
						}
						values[or.path] = or
					}
				}
			}
		case "*node.ObjectInstance":
			if n, okay := node.(*ObjectInstance); okay {
				for _, or := range n.Resources {
					if !or.path.IsChildOfOrEq(parentPath) {
						continue
					}
					values[or.path] = or
				}
			}
		default:
			fmt.Printf("unhandle node type (%v)\n", reflect.TypeOf(node).String())
		}
	}
	if len(values) == 0 {
		return nil, ErrNotFound
	}
	return values, nil
}

func GetObjectByPath(nodes []Node, p Path) (o *Object, err error) {
	if !p.IsObject() {
		err = ErrPathNotMatch
		return
	}
	oid, err := p.ObjectId()
	if err != nil {
		err = ErrPathNotMatch
		return
	}
	o = NewObject(oid)
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*node.Object":
			if n, okay := node.(*Object); okay {
				if oid != node.ID() {
					continue
				}
				o = n
				return
			}
		case "*node.ObjectInstance":
			if n, okay := node.(*ObjectInstance); okay {
				o.Instances[n.ID()] = n
			}
		default:
			fmt.Printf("unhandle node type (%v)\n", reflect.TypeOf(node).String())
		}
	}
	if len(o.Instances) == 0 {
		err = ErrNotFound
	}
	return
}

func GetResourceByPath(nodes []Node, p Path) (r *Resource, err error) {
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
		case "*node.Resource":
			if n, okay := node.(*Resource); okay {
				if rid != node.ID() {
					continue
				}
				r = n
				return
			}
		case "*node.Object":
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
				r = ri
				return
			}
		case "*node.ObjectInstance":
			if n, okay := node.(*ObjectInstance); okay {
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
			fmt.Printf("unhandle node type (%v)\n", reflect.TypeOf(node).String())
		}
	}
	err = ErrNotFound
	return
}
