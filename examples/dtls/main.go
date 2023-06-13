package main

import (
	"github.com/yplam/lwm2m/core"
	"github.com/yplam/lwm2m/registration"
	"github.com/yplam/lwm2m/server"
	"log"
)

func PSKFromIdentity(hint []byte) ([]byte, error) {
	return []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	}, nil
}

func main() {
	r := server.DefaultRouter()
	deviceManager := core.DefaultManager()
	registration.EnableHandler(r, deviceManager)
	err := server.ListenAndServe(r,
		server.EnableUDPListener("udp", ":5683"),
		server.EnableDTLSListener("udp", ":5684", PSKFromIdentity),
	)
	if err != nil {
		log.Printf("serve lwm2m with err: %v", err)
	}
}
