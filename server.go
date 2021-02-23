package lwm2m

import (
	"context"
	piondtls "github.com/pion/dtls/v2"
	"github.com/pion/logging"
	coapdtls "github.com/plgd-dev/go-coap/v2/dtls"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/go-coap/v2/net"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

var defaultServerOptions = serverOptions{
	ctx:   context.Background(),
	store: NewDummy(),
}

type ServerOption interface {
	apply(*serverOptions)
}

type Store interface {
	PSKIdentityFromEP([]byte) ([]byte, error)
	PSKFromIdentity([]byte) ([]byte, error)
}

type serverOptions struct {
	ctx   context.Context
	store Store
}

type Server struct {
	ctx           context.Context
	cancel        context.CancelFunc
	loggerFactory logging.LoggerFactory
	router        *mux.Router
	store         Store
	DeviceManage  *DeviceManager
	reg           *Registration
}

// Stop stops server without wait of ends Serve function.
func (s *Server) Stop() {
	s.cancel()
}

func (s *Server) ListenAndServeDTLS(network string, addr string) error {
	ds := NewDTLSServer(s.store)
	dtlsConfig := piondtls.Config{
		CipherSuites:         []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
		ExtendedMasterSecret: piondtls.DisableExtendedMasterSecret,
		PSK: func(id []byte) ([]byte, error) {
			return ds.PSK(id)
		},
		LoggerFactory: s.loggerFactory,
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithCancel(s.ctx)
		},
	}
	l, err := net.NewDTLSListener(network, addr, &dtlsConfig)
	if err != nil {
		return err
	}
	defer l.Close()

	s.reg.ValidateClientConn = ds.ValidateClientConn
	cs := coapdtls.NewServer(coapdtls.WithMux(s.router),
		coapdtls.WithKeepAlive(nil),
		coapdtls.WithOnNewClientConn(func(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
			//cc.SetContextValue(store.PSK_ID_HINT, dtlsConn.ConnectionState().IdentityHint)
			ds.OnNewClientConn(cc, dtlsConn)
		}))
	return cs.Serve(l)
}

func NewServer(opt ...ServerOption) *Server {
	opts := defaultServerOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	ctx, cancel := context.WithCancel(opts.ctx)

	loggerFactory := logging.NewDefaultLoggerFactory()
	loggerFactory.DefaultLogLevel = logging.LogLevelDebug

	dm := NewManager(ctx)
	m := mux.NewRouter()
	reg := NewRegistration(dm)
	_ = m.Handle("/rd", reg)
	_ = m.Handle("/rd/", reg)

	return &Server{
		ctx:           ctx,
		cancel:        cancel,
		loggerFactory: loggerFactory,
		router:        m,
		store:         opts.store,
		reg:           reg,
	}
}
