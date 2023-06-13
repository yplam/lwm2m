package encoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpaque(t *testing.T) {
	{ // int -> int64
		val, err := NewOpaqueValue(42)
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a}, val.Raw())
	}
	{ // int64
		val, err := NewOpaqueValue(int64(42))
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a}, val.Raw())
	}
	{ // int32
		val, err := NewOpaqueValue(int32(42))
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x2a}, val.Raw())
	}
	{ // int16
		val, err := NewOpaqueValue(int16(42))
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x00, 0x2a}, val.Raw())
	}
	{ // float #1
		val, err := NewOpaqueValue(42.0)
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x40, 0x45, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, val.Raw())
	}
	{ // float #2
		val, err := NewOpaqueValue(3.14151617181920)
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x40, 0x09, 0x21, 0xd3, 0x3b, 0xe, 0x8c, 0x74}, val.Raw())
	}
	{ // string
		val, err := NewOpaqueValue("3.14151617181920")
		assert.Nil(t, err)
		assert.Equal(t, []byte("3.14151617181920"), val.Raw())
	}
	{ // bool
		val, err := NewOpaqueValue(true)
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x01}, val.Raw())
		val, err = NewOpaqueValue(false)
		assert.Equal(t, nil, err)
		assert.Equal(t, []byte{0x00}, val.Raw())
	}
	{ // opaque
		val, err := NewOpaqueValue([]byte{0x01, 0x02, 0x03, 0x04})
		assert.Nil(t, err)
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, val.Raw())
	}
}
