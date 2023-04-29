package encoding

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTLVLen(t *testing.T) {
	assert.Equal(t, tlvLen([]byte{0x03}), uint32(3))
	assert.Equal(t, tlvLen([]byte{0x03, 0x34}), uint32(0x0334))
}

func _strToByte(str string) (dst []byte, err error) {
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, " ", "")
	dst = make([]byte, hex.DecodedLen(len(str)))
	_, err = hex.Decode(dst, []byte(str))
	return
}

func TestSingleObjectTLV(t *testing.T) {
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
	//log.Printf("%#v", data)
	v, err := DecodeTlv(data)
	assert.Nil(t, err)
	//e, _ := json.MarshalIndent(v, "", "\t")
	//t.Logf("%v", string(e))
	assert.Equal(t, len(v), 13)
	assert.Equal(t, len(v[4].Children), 2)
	assert.Equal(t, len(v[5].Children), 2)
	assert.Equal(t, len(v[6].Children), 2)
	assert.Equal(t, len(v[9].Children), 1)

	s := v[0].String()
	assert.Equal(t, s, "Open Mobile Alliance")
	s = v[1].String()
	assert.Equal(t, s, "Lightweight M2M Client")

	i8, err := v[7].Integer()
	assert.Nil(t, err)
	assert.Equal(t, i8, int64(0x64))
	u8, err := v[7].Integer()
	assert.Nil(t, err)
	assert.Equal(t, u8, int64(0x64))

	i16, err := v[5].Children[0].Integer()
	assert.Nil(t, err)
	assert.Equal(t, i16, int64(0x0ed8))
	u16, err := v[5].Children[1].Integer()
	assert.Nil(t, err)
	assert.Equal(t, u16, int64(0x1388))

	u32, err := v[10].Integer()
	assert.Nil(t, err)
	assert.Equal(t, u32, int64(0x5182428F))
	i32, err := v[10].Integer()
	assert.Nil(t, err)
	assert.Equal(t, i32, int64(0x5182428F))

}

func TestSingleInstanceObjectTLV(t *testing.T) {
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
	//log.Printf("%#v", data)
	v, err := DecodeTlv(data)
	//e, _ := json.MarshalIndent(v, "", "\t")
	//t.Logf("%v", string(e))
	assert.Nil(t, err)
	assert.Equal(t, len(v), 1)
	assert.Equal(t, len(v[0].Children), 13)
}

func TestMultipleInstanceObjectTLV(t *testing.T) {
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
	//log.Printf("%#v", data)
	v, err := DecodeTlv(data)
	assert.Nil(t, err)
	assert.Equal(t, len(v), 2)
	assert.Equal(t, len(v[0].Children), 4)
	assert.Equal(t, len(v[1].Children), 4)
	assert.Equal(t, len(v[1].Children[2].Children), 2)

	str = `
08 00 0F
C1 00 01
C4 01 00 01 51 80
C1 06 01
C1 07 55
`
	data, err = _strToByte(str)
	assert.Nil(t, err)
	//log.Printf("%#v", data)
	_, err = DecodeTlv(data)
	assert.Nil(t, err)
}

func TestObjectLinkTLV(t *testing.T) {
	str := `
88 00 0C
44 00 00 42 00 00
44 01 00 42 00 01
C8 01 0D 38 36 31 33 38 30 30 37 35 35 35 30 30
C4 02 12 34 56 78
`
	data, err := _strToByte(str)
	assert.Nil(t, err)
	//log.Printf("%#v", data)
	_, err = DecodeTlv(data)
	assert.Nil(t, err)

	str = `
08 00 26
C8 00 0B 6D 79 53 65 72 76 69 63 65 20 31
C8 01 0F 49 6E 74 65 72 6E 65 74 2E 31 35 2E 32 33 34
C4 02 00 43 00 00
08 01 26
C8 00 0B 6D 79 53 65 72 76 69 63 65 20 32
C8 01 0F 49 6E 74 65 72 6E 65 74 2E 31 35 2E 32 33 35
C4 02 FF FF FF FF
`
	data, err = _strToByte(str)
	assert.Nil(t, err)
	//log.Printf("%#v", data)
	_, err = DecodeTlv(data)
	assert.Nil(t, err)

}

func TestSimpleEncodeDecode(t *testing.T) {
	data := []byte{0xe1, 0x15, 0x7c, 0x0}
	tlvs, err := DecodeTlv(data)
	assert.Nil(t, err)
	data2 := EncodeTlv(tlvs)
	assert.Equal(t, data, data2)
}
