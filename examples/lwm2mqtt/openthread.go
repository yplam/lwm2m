package main

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/ipv6"
)

func udpMulticast(ctx context.Context) {
	wpan, err := net.InterfaceByName("wpan0")
	if err != nil {
		logrus.Errorf("get network interface error %v", err)
		return
	}
	gip := net.ParseIP("ff03::fd")

	c, err := net.ListenPacket("udp6", "[::]:0")
	if err != nil {
		logrus.Errorf("listen error %v", err)
		return
	}
	defer c.Close()

	p := ipv6.NewPacketConn(c)
	group := &net.UDPAddr{IP: gip}
	err = p.JoinGroup(wpan, group)
	if err != nil {
		logrus.Errorf("join group error %v", err)
		return
	}
	<-ctx.Done()
	err = p.LeaveGroup(wpan, group)
	if err != nil {
		logrus.Errorf("leave group error %v", err)
		return
	}
}
