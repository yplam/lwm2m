package server

import (
	"context"
	"errors"
	piondtls "github.com/pion/dtls/v2"
	"github.com/pion/logging"
	coapdtls "github.com/plgd-dev/go-coap/v2/dtls"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/go-coap/v2/net"
	"github.com/plgd-dev/go-coap/v2/udp/client"
	"log"
	"lwm2m"
	"lwm2m/corelink"
	"lwm2m/model"
	"math/rand"
	"sync"
	"time"
)

var defaultServerOptions = serverOptions{
	store:    lwm2m.NewDummy(),
	registry: model.NewDefaultRegistry(),
}

type ServerOption interface {
	apply(*serverOptions)
}

type Store interface {
	PSKIdentityFromEP([]byte) ([]byte, error)
	PSKFromIdentity([]byte) ([]byte, error)
}

type serverOptions struct {
	store           Store
	registry        *model.Registry
	onNewDeviceConn OnNewDeviceConnFunc
}

type OnNewDeviceConnFunc = func(d *Device)

type Server struct {
	ctx             context.Context
	cancel          context.CancelFunc
	loggerFactory   logging.LoggerFactory
	store           Store
	lock            sync.RWMutex
	devices         map[string]*Device
	epToID          map[string]string
	registry        *model.Registry
	onNewDeviceConn OnNewDeviceConnFunc
}

func (s *Server) DeRegister(id string) error {
	old := s.GetByID(id)
	if old == nil {
		return errors.New("device not found")
	}
	s.lock.Lock()
	delete(s.devices, id)
	delete(s.epToID, old.EndPoint)
	s.lock.Unlock()
	return nil
}

func (s *Server) Register(ep string, lifetime int, version string, binding string,
	sms string, links []*corelink.CoreLink, client mux.Client) (*Device, error) {
	if id, err := s.getIdByEndpoint(ep); err == nil {
		_ = s.DeRegister(id)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	d := &Device{
		ID:       s.generateRegId(),
		EndPoint: ep,
		Version:  version,
		Lifetime: lifetime,
		client:   client,
		Binding:  binding,
		Sms:      sms,
		Objs:     make(map[uint16]*model.Object),
	}
	d.ParseCoreLinks(links)
	s.devices[d.ID] = d
	s.epToID[ep] = d.ID
	return d, nil
}

func (s *Server) PostRegister(id string) {
	d := s.GetByID(id)
	if d == nil {
		return
	}
	if s.onNewDeviceConn != nil {
		s.onNewDeviceConn(d)
	}
	log.Println("after device register")
}

func (s *Server) PostUpdate(id string) {

}

func (s *Server) Update(id string, lifetime int, binding string, sms string,
	links []*corelink.CoreLink) error {
	d, ok := s.devices[id]
	if !ok {
		return errors.New("device not found")
	}
	if lifetime > 0 {
		d.Lifetime = lifetime
	}
	if len(binding) > 0 {
		d.Binding = binding
	}
	if len(sms) > 0 {
		d.Sms = sms
	}
	if links != nil {
		d.ParseCoreLinks(links)
	}
	return nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (s *Server) generateRegId() string {
	for {
		b := make([]byte, 5)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		if _, ok := s.devices[string(b)]; !ok {
			return string(b)
		}
	}
}

func (s *Server) getIdByEndpoint(ep string) (string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if id, ok := s.epToID[ep]; ok {
		return id, nil
	}
	return "", errors.New("id not found")
}

func (s *Server) GetByID(id string) *Device {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if d, ok := s.devices[id]; ok {
		return d
	}
	return nil
}

func (s *Server) GetByEndpoint(ep string) *Device {
	id, err := s.getIdByEndpoint(ep)
	if err != nil {
		return nil
	}
	return s.GetByID(id)
}

func (s *Server) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

// Stop stops server without wait of ends Serve function.
func (s *Server) Stop() {
	s.cancel()
}

func (s *Server) ListenAndServeDTLS(network string, addr string) error {
	dc := NewDTLSConnector(s.store)
	dtlsConfig := piondtls.Config{
		CipherSuites:         []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
		ExtendedMasterSecret: piondtls.DisableExtendedMasterSecret,
		PSK: func(id []byte) ([]byte, error) {
			return dc.psk(id)
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

	m := mux.NewRouter()
	reg := newRegistration(s)
	_ = m.Handle("/rd", reg)
	_ = m.Handle("/rd/", reg)
	reg.ValidateClientConn = dc.validateClientConn

	cs := coapdtls.NewServer(coapdtls.WithMux(m),
		//coapdtls.WithKeepAlive(0,0, nil),
		coapdtls.WithOnNewClientConn(func(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
			dc.onNewClientConn(cc, dtlsConn)
		}))
	return cs.Serve(l)
}

func NewServer(opt ...ServerOption) *Server {
	opts := defaultServerOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	ctx, cancel := context.WithCancel(context.Background())

	loggerFactory := logging.NewDefaultLoggerFactory()
	loggerFactory.DefaultLogLevel = logging.LogLevelDebug

	return &Server{
		ctx:             ctx,
		cancel:          cancel,
		loggerFactory:   loggerFactory,
		store:           opts.store,
		devices:         make(map[string]*Device),
		epToID:          make(map[string]string),
		registry:        opts.registry,
		onNewDeviceConn: opts.onNewDeviceConn,
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
