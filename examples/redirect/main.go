package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yplam/lwm2m"
)

var s *lwm2m.Server

var rules = map[string]string{
	"f4ce364d480904ee,/3342/0/5500": "f4ce36679cbbfb86,/3311/0/5850",
}

func handleNotify(d *lwm2m.Device, p lwm2m.Path, notify []lwm2m.Node) {
	log.Printf("new notify from %s:%s", d.EndPoint, p.String())
	for k, v := range rules {
		if strings.HasPrefix(k, d.EndPoint) {
			sps := strings.Split(k, ",")
			if len(sps) < 2 {
				return
			}
			fp, err := lwm2m.NewPathFromString(sps[1])
			if err != nil {
				return
			}
			if p != fp {
				return
			}
			log.Printf("match path %s", p.String())
			sps = strings.Split(v, ",")
			if len(sps) < 2 {
				return
			}
			td := s.GetByEndpoint(sps[0])
			if td == nil {
				return
			}
			tp, err := lwm2m.NewPathFromString(sps[1])
			if err != nil {
				return
			}
			log.Printf("sending to %s path %s", td.EndPoint, tp.String())
			rid, err := tp.ResourceId()
			if err != nil {
				log.Printf("%v", err)
				return
			}
			r := lwm2m.NewResource(rid, false)
			val, err := lwm2m.NodeGetResourceByPath(notify, p)
			if err != nil {
				log.Printf("%v", err)
				return
			}
			log.Printf("get value OK %#v", val)
			r.SetValue(val.Values[0])
			td.Write(tp, r)
		}
	}
}

func onDeviceConn(d *lwm2m.Device) {
	log.Printf("new device: %s", d.EndPoint)
	for k, v := range rules {
		if strings.HasPrefix(k, d.EndPoint) {
			sps := strings.Split(k, ",")
			if len(sps) >= 2 {
				p, err := lwm2m.NewPathFromString(sps[1])
				if err != nil {
					return
				}
				_, _ = d.Observe(p, handleNotify)
				log.Printf("redirecting %s to %s", k, v)
			}

		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	s := lwm2m.NewServer(
		lwm2m.WithOnNewDeviceConn(onDeviceConn),
		lwm2m.EnableUDPListener("udp", ":5683"),
		lwm2m.EnableDTLSListener("udp", ":5684", lwm2m.NewDummy()))
	go s.Serve(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
		// Exit by user
	}
	cancel()
	log.Println("Shutting down.")
}
