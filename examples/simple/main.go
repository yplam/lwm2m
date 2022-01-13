package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yplam/lwm2m"
	"golang.org/x/net/ipv6"
)

var (
	objIds = []uint16{3303, 3304}
)

func handleNotify(d *lwm2m.Device, p lwm2m.Path, notify []lwm2m.Node) {
	logrus.Infof("new notify from %s:%s", d.EndPoint, p.String())
	val, err := lwm2m.NodeGetAllResources(notify, p)
	if err != nil {
		logrus.Warnf("read val err (%v), (%v)", p.String(), err)
	} else {
		sp := lwm2m.NewResourcePath(3303, 0, 5700)
		if v, ok := val[sp]; ok {
			logrus.Info("...get from 3303/0/5700")
			spn := lwm2m.NewResourcePath(32769, 0, 26241)
			if nr, err := lwm2m.NewResource(spn, false, v.Data()); err == nil {
				go func(r *lwm2m.Resource) {
					<-time.After(time.Second)
					d.Write(spn, r)
				}(nr)
			}

		}
		logrus.Infof("get val from (%v)", p)
		for k, v := range val {
			logrus.Infof("(%v), (%v)", k, v)
		}
	}
}

func onDeviceConn(d *lwm2m.Device) {
	logrus.Infof("new device: %s", d.EndPoint)
	for i := 0; i < len(objIds); i++ {
		if d.HasObject(objIds[i]) {
			_, err := d.Observe(lwm2m.NewObjectPath(objIds[i]), handleNotify)
			if err != nil {
				logrus.Errorf("can not observe %v, %v", objIds[i], err)
			}
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("starting lwm2m server")
	lwm2m.GetRegistry().Append("custom")
	s := lwm2m.NewServer(
		lwm2m.WithOnNewDeviceConn(onDeviceConn),
		lwm2m.EnableUDPListener("udp6", ":5683"),
		lwm2m.EnableDTLSListener("udp6", ":5684", lwm2m.NewDummy()))
	go s.Serve(ctx)
	go udpMulticast(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	cancel()
	logrus.Info("Shutting down.")
}

func udpMulticast(ctx context.Context) {
	wpan, err := net.InterfaceByName("wpan0")
	if err != nil {
		logrus.Errorf("get network interface error %v", err)
		return
	}
	group := net.ParseIP("ff03::fd")

	c, err := net.ListenPacket("udp6", "[::]:0")
	if err != nil {
		logrus.Errorf("listen error %v", err)
		return
	}
	defer c.Close()

	p := ipv6.NewPacketConn(c)
	if err := p.JoinGroup(wpan, &net.UDPAddr{IP: group}); err != nil {
		logrus.Errorf("join error %v", err)
		return
	}
	<-ctx.Done()
}
