package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/yplam/lwm2m"
)

var (
	objIds = []uint16{3300, 3303, 3304}
)

func handleNotify(d *lwm2m.Device, p lwm2m.Path, notify []lwm2m.Node) {
	logrus.Infof("new notify from %s:%s", d.EndPoint, p.String())
	val, err := lwm2m.NodeGetAllResources(notify, p)
	if err != nil {
		logrus.Warnf("read val err (%v), (%v)", p.String(), err)
	} else {
		logrus.Infof("get val from (%v)", p)
		for k, v := range val {
			logrus.Infof("(%v), (%v)", k, v)
		}
	}
}

func onDeviceConn(d *lwm2m.Device) {
	ctx := context.Background()
	logrus.Infof("new device: %s", d.EndPoint)
	if d.HasObjectInstance(3, 0) {
		p, _ := lwm2m.NewPathFromString("3/0")
		nodes, err := d.Read(ctx, p)
		if err != nil {
			return
		}
		val, err := lwm2m.NodeGetAllResources(nodes, p)
		if err != nil {
			logrus.Warnf("read val err (%v), (%v)", p.String(), err)
		} else {
			logrus.Infof("get val from (%v)", p)
			for k, v := range val {
				logrus.Infof("(%v), (%v)", k, v)
			}
		}
	}
	for i := 0; i < len(objIds); i++ {
		if d.HasObjectWithInstance(objIds[i]) {
			d.AddObservation(lwm2m.NewObjectPath(objIds[i]), handleNotify)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("starting lwm2m server")
	lwm2m.GetRegistry().Append("custom")
	s := lwm2m.NewServer(
		lwm2m.WithOnNewDeviceConn(onDeviceConn),
		lwm2m.EnableUDPListener("udp6", ":5683"),
		lwm2m.EnableDTLSListener("udp6", ":5684", lwm2m.NewDummy()))
	go s.Serve(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	cancel()
	logrus.Info("Shutting down.")
}
