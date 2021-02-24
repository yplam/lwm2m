package lwm2m

import (
	"encoding/binary"
	"errors"
	"math"
)

type TLVType byte

var (
	ObjectInstance       TLVType = 0 // Object Instance in which case the Value contains one or more Resource TLVs
	MultipleResourceItem TLVType = 1 // Resource Instance with Value for use	within a multiple Resource TLV
	MultipleResource     TLVType = 2 // multiple Resource, in which case the	Value contains one or more Resource Instance TLVs
	SingleResource       TLVType = 3 // Resource with Value
)

type TLV struct {
	Type       TLVType
	Identifier uint16
	Value      []byte
	Length     uint32
	Children   []*TLV
}

func (t TLV) StringVal() (string, error) {
	return string(t.Value), nil
}

func (t TLV) Uint8Val() (uint8, error) {
	if t.Length != 1 {
		return 0, errors.New("invalid length")
	}
	return uint8(t.Value[0]), nil
}

func (t TLV) Int8Val() (int8, error) {
	if t.Length != 1 {
		return 0, errors.New("invalid length")
	}
	return int8(t.Value[0]), nil
}

func (t TLV) Int16Val() (int16, error) {
	if t.Length != 2 {
		return 0, errors.New("invalid length")
	}
	return int16(t.Value[0])*256 + int16(t.Value[1]), nil
}

func (t TLV) Uint16Val() (uint16, error) {
	if t.Length != 2 {
		return 0, errors.New("invalid length")
	}
	return uint16(t.Value[0])*256 + uint16(t.Value[1]), nil
}

func (t TLV) Int32Val() (int32, error) {
	v, err := t.Uint32Val()
	return int32(v), err
}

func (t TLV) Uint32Val() (uint32, error) {
	if t.Length != 4 {
		return 0, errors.New("invalid length")
	}
	return binary.BigEndian.Uint32(t.Value), nil
}

func (t TLV) Float32Val() (float32, error) {
	if t.Length != 4 {
		return 0, errors.New("invalid length")
	}
	return math.Float32frombits(binary.BigEndian.Uint32(t.Value)), nil
}

func (t TLV) Float64Val() (float64, error) {
	if t.Length != 8 {
		return 0, errors.New("invalid length")
	}
	return math.Float64frombits(binary.BigEndian.Uint64(t.Value)), nil
}

func tlvLen(d []byte) (l uint32) {
	for _, v := range d {
		l = l<<8 + uint32(v)
	}
	return
}

func tlvUnmarshal(data []byte) (*TLV, uint32, error) {
	dlen := uint32(len(data))
	if dlen == 0 {
		return nil, 0, nil
	}
	if dlen < 2 {
		return nil, 0, errors.New("no enough data")
	}
	//log.Printf("%#v", data)
	t := TLV{}
	t.Type = TLVType((data[0] >> 6) & 0x03)
	var offset uint32 = 1
	idLen := uint32((data[0]>>5)&0x01 + 1)
	if dlen < offset+idLen {
		return nil, 0, errors.New("no enough data id")
	}
	t.Identifier = uint16(tlvLen(data[offset : offset+idLen]))
	offset = offset + idLen
	lenType := uint32((data[0] >> 3) & 0x03)
	if dlen < offset+lenType {
		return nil, 0, errors.New("no enough data dlen")
	}
	if lenType == 0 {
		t.Length = uint32(data[0] & 0x07)
	} else {
		t.Length = tlvLen(data[offset : offset+lenType])
		offset = lenType + offset
	}
	if dlen < offset+t.Length {
		return nil, 0, errors.New("no enough data len")
	}
	t.Value = data[offset : offset+t.Length]
	offset = offset + t.Length
	if t.Type == ObjectInstance || t.Type == MultipleResource {
		//log.Printf("decode children %v, %#v", t.Length, t.Value)
		c, err := DecodeTLVs(t.Value)
		if err != nil {
			return nil, 0, err
		}
		t.Children = c
	}
	//log.Printf("total : %v, %#v", offset, t)
	return &t, offset, nil
}

func DecodeTLVs(data []byte) ([]*TLV, error) {
	var tlvs []*TLV
	var offset uint32 = 0
	var err error
	for {
		t, o, err := tlvUnmarshal(data[offset:])
		offset += o
		if err != nil {
			return nil, err
		}
		if t == nil {
			break
		}
		tlvs = append(tlvs, t)
	}
	return tlvs, err
}
