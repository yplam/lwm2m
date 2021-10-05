package lwm2m

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPathFromString(t *testing.T) {
	p, err := NewPathFromString("/")
	assert.Nil(t, err)
	assert.Equal(t, true, p.IsRoot())

	p, err = NewPathFromString("/3/")
	assert.Nil(t, err)
	assert.Equal(t, true, p.IsObject())
	v, err := p.ObjectId()
	assert.Nil(t, err)
	assert.Equal(t, uint16(3), v)
	p, err = NewPathFromString("3")
	assert.Nil(t, err)
	assert.Equal(t, true, p.IsObject())
	assert.Equal(t, "/3", p.String())

	p, err = NewPathFromString("/3/4")
	assert.Nil(t, err)
	assert.Equal(t, true, p.IsObjectInstance())

	p, err = NewPathFromString("/3/4/5")
	assert.Nil(t, err)
	assert.Equal(t, true, p.IsResource())

	p, err = NewPathFromString("/3/4/5/6")
	assert.Nil(t, err)
	assert.Equal(t, true, p.IsResourceInstance())
	assert.Equal(t, "/3/4/5/6", p.String())
}
