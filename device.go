package lwm2m

import (
	"github.com/plgd-dev/go-coap/v2/mux"
)

type Device struct {
	ID        string
	EndPoint  string
	Version   string
	Lifetime  int
	client    mux.Client
	Binding   string
	Sms       string
	Objs      map[uint16]*Object
}
