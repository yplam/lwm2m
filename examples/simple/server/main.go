package main

import (
	"log"
	"lwm2m"
)

func main() {
	s := lwm2m.NewServer()
	log.Fatal(s.ListenAndServeDTLS("udp", ":5684"))
}