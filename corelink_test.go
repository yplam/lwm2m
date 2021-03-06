package lwm2m

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestCoreLinksFromString(t *testing.T) {
	var links []*coreLink

	links = coreLinksFromString("</>;rt=\"oma.lwm2m\";" +
		"ct=11543,</1/0>,</3/0>,</5/0>,</3303/0>,</3300>,</3304>,</3323>," +
		"</3311>,</3340>,</3342>,</3347>")

	assert.Equal(t, 12, len(links))
}

