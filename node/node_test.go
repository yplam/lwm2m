package node

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func _strToByte(str string) (dst []byte, err error) {
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, " ", "")
	dst = make([]byte, hex.DecodedLen(len(str)))
	_, err = hex.Decode(dst, []byte(str))
	return
}

func decodeSingleObjectTLV(t *testing.T) ([]Node, error) {
	str := `
	C8 00 14 4F 70 65 6E 20 4D 6F 62 69 6C 65 20 41 6C 6C 69 61 6E 63 65
	C8 01 16 4C 69 67 68 74 77 65 69 67 68 74 20 4D 32 4D 20 43 6C 69 65 6E 74
	C8 02 09 33 34 35 30 30 30 31 32 33
	C3 03 31 2E 30
	86 06
	41 00 01
	41 01 05
	88 07 08
	42 00 0E D8
	42 01 13 88
	87 08
	41 00 7D
	42 01 03 84
	C1 09 64
	C1 0A 0F
	83 0B
	41 00 00
	C4 0D 51 82 42 8F
	C6 0E 2B 30 32 3A 30 30
	C1 10 55`
	data, err := _strToByte(str)
	assert.Nil(t, err)
	msg := pool.NewMessage(context.Background())
	msg.SetContentFormat(message.AppLwm2mTLV)
	msg.SetBody(bytes.NewReader(data))
	return DecodeMessage(NewObjectInstancePath(3, 0), msg)
}

func TestDecodeSingleObjectTLV(t *testing.T) {
	nodes, err := decodeSingleObjectTLV(t)
	assert.Nil(t, err)
	assert.Equal(t, 13, len(nodes))

	res, ok := nodes[0].(*Resource)
	assert.Equal(t, true, ok)
	ins, err := res.GetInstance(0)
	assert.Nil(t, err)
	assert.Equal(t, "Open Mobile Alliance", ins.String())
	assert.Equal(t, "Open Mobile Alliance", ins.Value())

	res, ok = nodes[4].(*Resource)
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, res.InstanceCount())
	ins, err = res.GetInstance(0)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), ins.Value())
	ins, err = res.GetInstance(1)
	assert.Nil(t, err)
	assert.Equal(t, int64(5), ins.Value())
}

func TestNodeGetAllResources(t *testing.T) {
	nodes, err := decodeSingleObjectTLV(t)
	assert.Nil(t, err)
	ress, err := GetAllResources(nodes, NewObjectInstancePath(3, 0))
	assert.Nil(t, err)
	assert.Equal(t, 13, len(ress))
	res, ok := ress[NewResourcePath(3, 0, 3)]
	assert.Equal(t, true, ok)
	ins, err := res.GetInstance(0)
	assert.Nil(t, err)
	assert.Equal(t, "1.0", ins.Value())

	nodes, err = decodeSingleInstanceObjectTLV(t)
	assert.Nil(t, err)
	ress, err = GetAllResources(nodes, NewObjectInstancePath(3, 0))
	assert.Nil(t, err)
	res, ok = ress[NewResourcePath(3, 0, 6)]
	assert.Equal(t, true, ok)
	ins, err = res.GetInstance(0)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), ins.Value())
	ins, err = res.GetInstance(1)
	assert.Nil(t, err)
	assert.Equal(t, int64(5), ins.Value())
}

func decodeSingleInstanceObjectTLV(t *testing.T) ([]Node, error) {
	str := `
08 00 79
C8 00 14 4F 70 65 6E 20 4D 6F 62 69 6C 65 20 41 6C 6C 69 61 6E 63 65
C8 01 16 4C 69 67 68 74 77 65 69 67 68 74 20 4D 32 4D 20 43 6C 69 65 6E 74
C8 02 09 33 34 35 30 30 30 31 32 33
C3 03 31 2E 30
86 06
41 00 01
41 01 05
88 07 08
42 00 0E D8
42 01 13 88
87 08
41 00 7D
42 01 03 84
C1 09 64
C1 0A 0F
83 0B
41 00 00
C4 0D 51 82 42 8F
C6 0E 2B 30 32 3A 30 30
C1 10 55
	`
	data, err := _strToByte(str)
	assert.Nil(t, err)
	msg := pool.NewMessage(context.Background())
	msg.SetContentFormat(message.AppLwm2mTLV)
	msg.SetBody(bytes.NewReader(data))
	return DecodeMessage(NewObjectPath(3), msg)
}

