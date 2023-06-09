package encoding

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type PlainTextValue struct {
	Value []byte
}

func (p *PlainTextValue) StringVal() string {
	return string(p.Value)
}

func (p *PlainTextValue) Integer() (val int64, err error) {
	return strconv.ParseInt(string(p.Value), 10, 64)
}

func (p *PlainTextValue) Float() (float64, error) {
	return strconv.ParseFloat(string(p.Value), 64)
}

func (p *PlainTextValue) Boolean() (val bool, err error) {
	// Represented as the ASCII value 0 or 1
	if string(p.Value) == "1" {
		return true, nil
	} else if string(p.Value) == "0" {
		return false, nil
	}
	return false, errors.New("not a boolean representation")
}

func (p *PlainTextValue) Opaque() []byte {
	// Represented as a Base64 encoding of binary bytes
	opaque, err := base64.StdEncoding.DecodeString(string(p.Value))
	if err != nil {
		return nil
	}
	return opaque
}

func (p *PlainTextValue) Time() (int64, error) {
	// same as integer
	return p.Integer()
}

func (p *PlainTextValue) ObjectLink() (uint16, uint16, error) {
	splitted := strings.Split(string(p.Value), ":")
	if len(splitted) != 2 {
		return 0, 0, errors.New("wrong ObjLink text representation")
	}
	od, err := strconv.ParseUint(splitted[0], 10, 16)
	if err != nil {
		return 0, 0, err
	}
	oid, err := strconv.ParseUint(splitted[1], 10, 16)
	if err != nil {
		return 0, 0, err
	}
	return uint16(od), uint16(oid), nil
}

func (p *PlainTextValue) Raw() []byte {
	return p.Value
}

func NewPlainTextRaw(data []byte) *PlainTextValue {
	return &PlainTextValue{Value: data}
}

func (p *PlainTextValue) fromInteger(val int64) error {
	str := fmt.Sprintf("%d", val)
	p.Value = []byte(str)
	return nil
}
func (p *PlainTextValue) fromFloat(val float64) error {
	str := fmt.Sprintf("%v", val)
	p.Value = []byte(str)
	return nil
}
func (p *PlainTextValue) fromBool(val bool) error {
	str := "0"
	if val {
		str = "1"
	}
	p.Value = []byte(str)
	return nil
}
func (p *PlainTextValue) fromTime(val int64) error {
	return p.fromInteger(val)
}
func (p *PlainTextValue) fromString(val string) error {
	p.Value = []byte(val)
	return nil
}
func (p *PlainTextValue) fromOpaque(val []byte) error {
	raw := base64.StdEncoding.EncodeToString(val)
	p.Value = []byte(raw)
	return nil
}
func NewPlainTextValue(val any) (*PlainTextValue, error) {
	var pt PlainTextValue
	var err error
	switch val.(type) {
	case int:
		err = pt.fromInteger(int64(val.(int)))
	case int8:
		err = pt.fromInteger(int64(val.(int8)))
	case int16:
		err = pt.fromInteger(int64(val.(int16)))
	case int32:
		err = pt.fromInteger(int64(val.(int32)))
	case int64:
		err = pt.fromInteger(val.(int64))
	case uint:
		err = pt.fromInteger(int64(val.(uint)))
	case uint8:
		err = pt.fromInteger(int64(val.(uint8)))
	case uint16:
		err = pt.fromInteger(int64(val.(uint16)))
	case uint32:
		err = pt.fromInteger(int64(val.(uint32)))
	case float32:
		err = pt.fromFloat(float64(val.(float32)))
	case float64:
		err = pt.fromFloat(val.(float64))
	case string:
		err = pt.fromString(val.(string))
	case []byte:
		err = pt.fromOpaque(val.([]byte))
	case bool:
		err = pt.fromBool(val.(bool))
	default:
		err = errors.New("unknown type")
	}
	if err != nil {
		return nil, err
	}
	return &pt, err
}
