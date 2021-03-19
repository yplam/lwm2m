package encoding

import (
	"bytes"
	"errors"
	"github.com/plgd-dev/go-coap/v2/message"
	"io"
	"io/ioutil"
	"lwm2m/model"
	"reflect"
)

var (
	ErrEmpty = errors.New("empty")
)

func EncodeMessage(t message.MediaType, node []model.Node) (io.ReadSeeker, error) {
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

func DecodeMessage(t message.MediaType, msg io.ReadSeeker) ([]model.Node, error) {
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

func tlvsToNodes(tlvs []*TLV) ([]model.Node, error) {
	nodes := make([]model.Node, 0)
	for _, tlv := range tlvs {
		//log.Printf("%v", tlv.Type)
		//log.Printf("%v", tlv.Value)
		switch tlv.Type {
		case ObjectInstance:
			n := model.NewObjectInstance(tlv.Identifier)
			if len(tlv.Children) > 0 {
				if nn, err := tlvsToNodes(tlv.Children); err == nil {
					for _, v := range nn {
						//log.Printf("name: %#v",reflect.TypeOf(v).String())
						if reflect.TypeOf(v).String() == "*model.Resource" {
							n.Resources[v.ID()] = v.(*model.Resource)
						}
					}

				}
			}
			nodes = append(nodes, n)
		case SingleResource:
			n := model.NewResource(tlv.Identifier, false)
			n.SetValue(tlv.Value)
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func nodesToTlvs(nodes []model.Node) ([]*TLV, error) {
	tlvs := make([]*TLV, 0)
	for _, node := range nodes {
		switch reflect.TypeOf(node).String() {
		case "*model.Resource":
			n := node.(*model.Resource)
			tlv := NewTLV(SingleResource, n.Id, n.Values[0])
			tlvs = append(tlvs, tlv)
		}
	}
	return tlvs, nil
}
