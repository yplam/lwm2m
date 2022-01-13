package lwm2m

import (
	"fmt"
)

// Dummy store
type Dummy struct{}

// All device has the same key
func (d *Dummy) PSKFromIdentity(hint []byte) ([]byte, error) {
	fmt.Printf("Client's hint: %s \n", hint)
	return []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	}, nil
}

// ep == pskid
func (d *Dummy) PSKIdentityFromEP(ep []byte) ([]byte, error) {
	return ep, nil
}

func NewDummy() *Dummy {
	return &Dummy{}
}
