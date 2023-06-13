package encoding

import (
	"bytes"
	"encoding/binary"
	"math"
)

type TlvType byte

var (
	TlvObjectInstance       TlvType = 0 // Object Instance in which case the Value contains one or more Resource TLVs
	TlvMultipleResourceItem TlvType = 1 // Resource Instance with Value for use	within a multiple Resource Encoding
	TlvMultipleResource     TlvType = 2 // multiple Resource, in which case the	Value contains one or more Resource Instance TLVs
	TlvSingleResource       TlvType = 3 // Resource with Value
)

// Encoding is Type-Length-Value encoding format for lwm2m
//
// -------------------------------------------------------------------------------
// + Type       + 8 bits    + Bit 7-6: Indicates the type of Identifier.
// +            +           + 00 = Object Instance in which case the Value
// +            +           +      contains one or more Resource TLVs
// +            +           + 01 = Resource Instance with Value for use
// +            +           +      within a multiple Resource Encoding
// +            +           + 10 = multiple Resource, in which case the Value
// +            +           +       contains one or more Resource Instance TLVs
// +            +           + 11 = Resource with Value
// +            +           +------------------------------------------------------
// +            +           + Bit 5: Indicates the Length of the Identifier.
// +            +           + 0 = The Identifier field of this Encoding is 8 bits long
// +            +           + 1 =The Identifier field of this Encoding is 16 bits long
// +            +           +------------------------------------------------------
// +            +           + Bit 4-3: Indicates the type of Length.
// +            +           + 00 = No length field, the value immediately
// +            +           +      follows the Identifier field in is of the length
// +            +           +      indicated by Bits 2-0 of this field
// +            +           + 01 = The Length field is 8-bits and Bits 2-0 MUST be ignored
// +            +           + 10 = The Length field is 16-bits and Bits 2-0 MUST be ignored
// +            +           + 11 = The Length field is 24-bits and Bits 2-0 MUST be ignored
// +            +           +------------------------------------------------------
// +            +           + Bits 2-0: A 3-bit unsigned integer indicating
// +            +           + the Length of the Value.
// +------------+-----------+------------------------------------------------------
// + Identifier + 8/16 bits + The Object Instance, Resource, or Resource
// +            +           + Instance ID as indicated by the Type field.
// +------------+-----------+------------------------------------------------------
// + Length     + 0-24 bits + The Length of the following field in bytes.
// +------------+-----------+------------------------------------------------------
// + Value      + bytes     + Value of the tag.
// -----------------------------------------------------------------------------
type Tlv struct {
	Type       TlvType
	Identifier uint16
	Value      []byte
	Length     uint32
	Children   []*Tlv
}

func (t *Tlv) UnmarshalBinary(data []byte) (err error) {
	_, err = t.Unmarshal(data)
	return
}

func (t *Tlv) Unmarshal(data []byte) (offset uint32, err error) {
	dataLen := uint32(len(data))
	if dataLen == 0 {
		return
	}
	if dataLen < 2 {
		err = ErrNotEnoughData
		return
	}
	t.Type = TlvType((data[0] >> 6) & 0x03)

	offset = 1
	idLen := uint32((data[0]>>5)&0x01 + 1)
	if dataLen < offset+idLen {
		err = ErrNotEnoughData
		return
	}
	t.Identifier = uint16(tlvLen(data[offset : offset+idLen]))

	offset = offset + idLen
	lenType := uint32((data[0] >> 3) & 0x03)
	if dataLen < offset+lenType {
		err = ErrNotEnoughData
		return
	}
	if lenType == 0 {
		t.Length = uint32(data[0] & 0x07)
	} else {
		t.Length = tlvLen(data[offset : offset+lenType])
		offset = lenType + offset
	}

	if dataLen < offset+t.Length {
		err = ErrNotEnoughData
		return
	}
	t.Value = data[offset : offset+t.Length]
	offset = offset + t.Length
	if t.Type == TlvObjectInstance || t.Type == TlvMultipleResource {
		var c []*Tlv
		c, err = DecodeTlv(t.Value)
		if err != nil {
			return
		}
		t.Children = c
	}
	return offset, nil
}

