package store

import (
	"fmt"
)

type Dummy struct {
	
}

func (d *Dummy) PSKFromIdentity(hint []byte) ([]byte, error) {
	fmt.Printf("Client's hint: %s \n", hint)
	if string(hint) == "abcd" {
		return []byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		}, nil
	}
	return nil, fmt.Errorf("404 Not Found")
}
func (d *Dummy) PSKIdentityFromEP(ep []byte) ([]byte, error) {
	if string(ep) == "f4ce36304d0c224b" {
		return []byte("abcd"), nil
	}
	return nil, fmt.Errorf("404 Not Found")
}


func NewDummy() *Dummy {
	return & Dummy{}
}