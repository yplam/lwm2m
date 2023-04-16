package server

import (
	"context"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/tcp"
	"github.com/plgd-dev/go-coap/v3/udp"
	"golang.org/x/sync/errgroup"
)

func ListenAndServe(handler mux.Handler, opts ...Option) error {
	return ListenAndServeWithContext(context.Background(), handler, opts...)
}

func ListenAndServeWithContext(ctx context.Context, handler mux.Handler, opts ...Option) error {
	cfg := newServeConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	eg, ctx := errgroup.WithContext(ctx)
	if (len(cfg.udpAddr) > 0) && (len(cfg.udpNetwork) > 0) {
		eg.Go(func() error {
			l, err := net.NewListenUDP(cfg.udpNetwork, cfg.udpAddr)
			if err != nil {
				return err
			}
			defer func() {
				if errC := l.Close(); errC != nil && err == nil {
					err = errC
				}
			}()
			s := udp.NewServer(options.WithContext(ctx), options.WithMux(handler))
			fmt.Printf("starting udp\n")
			return s.Serve(l)
		})
	}
	if (len(cfg.tcpAddr) > 0) && (len(cfg.tcpNetwork) > 0) {
		eg.Go(func() error {
			l, err := net.NewTCPListener(cfg.tcpNetwork, cfg.tcpAddr)
			if err != nil {
				return err
			}
			defer func() {
				if errC := l.Close(); errC != nil && err == nil {
					err = errC
				}
			}()
			s := tcp.NewServer(options.WithContext(ctx), options.WithMux(handler))
			fmt.Printf("starting tcp\n")
			return s.Serve(l)
		})
	}
	return eg.Wait()
}