func (t *Tlv) MarshalBinary() ([]byte, error) {
	if len(t.Children) > 0 {
		t.Value = EncodeTlv(t.Children)
		t.Length = uint32(len(t.Value))
	}
	data := make([]byte, 0, t.Length+6)
	// Identifier is always two bytes
	var dType byte = 0x20
	dType |= byte(t.Type << 6)
	//log.Printf("%#v", t.Length)
	if t.Length <= 7 {
		dType |= byte(t.Length)
		data = append(data, dType, byte((t.Identifier>>8)&0xFF), byte(t.Identifier&0xFF))
	} else if t.Length <= 0xFF {
		dType |= byte(0x01 << 3)
		data = append(data, dType, byte((t.Identifier>>8)&0xFF), byte(t.Identifier&0xFF),
			byte(t.Length))
	} else if t.Length <= 0xFFFF {
		dType |= byte(0x10 << 3)
		data = append(data, dType, byte((t.Identifier>>8)&0xFF), byte(t.Identifier&0xFF),
			byte((t.Length&0xFF00)>>8), byte(t.Length&0xFF))
	} else if t.Length <= 0xFFFFFF {
		dType |= byte(0x11 << 3)
		data = append(data, dType, byte((t.Identifier>>8)&0xFF), byte(t.Identifier&0xFF),
			byte((t.Length&0xFF0000)>>16), byte((t.Length&0xFF00)>>8), byte(t.Length&0xFF))
	}
	data = append(data, t.Value...)
	return data, nil
}

// marshalCap return the total size of the encoded bytes
func (t *Tlv) marshalCap() (c uint32) {
	c = 3
	var l uint32
	if len(t.Children) > 0 {
		for _, cc := range t.Children {
			c += cc.marshalCap()
		}
		l = c - 3
	} else {
		l = t.Length
	}
	if l <= 0xFF {
		c += 1
	} else if l <= 0xFFFF {
		c += 2
	} else if l <= 0xFFFFFF {
		c += 3
	}
	return
}

// UTF-8 string
func (t *Tlv) StringVal() string {
	return string(t.Value)
}

func (t *Tlv) Integer() (val int64, err error) {
	buff := bytes.NewBuffer(t.Value)
	if t.Length == 1 {
		var i1 int8
		err = binary.Read(buff, binary.BigEndian, &i1)
		val = int64(i1)
	} else if t.Length == 2 {
		var i2 int16
		err = binary.Read(buff, binary.BigEndian, &i2)
		val = int64(i2)
	} else if t.Length == 4 {
		var i4 int32
		err = binary.Read(buff, binary.BigEndian, &i4)
		val = int64(i4)
	} else if t.Length == 8 {
		var i8 int64
		err = binary.Read(buff, binary.BigEndian, &i8)
		val = i8
	} else {
		err = ErrInvalidLength
	}
	return
}

func (t *Tlv) Float() (float64, error) {
	if t.Length == 4 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(t.Value))), nil
	} else if t.Length == 8 {
		return math.Float64frombits(binary.BigEndian.Uint64(t.Value)), nil
	} else {
		return 0, ErrInvalidLength
	}
}

func (t *Tlv) Boolean() (val bool, err error) {
	if t.Length != 1 {
		return false, ErrInvalidLength
	}
	val = t.Value[0] != 0
	return
}

func (t *Tlv) Opaque() []byte {
	return t.Value
}

func (t *Tlv) Time() (int64, error) {
	if t.Length == 4 {
		return int64(binary.BigEndian.Uint32(t.Value)), nil
	} else if t.Length == 8 {
		return int64(binary.BigEndian.Uint64(t.Value)), nil
	} else {
		return 0, ErrInvalidLength
	}
}

func (t *Tlv) ObjectLink() (uint16, uint16, error) {
	if t.Length != 4 {
		return 0, 0, ErrInvalidLength
	}
	_ = t.Value[3]
	return uint16(t.Value[0])*256 + uint16(t.Value[1]),
		uint16(t.Value[2])*256 + uint16(t.Value[3]), nil
}

func (t *Tlv) Raw() []byte {
	return t.Value
}

func tlvLen(d []byte) (l uint32) {
	for _, v := range d {
		l = l<<8 + uint32(v)
	}
	return
}

func DecodeTlv(data []byte) (tlvs []*Tlv, err error) {
	var offset uint32 = 0
	var o uint32
	for {
		t := &Tlv{}
		o, err = t.Unmarshal(data[offset:])
		offset += o
		if err != nil {
			return
		}
		if o == 0 {
			break
		}
		tlvs = append(tlvs, t)
	}
	return
}

func EncodeTlv(tlvs []*Tlv) []byte {
	var c uint32 = 0
	for _, tlv := range tlvs {
		c += tlv.marshalCap()
	}
	data := make([]byte, 0, c)
	for _, tlv := range tlvs {
		d, _ := tlv.MarshalBinary()
		data = append(data, d...)
	}
	return data
}

func NewTlv(t TlvType, id uint16, v any) *Tlv {
	var value []byte
	switch v.(type) {
	case string:
		value = []byte(v.(string))
	case *string:
		value = []byte(*v.(*string))
	default:
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, v)
		if err == nil {
			value = buf.Bytes()
		} else {
			value = make([]byte, 0)
		}
	}
	return &Tlv{
		Type:       t,
		Identifier: id,
		Value:      value,
		Length:     uint32(len(value)),
		Children:   make([]*Tlv, 0),
	}
}
