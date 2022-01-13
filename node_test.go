package lwm2m

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeGetAllResources(t *testing.T) {
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
	p := NewObjectInstancePath(3, 0)
	v, err := DecodeTLVs(data)
	assert.Nil(t, err)
	ns, err := decodeTLVMessage(p, v)
	assert.Nil(t, err)
	assert.Equal(t, 13, len(ns))
	assert.Equal(t, uint16(0), ns[0].ID())
	assert.IsType(t, &Resource{}, ns[0])
	assert.Equal(t, R_STRING, ns[0].(*Resource).resType)
	assert.Equal(t, uint16(1), ns[1].ID())
	assert.Equal(t, uint16(6), ns[4].ID())
	assert.IsType(t, &Resource{}, ns[4])
	r, ok := ns[4].(*Resource)
	assert.Equal(t, R_INTEGER, r.resType)
	assert.Equal(t, R_INTEGER, r.instances[0].resType)
	assert.Equal(t, int64(1), r.instances[0].Value())
	assert.Equal(t, int64(5), r.instances[1].Value())
	assert.True(t, ok)
	assert.Equal(t, 2, len(r.instances))
	//for _, n := range ns {
	//	t.Logf("%v", n)
	//}

	rs, err := NodeGetAllResources(ns, p)
	assert.Nil(t, err)
	assert.Equal(t, 13, len(rs))
	//for k, rr := range rs {
	//	t.Logf("%v, %v", k, rr)
	//}
}

func TestDecodeObjectInstance(t *testing.T) {
	data := []byte{
		0x8, 0x0, 0x26,
		0xe4, 0x16, 0x44, 0x42, 0x80, 0xe7, 0xb7, 0xe0, 0x16, 0x45, 0xe4, 0x15, 0xe1,
		0x42, 0x80, 0xe0, 0xaf, 0xe4, 0x15, 0xe2, 0x42, 0x80, 0xf5, 0xc7, 0xe4, 0x15,
		0xe3, 0x0, 0x0, 0x0, 0x0, 0xe4, 0x15, 0xe4, 0x0, 0x0, 0x0, 0x0, 0x8, 0x1,
		0x26, 0xe4, 0x16, 0x44, 0x0, 0x0, 0x0, 0x0, 0xe0, 0x16,
		0x45, 0xe4, 0x15, 0xe1, 0x4e, 0xff, 0xff, 0xff, 0xe4,
		0x15, 0xe2, 0xce, 0xff, 0xff, 0xff, 0xe4, 0x15, 0xe3, 0x0, 0x0, 0x0, 0x0, 0xe4,
		0x15, 0xe4, 0x0, 0x0, 0x0, 0x0}
	v, err := DecodeTLVs(data)
	assert.Nil(t, err)
	//e, _ := json.MarshalIndent(v, "", "\t")
	//t.Logf("%v", string(e))
	p := NewObjectPath(3303)
	ns, err := decodeTLVMessage(p, v)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(ns))
	_, err = NodeGetAllResources(ns, p)
	assert.Nil(t, err)
	//for k, rr := range rs {
	//	t.Logf("%v, %v", k, rr)
	//}
}

func TestSimplePackage(t *testing.T) {
	data := []byte{228, 22, 68, 65, 177, 134, 201}
	v, err := DecodeTLVs(data)
	assert.Nil(t, err)
	//e, _ := json.MarshalIndent(v, "", "\t")
	//t.Logf("%v", string(e))
	p := NewResourcePath(3303, 1, 5700)
	ns, err := decodeTLVMessage(p, v)
	assert.Nil(t, err)
	_, err = NodeGetAllResources(ns, p)
	assert.Nil(t, err)
	//for k, rr := range rs {
	//	t.Logf("%v, %v", k, rr)
	//}
}
