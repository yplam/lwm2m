package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/yplam/lwm2m"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("starting lwm2m server")
	s := lwm2m.NewServer(
		lwm2m.EnableUDPListener("udp", ":5683"),
		lwm2m.EnableDTLSListener("udp", ":5684", lwm2m.NewDummy()))
	go s.Serve(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	cancel()
	logrus.Info("Shutting down.")
}
