package lwm2m

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestLinksFromString(t *testing.T) {
	var links []*CoreLink

	links = CoreLinksFromString("</>;rt=\"oma.lwm2m\";ct=11543,</1/0>,</3/0>,</5/0>,</3303/0>,</3300>,</3304>,</3323>,</3311>,</3340>,</3342>,</3347>")

	assert.Equal(t, 12, len(links))
	//
	//resources = CoreResourcesFromString("</sensors/temp>;ct=41;rt=\"temperature-c\";if=\"sensor\", </sensors/light>;ct=41;rt=\"light-lux\";if=\"sensor\"")
	//
	//assert.Equal(t, 2, len(resources))
	//resource1 := resources[0]
	//
	//assert.Equal(t, "/sensors/temp", resource1.Target)
	//assert.Equal(t, 3, len(resource1.Attributes))
	//
	//assert.Nil(t, resource1.GetAttribute("invalid_attr"))
	//
	//assert.NotNil(t, resource1.GetAttribute("ct"))
	//assert.Equal(t, "ct", resource1.GetAttribute("ct").Key)
	//assert.Equal(t, "41", resource1.GetAttribute("ct").Value)
	//
	//assert.NotNil(t, resource1.GetAttribute("rt"))
	//assert.Equal(t, "rt", resource1.GetAttribute("rt").Key)
	//assert.Equal(t, "temperature-c", resource1.GetAttribute("rt").Value)
	//
	//assert.NotNil(t, resource1.GetAttribute("if"))
	//assert.Equal(t, "if", resource1.GetAttribute("if").Key)
	//assert.Equal(t, "sensor", resource1.GetAttribute("if").Value)
	//
	//resource2 := resources[1]
	//assert.Equal(t, "/sensors/light", resource2.Target)
	//assert.Equal(t, 3, len(resource2.Attributes))
	//
	//assert.NotNil(t, resource2.GetAttribute("ct"))
	//assert.Equal(t, "ct", resource2.GetAttribute("ct").Key)
	//assert.Equal(t, "41", resource2.GetAttribute("ct").Value)
	//
	//assert.NotNil(t, resource2.GetAttribute("rt"))
	//assert.Equal(t, "rt", resource2.GetAttribute("rt").Key)
	//assert.Equal(t, "light-lux", resource2.GetAttribute("rt").Value)
	//
	//assert.NotNil(t, resource2.GetAttribute("if"))
	//assert.Equal(t, "if", resource2.GetAttribute("if").Key)
	//assert.Equal(t, "sensor", resource2.GetAttribute("if").Value)
}

