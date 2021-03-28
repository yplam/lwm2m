package encoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTlvsToNodes(t *testing.T) {
	tlvs, err := DecodeTLVs([]byte{0xe1, 0x15, 0x7c, 0x0,
		0xe8, 0x15, 0x7d, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0xe8, 0x16, 0xdc, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0xe8, 0x16, 0xde, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3})
	assert.Nil(t, err)
	nodes, err := tlvsToNodes(tlvs)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(nodes))

	tlvs, err = DecodeTLVs([]byte{0x8, 0x0, 0x28, 0xe1, 0x15, 0x7c, 0x0,
		0xe8, 0x15, 0x7d, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0xe8, 0x16, 0xdc, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0xe8, 0x16, 0xde, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1a, 0xd2})
	assert.Nil(t, err)
	nodes, err = tlvsToNodes(tlvs)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(nodes))
}