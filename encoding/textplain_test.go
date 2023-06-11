package encoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlainText(t *testing.T) {
	{ // int
		val, err := NewPlainTextValue(42)
		assert.Nil(t, err)
		assert.Equal(t, []byte("42"), val.Raw())
	}
	{ // float #1
		val, err := NewPlainTextValue(42.21)
		assert.Nil(t, err)
		assert.Equal(t, []byte("42.21"), val.Raw())
	}
	{ // float #2
		val, err := NewPlainTextValue(3.14151617181920)
		assert.Nil(t, err)
		assert.Equal(t, []byte("3.1415161718192"), val.Raw())
	}
	{ // string
		val, err := NewPlainTextValue("3.14151617181920")
		assert.Nil(t, err)
		assert.Equal(t, []byte("3.14151617181920"), val.Raw())
	}
	{ // bool
		val, err := NewPlainTextValue(true)
		assert.Nil(t, err)
		assert.Equal(t, []byte("1"), val.Raw())
		val, err = NewPlainTextValue(false)
		assert.Equal(t, nil, err)
		assert.Equal(t, []byte("0"), val.Raw())
	}
	{ // opaque
		val, err := NewPlainTextValue([]byte{0x01, 0x02, 0x03, 0x04})
		assert.Nil(t, err)
		assert.Equal(t, []byte("AQIDBA=="), val.Raw())
	}
}
