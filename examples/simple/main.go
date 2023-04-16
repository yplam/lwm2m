package main

import (
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/yplam/lwm2m/registration"
	"github.com/yplam/lwm2m/server"
	"log"
)

func main() {
	r := mux.NewRouter()
	_ = registration.NewHandler(r)
	err := server.ListenAndServe(r,
		server.EnableUDPListener("udp", ":5683"),
		server.EnableTCPListener("tcp", ":5685"))
	if err != nil {
		log.Printf("serve lwm2m with err: %v", err)
	}
}
