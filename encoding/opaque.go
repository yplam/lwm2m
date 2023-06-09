package encoding

import (
	"bytes"
	"encoding/binary"
	"math"
)

// Opaque value should be used for binary resources like firmware image
// although  define it for every type
// with same rules as Value field in TLV

type OpaqueValue struct {
	Value []byte
}

func (o *OpaqueValue) StringVal() string {
	return string(o.Value)
}

func (o *OpaqueValue) Integer() (val int64, err error) {
	buff := bytes.NewBuffer(o.Value)
	if len(o.Value) == 1 {
		var i1 int8
		err = binary.Read(buff, binary.BigEndian, &i1)
		val = int64(i1)
	} else if len(o.Value) == 2 {
		var i2 int16
		err = binary.Read(buff, binary.BigEndian, &i2)
		val = int64(i2)
	} else if len(o.Value) == 4 {
		var i4 int32
		err = binary.Read(buff, binary.BigEndian, &i4)
		val = int64(i4)
	} else if len(o.Value) == 8 {
		var i8 int64
		err = binary.Read(buff, binary.BigEndian, &i8)
		val = i8
	} else {
		err = ErrTlvInvalidLength
	}
	return
}

func (o *OpaqueValue) Float() (float64, error) {
	if len(o.Value) == 4 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(o.Value))), nil
	} else if len(o.Value) == 8 {
		return math.Float64frombits(binary.BigEndian.Uint64(o.Value)), nil
	} else {
		return 0, ErrTlvInvalidLength
	}
}

func (o *OpaqueValue) Boolean() (val bool, err error) {
	if len(o.Value) != 1 {
		return false, ErrTlvInvalidLength
	}
	val = o.Value[0] != 0
	return val, nil
}

func (o *OpaqueValue) Opaque() []byte {
	return o.Value
}

func (o *OpaqueValue) Time() (int64, error) {
	if len(o.Value) == 4 {
		return int64(binary.BigEndian.Uint32(o.Value)), nil
	} else if len(o.Value) == 8 {
		return int64(binary.BigEndian.Uint64(o.Value)), nil
	} else {
		return 0, ErrTlvInvalidLength
	}
}

func (o *OpaqueValue) ObjectLink() (uint16, uint16, error) {
	if len(o.Value) != 4 {
		return 0, 0, ErrTlvInvalidLength
	}
	_ = o.Value[3]
	return uint16(o.Value[0])*256 + uint16(o.Value[1]),
		uint16(o.Value[2])*256 + uint16(o.Value[3]), nil
}

func (o *OpaqueValue) Raw() []byte {
	return o.Value
}

func NewOpaqueValue(val any) (*OpaqueValue, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, val)
	if err != nil {
		return nil, err
	}
	return &OpaqueValue{Value: buf.Bytes()}, nil
}
