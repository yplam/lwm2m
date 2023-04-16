package encoding

//package encoding
//
//import (
//	"bytes"
//	"errors"
//	"github.com/plgd-dev/go-coap/v3/message"
//	"github.com/yplam/lwm2m/encoding/tlv"
//	"github.com/yplam/lwm2m/node"
//	"io"
//	"io/ioutil"
//	"reflect"
//)
//
//var (
//	ErrEmpty = errors.New("empty")
//)
//
//func EncodeMessage(t message.MediaType, node []node.Node) (io.ReadSeeker, error) {
//	if len(node) == 0 {
//		return nil, ErrEmpty
//	}
//	switch t {
//	case message.AppLwm2mTLV:
//		tlvs, err := nodesToTlvs(node)
//		if err != nil {
//			return nil, err
//		}
//		c := tlv.EncodeTLVs(tlvs)
//		//log.Printf("send: %#v", c)
//		return bytes.NewReader(c), nil
//	}
//	return nil, ErrEmpty
//}
//
//func DecodeMessage(t message.MediaType, p node.Path, msg io.ReadSeeker) ([]node.Node, error) {
//	c, err := ioutil.ReadAll(msg)
//	if err != nil {
//		return nil, err
//	}
//	switch t {
//	case message.AppLwm2mTLV:
//		tlvs, err := tlv.DecodeTLVs(c)
//		if err != nil {
//			return nil, err
//		}
//		return decodeTLVMessage(p, tlvs)
//	}
//	return nil, ErrEmpty
//}
//
//func decodeTLVMessage(p node.Path, tlvs []*tlv.Encoding) ([]node.Node, error) {
//	//logrus.Debugf("decode path %v", p.String())
//	var curInstanceId uint16 = 0
//	nodes := make([]node.Node, 0)
//	for _, item := range tlvs {
//		switch item.Type {
//		case tlv.ObjectInstance:
//			n := node.NewObjectInstance(item.Identifier)
//			if len(item.Children) > 0 {
//				p.SetObjectInstanceId(curInstanceId)
//				curInstanceId += 1
//				if nn, err := decodeTLVMessage(p, item.Children); err == nil {
//					for _, v := range nn {
//						if reflect.TypeOf(v).String() == "lwm2m.Resource" {
//							n.Resources[v.ID()] = v.(node.Resource)
//						}
//					}
//
//				}
//			}
//			nodes = append(nodes, n)
//		case tlv.SingleResource:
//			p.SetResourceId(item.Identifier)
//			if n, err := node.NewResource(p, false, item.Value); err == nil {
//				nodes = append(nodes, n)
//			}
//		case tlv.MultipleResource:
//			p.SetResourceId(item.Identifier)
//			if n, err := node.NewResource(p, true, item.Value); err == nil {
//				if nn, errr := decodeTLVMessage(p, item.Children); errr == nil {
//					for _, v := range nn {
//						if reflect.TypeOf(v).String() == "lwm2m.Resource" {
//							n.AddInstance(v.(node.Resource))
//						}
//					}
//				}
//				nodes = append(nodes, n)
//			}
//		case tlv.MultipleResourceItem:
//			p.resourceInstanceId = int32(itlv.Identifier)
//			if n, err := NewResource(p, false, itlv.Value); err == nil {
//				nodes = append(nodes, n)
//			}
//		}
//	}
//	return nodes, nil
//}
//
//func nodesToTlvs(nodes []Node) ([]*TLV, error) {
//	tlvs := make([]*TLV, 0)
//	for _, node := range nodes {
//		switch reflect.TypeOf(node).String() {
//		case "lwm2m.Resource":
//			//logrus.Debugf("encode resource")
//			if n, okay := node.(Resource); okay {
//				tlv := NewTLV(TLVSingleResource, n.id, n.data)
//				tlvs = append(tlvs, tlv)
//			}
//		}
//	}
//	return tlvs, nil
//}
