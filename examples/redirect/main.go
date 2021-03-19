package main

import (
	"log"
	"lwm2m/model"
	"lwm2m/path"
	"lwm2m/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var s *server.Server

func onDeviceConn(d *server.Device) {
	log.Println("new Device")
	log.Printf("%#v", d.Objs)
	//var oldval uint8 = 0
	if obj, ok := d.Objs[3342]; ok {
		if _, ok := obj.Instances[0]; ok {
			p, _ := path.NewPathFromString("3342/0/5500")
			_, _ = d.Observe(p, func(d *server.Device, notify []model.Node) {
				log.Println("callback")
				go func(d *server.Device) {
					time.Sleep(time.Second * 5)
					pp, _ := path.NewPathFromString("3311/0/5850")
					notify[0].(*model.Resource).Id = 5850
					d.Write(pp, notify...)
				}(d)
			})
			log.Println("Observe OK")
		}
	}
}

func main() {
	s = server.NewServer(server.WithOnNewDeviceConn(onDeviceConn))
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
