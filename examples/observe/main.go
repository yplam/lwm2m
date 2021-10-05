package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/yplam/lwm2m"
)

func handleNotify(d *lwm2m.Device, p lwm2m.Path, notify []lwm2m.Node) {
	logrus.Infof("new notify from %s:%s", d.EndPoint, p.String())
	val, err := lwm2m.NodeGetAllResources(notify, p)
	if err != nil {
		logrus.Warnf("read val err (%v), (%v)", p.String(), err)
	} else {
		logrus.Infof("get val from (%v)", p.String())
		for k, v := range val {
			logrus.Infof("(%v), (%v)", k.String(), v)
		}

	}
}

func onDeviceConn(d *lwm2m.Device) {
	logrus.Infof("new device: %s", d.EndPoint)
	p, _ := lwm2m.NewPathFromString("3303")
	d.Observe(p, handleNotify)
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("starting lwm2m server")
	s := lwm2m.NewServer(
		lwm2m.WithOnNewDeviceConn(onDeviceConn),
		lwm2m.EnableUDPListener("udp", ":5683"),
		lwm2m.EnableDTLSListener("udp", ":5684", lwm2m.NewDummy()))
	go s.Serve(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	cancel()
	logrus.Info("Shutting down.")
}