func TestDecodeSingleInstanceObjectTLV(t *testing.T) {
	nodes, err := decodeSingleInstanceObjectTLV(t)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(nodes))
	obj, ok := nodes[0].(*ObjectInstance)
	assert.Equal(t, true, ok)
	res, ok := obj.Resources[3]
	assert.Equal(t, true, ok)
	ins, err := res.GetInstance(0)
	assert.Nil(t, err)
	assert.Equal(t, "1.0", ins.Value())
}

func decodeMultipleInstanceObjectTLV(t *testing.T) ([]Node, error) {
	str := `
08 00 0E
C1 00 01
C1 01 00
83 02
41 7F 07
C1 03 7F
08 02 12
C1 00 03
C1 01 00
87 02 41 7F 07 61 01 36 01
C1 03 7F
	`
	data, err := _strToByte(str)
	assert.Nil(t, err)
	msg := pool.NewMessage(context.Background())
	msg.SetContentFormat(message.AppLwm2mTLV)
	msg.SetBody(bytes.NewReader(data))
	return DecodeMessage(NewObjectPath(2), msg)
}

func TestDecodeMultipleInstanceObjectTLV(t *testing.T) {
	nodes, err := decodeMultipleInstanceObjectTLV(t)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(nodes))
	obj, ok := nodes[0].(*ObjectInstance)
	assert.Equal(t, true, ok)
	res, ok := obj.Resources[2]
	assert.Equal(t, true, ok)
	assert.Equal(t, true, res.isMultiple)
}

func decodePlainText(t *testing.T, str string) ([]Node, error) {
	msg := pool.NewMessage(context.Background())
	msg.SetContentFormat(message.TextPlain)
	msg.SetBody(bytes.NewReader([]byte(str)))
	return DecodeMessage(NewResourcePath(1, 0, 1), msg)
}
func TestDecodePlainText(t *testing.T) {

	{ // payload #0
		data := "0"
		nn, err := decodePlainText(t, data)
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			bb, err := rr.Data().Boolean()
			assert.Nil(t, err)
			assert.False(t, bb)
			ii, err := rr.Data().Integer()
			assert.Nil(t, err)
			assert.Equal(t, int64(0), ii)
			ff, err := rr.Data().Float()
			assert.Nil(t, err)
			assert.Equal(t, float64(0.0), ff)
			ss := rr.Data().StringVal()
			assert.Equal(t, data, ss)
			op := rr.Data().Opaque()
			assert.Nil(t, op)
		default:
			t.Fatal("not Resource")
		}
	}

	{ // payload #1
		data := "11.10"
		nn, err := decodePlainText(t, data)
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			_, err = rr.Data().Boolean()
			assert.NotNil(t, err)
			_, err = rr.Data().Integer()
			assert.NotNil(t, err)
			ff, err := rr.Data().Float()
			assert.Nil(t, err)
			assert.Equal(t, float64(11.1), ff)
			ss := rr.Data().StringVal()
			assert.Equal(t, "11.10", ss)
			op := rr.Data().Opaque()
			assert.Nil(t, op)
		default:
			t.Fatal("not Resource")
		}
	}
	{ // payload #2
		data := "AQIDBA=="
		nn, err := decodePlainText(t, data)
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			_, err = rr.Data().Boolean()
			assert.NotNil(t, err)
			_, err = rr.Data().Integer()
			assert.NotNil(t, err)
			_, err = rr.Data().Float()
			assert.NotNil(t, err)
			ss := rr.Data().StringVal()
			assert.Equal(t, data, ss)
			op := rr.Data().Opaque()
			assert.True(t, bytes.Equal(op, []byte{0x01, 0x02, 0x03, 0x04}))
		default:
			t.Fatal("not Resource")
		}
	}
	{ // payload #3
		data := "1"
		nn, err := decodePlainText(t, data)
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			bb, err := rr.Data().Boolean()
			assert.Nil(t, err)
			assert.True(t, bb)
			ii, err := rr.Data().Integer()
			assert.Nil(t, err)
			assert.Equal(t, int64(1), ii)
			ff, err := rr.Data().Float()
			assert.Nil(t, err)
			assert.Equal(t, 1.0, ff)
			ss := rr.Data().StringVal()
			assert.Equal(t, data, ss)
			op := rr.Data().Opaque()
			assert.Nil(t, op)
		default:
			t.Fatal("not Resource")
		}
	}
	{ // payload #4
		data := "hello friend"
		nn, err := decodePlainText(t, data)
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			_, err = rr.Data().Boolean()
			assert.NotNil(t, err)
			_, err = rr.Data().Integer()
			assert.NotNil(t, err)
			_, err = rr.Data().Float()
			assert.NotNil(t, err)
			ss := rr.Data().StringVal()
			assert.Equal(t, data, ss)
			op := rr.Data().Opaque()
			assert.Nil(t, op)
		default:
			t.Fatal("not Resource")
		}
	}
}

