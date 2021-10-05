package lwm2m

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/plgd-dev/go-coap/v2/message"
)

var (
	ErrEmpty = errors.New("empty")
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

func DecodeMessage(t message.MediaType, msg io.ReadSeeker) ([]Node, error) {
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
		return tlvsToNodes(tlvs)
	}
	return nil, ErrEmpty
}

func tlvsToNodes(tlvs []*TLV) ([]Node, error) {
	nodes := make([]Node, 0)
	for _, tlv := range tlvs {
		//logrus.Printf("%v", tlv.Type)
		//logrus.Printf("%v", tlv.Value)
		switch tlv.Type {
		case TLVObjectInstance:
			n := NewObjectInstance(tlv.Identifier)
			if len(tlv.Children) > 0 {
				if nn, err := tlvsToNodes(tlv.Children); err == nil {
					for _, v := range nn {
						//logrus.Printf("name: %#v", reflect.TypeOf(v).String())
						if reflect.TypeOf(v).String() == "*lwm2m.Resource" {
							n.Resources[v.ID()] = v.(*Resource)
						}
					}

				}
			}
			nodes = append(nodes, n)
		case TLVSingleResource:
			n := NewResource(tlv.Identifier, false)
			n.SetValue(tlv.Value)
			nodes = append(nodes, n)
		case TLVMultipleResource:
			logrus.Warnf("TLVMultipleResource not handle")
		case TLVMultipleResourceItem:
			logrus.Warnf("TLVMultipleResourceItem not handle")
		}
	}
	return nodes, nil
}

func nodesToTlvs(nodes []Node) ([]*TLV, error) {
	tlvs := make([]*TLV, 0)
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*lwm2m.Resource":
			if n, okay := node.(*Resource); okay {
				if v, okay := n.Values[0]; okay {
					tlv := NewTLV(TLVSingleResource, n.Id, v)
					tlvs = append(tlvs, tlv)
				}
			}
		}
	}
	return tlvs, nil
}
