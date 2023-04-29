package main

import (
	"github.com/yplam/lwm2m/core"
	"github.com/yplam/lwm2m/registration"
	"github.com/yplam/lwm2m/server"
	"log"
)

func main() {
	r := server.DefaultRouter()
	deviceManager := core.DefaultManager()
	registration.EnableHandler(r, deviceManager)
	err := server.ListenAndServe(r,
		server.EnableUDPListener("udp", ":5683"))
	if err != nil {
		log.Printf("serve lwm2m with err: %v", err)
	}
}