func decodeAppOctets(t *testing.T, data []byte) ([]Node, error) {
	msg := pool.NewMessage(context.Background())
	msg.SetContentFormat(message.AppOctets)
	msg.SetBody(bytes.NewReader(data))
	return DecodeMessage(NewResourcePath(1, 0, 1), msg)
}
func TestDecodeOpaque(t *testing.T) {

	{ // payload #0
		data := []byte{0x00}
		nn, err := decodeAppOctets(t, data)
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			bb, err := rr.Data().Boolean()
			assert.Nil(t, err)
			assert.False(t, bb)
			ii, err := rr.Data().Integer()
			assert.Nil(t, err)
			assert.Equal(t, int64(0), ii)
			_, err = rr.Data().Float()
			assert.NotNil(t, err)
			ss := rr.Data().StringVal()
			assert.Equal(t, string([]byte{0x00}), ss)
			op := rr.Data().Opaque()
			assert.NotNil(t, op)
			assert.Equal(t, []byte{0x00}, op)
		default:
			t.Fatal("not Resource")
		}
	}

	{ // payload #1
		data := []byte{0x40, 0x09, 0x21, 0xd3, 0x3b, 0xe, 0x8c, 0x74}
		nn, err := decodeAppOctets(t, data)
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			_, err := rr.Data().Boolean()
			assert.NotNil(t, err)
			ii, err := rr.Data().Integer()
			assert.Nil(t, err)
			assert.Equal(t, int64(4614256484330409076), ii)
			ff, err := rr.Data().Float()
			assert.Nil(t, err)
			assert.Equal(t, float64(3.14151617181920), ff)
			ss := rr.Data().StringVal()
			assert.Equal(t, string([]byte{0x40, 0x09, 0x21, 0xd3, 0x3b, 0xe, 0x8c, 0x74}), ss)
			op := rr.Data().Opaque()
			assert.NotNil(t, op)
			assert.Equal(t, []byte{0x40, 0x09, 0x21, 0xd3, 0x3b, 0xe, 0x8c, 0x74}, op)
		default:
			t.Fatal("not Resource")
		}
	}
	{ // payload #2
		data := "long long string dorem ipsum; once upon a time in a far away lands..."
		nn, err := decodeAppOctets(t, []byte(data))
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			_, err = rr.Data().Boolean()
			assert.NotNil(t, err)
			_, err = rr.Data().Integer()
			assert.NotNil(t, err)
			_, err = rr.Data().Float()
			assert.NotNil(t, err)
			ss := rr.Data().StringVal()
			assert.Equal(t, data, ss)
			op := rr.Data().Opaque()
			assert.True(t, bytes.Equal(op, []byte(data)))
		default:
			t.Fatal("not Resource")
		}
	}
	{ // payload #3
		data := "1"
		nn, err := decodeAppOctets(t, []byte(data))
		assert.Nil(t, err)
		assert.Len(t, nn, 1)
		n := nn[0]
		switch n.(type) {
		case *Resource:
			rr, _ := n.(*Resource)
			_, err = rr.Data().Boolean()
			assert.NotNil(t, err)
			ii, err := rr.Data().Integer()
			assert.Nil(t, err)
			assert.Equal(t, int64(49), ii)
			_, err = rr.Data().Float()
			assert.NotNil(t, err)
			ss := rr.Data().StringVal()
			assert.Equal(t, data, ss)
			op := rr.Data().Opaque()
			assert.NotNil(t, op)
			assert.Equal(t, []byte{0x31}, op)
		default:
			t.Fatal("not Resource")
		}
	}
}
