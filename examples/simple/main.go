package main

import (
	"log"
	"lwm2m"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	s := lwm2m.NewServer()
	go func() {
		log.Fatal(s.ListenAndServeDTLS("udp", ":5684"))
	}()
	defer s.Stop()

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
		// Exit by user
	}
	log.Println("Shutting down.")
}
