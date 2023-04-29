package server

import (
	"context"
	"github.com/pion/logging"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/tcp"
	"github.com/plgd-dev/go-coap/v3/udp"
	udpClient "github.com/plgd-dev/go-coap/v3/udp/client"
	"github.com/yplam/lwm2m/core"
	"golang.org/x/sync/errgroup"
	"time"
)

func DefaultRouter() *mux.Router {
	return mux.NewRouter()
}

func ListenAndServe(router *mux.Router, opts ...Option) error {
	return ListenAndServeWithContext(context.Background(), router, opts...)
}

func ListenAndServeWithContext(ctx context.Context, router *mux.Router, opts ...Option) error {
	cfg := newServeConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.logger == nil {
		lf := logging.NewDefaultLoggerFactory()
		cfg.logger = lf.NewLogger("server")
	}
	router.DefaultHandleFunc(func(w mux.ResponseWriter, r *mux.Message) {
		if r.Code() == codes.Empty && r.Type() == message.Reset && len(r.Token()) == 0 && len(r.Options()) == 0 && r.Body() == nil {
			return
		}
		if obs, err := r.Observe(); err == nil && obs > 0 {
			msg := w.Conn().AcquireMessage(r.Context())
			defer w.Conn().ReleaseMessage(msg)
			msg.SetMessageID(r.MessageID())
			msg.SetType(message.Reset)
			msg.SetToken(r.Token())
			resp, err := w.Conn().Do(msg)
			if err == nil {
				defer w.Conn().ReleaseMessage(resp)
			}
		}
		cfg.logger.Debugf("request not handle %v", r)
		w.SetResponse(codes.NotFound, message.TextPlain, nil)
	})

	eg, ctx := errgroup.WithContext(ctx)
	if (len(cfg.udpAddr) > 0) && (len(cfg.udpNetwork) > 0) {
		eg.Go(func() error {
			l, err := net.NewListenUDP(cfg.udpNetwork, cfg.udpAddr)
			if err != nil {
				return err
			}
			go func() {
				select {
				case <-ctx.Done():
					l.Close()
				}
			}()
			defer func() {
				if errC := l.Close(); errC != nil && err == nil {
					err = errC
				}
				cfg.logger.Info("Udp server stop")
			}()
			s := udp.NewServer(options.WithContext(ctx),
				options.WithMux(router),
				options.WithTransmission(1, time.Second*30, 4),
				core.WithInactivityMonitor(func(cc *udpClient.Conn) {
					cfg.logger.Infof("inactive %v", cc.RemoteAddr())
					cc.Close()
				}),
			)
			cfg.logger.Info("Starting udp server")
			return s.Serve(l)
		})
	}
	if (len(cfg.tcpAddr) > 0) && (len(cfg.tcpNetwork) > 0) {
		eg.Go(func() error {
			l, err := net.NewTCPListener(cfg.tcpNetwork, cfg.tcpAddr)
			if err != nil {
				return err
			}
			go func() {
				select {
				case <-ctx.Done():
					l.Close()
				}
			}()
			defer func() {
				if errC := l.Close(); errC != nil && err == nil {
					err = errC
				}
				cfg.logger.Info("Tcp server stop")
			}()
			s := tcp.NewServer(options.WithContext(ctx), options.WithMux(router))
			cfg.logger.Info("Starting tcp server")
			return s.Serve(l)
		})
	}
	return eg.Wait()
}
