package lwm2m

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/sirupsen/logrus"
)

func EncodeMessage(t message.MediaType, node []Node) (io.ReadSeeker, error) {
	if len(node) == 0 {
		return nil, ErrEmpty
	}
	switch t {
	case message.AppLwm2mTLV:
		tlvs, err := nodesToTlvs(node)
		if err != nil {
			return nil, err
		}
		c := EncodeTLVs(tlvs)
		//log.Printf("send: %#v", c)
		return bytes.NewReader(c), nil
	}
	return nil, ErrEmpty
}

func DecodeMessage(t message.MediaType, p Path, msg io.ReadSeeker) ([]Node, error) {
	c, err := ioutil.ReadAll(msg)
	if err != nil {
		return nil, err
	}
	switch t {
	case message.AppLwm2mTLV:
		tlvs, err := DecodeTLVs(c)
		if err != nil {
			return nil, err
		}
		return decodeTLVMessage(p, tlvs)
	}
	return nil, ErrEmpty
}

func decodeTLVMessage(p Path, tlvs []*TLV) ([]Node, error) {
	logrus.Debugf("decode path %v", p.String())
	var curInstanceId int32 = 0
	nodes := make([]Node, 0)
	for _, tlv := range tlvs {
		switch tlv.Type {
		case TLVObjectInstance:
			logrus.Debugf("decode tlv object instance")
			n := NewObjectInstance(tlv.Identifier)
			if len(tlv.Children) > 0 {
				p.objectInstanceId = curInstanceId
				curInstanceId += 1
				if nn, err := decodeTLVMessage(p, tlv.Children); err == nil {
					for _, v := range nn {
						if reflect.TypeOf(v).String() == "*lwm2m.Resource" {
							n.Resources[v.ID()] = v.(*Resource)
						}
					}

				}
			}
			nodes = append(nodes, n)
		case TLVSingleResource:
			p.resourceId = int32(tlv.Identifier)
			if n, err := NewResource(p, false, tlv.Value); err == nil {
				nodes = append(nodes, n)
			}
		case TLVMultipleResource:
			p.resourceId = int32(tlv.Identifier)
			if n, err := NewResource(p, true, tlv.Value); err == nil {
				if nn, errr := decodeTLVMessage(p, tlv.Children); errr == nil {
					for _, v := range nn {
						if reflect.TypeOf(v).String() == "*lwm2m.Resource" {
							n.addInstance(v.(*Resource))
						}
					}
				}
				nodes = append(nodes, n)
			}
		case TLVMultipleResourceItem:
			p.resourceInstanceId = int32(tlv.Identifier)
			if n, err := NewResource(p, false, tlv.Value); err == nil {
				nodes = append(nodes, n)
			}
		}
	}
	return nodes, nil
}

func nodesToTlvs(nodes []Node) ([]*TLV, error) {
	tlvs := make([]*TLV, 0)
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*lwm2m.Resource":
			logrus.Debugf("encode resource")
			if n, okay := node.(*Resource); okay {
				tlv := NewTLV(TLVSingleResource, n.id, n.data)
				tlvs = append(tlvs, tlv)
			}
		}
	}
	return tlvs, nil
}
