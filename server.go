package lwm2m

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	piondtls "github.com/pion/dtls/v2"
	"github.com/pion/logging"
	coapdtls "github.com/plgd-dev/go-coap/v2/dtls"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/go-coap/v2/net"
	"github.com/plgd-dev/go-coap/v2/udp"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

var (
	ErrIDNotFound     = errors.New("id not found")
	ErrDeviceNotFound = errors.New("device not found")
)

type Store interface {
	// PSKIdentityFromEP
	PSKIdentityFromEP([]byte) ([]byte, error)
	PSKFromIdentity([]byte) ([]byte, error)
}

type OnNewDeviceConnFunc func(d *Device)

type Server struct {
	log             logging.LeveledLogger
	LoggerFactory   logging.LoggerFactory
	store           Store
	lock            sync.RWMutex
	devices         map[string]*Device
	epToID          map[string]string
	onNewDeviceConn OnNewDeviceConnFunc
	udpNetwork      string
	udpAddr         string
	dtlsNetwork     string
	dtlsAddr        string
}

func (s *Server) DeRegister(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.deRegister(id)
}

func (s *Server) deRegister(id string) error {
	d := s.getByID(id)
	if d == nil {
		return ErrDeviceNotFound
	}
	if d.client != nil {
		d.client.Close()
	}
	delete(s.devices, id)
	delete(s.epToID, d.EndPoint)
	return nil
}

func (s *Server) Register(ep string, lifetime int, version string, binding string,
	sms string, links []*CoreLink, client mux.Client) (*Device, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if id, err := s.getIdByEndpoint(ep); err == nil {
		_ = s.deRegister(id)
	}
	d := &Device{
		ID:           s.generateRegId(),
		EndPoint:     ep,
		Version:      version,
		Lifetime:     lifetime,
		client:       client,
		Binding:      binding,
		Sms:          sms,
		Objs:         make(map[uint16]Object),
		S:            s,
		Observations: make(map[Path]Observation),
	}
	d.ParseCoreLinks(links)
	s.devices[d.ID] = d
	s.epToID[ep] = d.ID
	return d, nil
}

func (s *Server) PostRegister(id string) {
	if s.onNewDeviceConn != nil {
		go func(id string) {
			time.Sleep(time.Second)
			d := s.GetByID(id)
			if d == nil {
				return
			}
			s.onNewDeviceConn(d)
			s.log.Debug("after device register")
		}(id)
	}
}

func (s *Server) PostUpdate(id string) {

}

func (s *Server) Update(id string, lifetime int, binding string, sms string,
	links []*CoreLink, client mux.Client) error {
	d, ok := s.devices[id]
	if !ok {
		return ErrDeviceNotFound
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
	d.client = client
	d.onUpdate()
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
	if id, ok := s.epToID[ep]; ok {
		return id, nil
	}
	return "", ErrIDNotFound
}

func (s *Server) GetByID(id string) *Device {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.getByID(id)
}

func (s *Server) getByID(id string) *Device {
	if d, ok := s.devices[id]; ok {
		return d
	}
	return nil
}

func (s *Server) GetByEndpoint(ep string) *Device {
	s.lock.RLock()
	defer s.lock.RUnlock()
	id, err := s.getIdByEndpoint(ep)
	if err != nil {
		return nil
	}
	return s.getByID(id)
}

func (s *Server) Serve(c context.Context) {
	ctx, cancel := context.WithCancel(c)
	r := mux.NewRouter()
	reg := NewRegistration(s, s.LoggerFactory.NewLogger("registration"))
	_ = r.Handle("/rd", reg)
	_ = r.Handle("/rd/", reg)

	signal := make(chan struct{}, 3)
	go s.ListenAndServeUDP(ctx, signal, r)
	go s.ListenAndServeDTLS(ctx, signal, r)

	select {
	case <-ctx.Done():
	case <-signal:
	}
	cancel()
}

func (s *Server) ListenAndServeUDP(ctx context.Context, c chan struct{}, r *mux.Router) {
	if len(s.udpNetwork) == 0 {
		return
	}
	s.log.Info("listening udp")
	l, err := net.NewListenUDP(s.udpNetwork, s.udpAddr)
	if err != nil {
		s.log.Errorf("listen udp error (%v)", err)
		c <- struct{}{}
		return
	}
	defer l.Close()
	us := udp.NewServer(
		udp.WithTransmission(10*time.Second, 10*time.Second, 4),
		udp.WithMux(r))
	go func() {
		s.log.Errorf("udp server stopped (%v)", us.Serve(l))
		c <- struct{}{}
	}()
	<-ctx.Done()
	us.Stop()
}

func (s *Server) ListenAndServeDTLS(ctx context.Context, c chan struct{}, r *mux.Router) {
	if len(s.dtlsNetwork) == 0 {
		return
	}
	if s.store == nil {
		s.log.Errorf("can not use dtls without store")
		return
	}
	s.log.Info("listening dtls")
	dc := NewDTLSConnector(s.store)
	dtlsConfig := piondtls.Config{
		CipherSuites:         []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
		ExtendedMasterSecret: piondtls.DisableExtendedMasterSecret,
		PSK: func(id []byte) ([]byte, error) {
			return dc.psk(id)
		},
		LoggerFactory: s.LoggerFactory,
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithCancel(ctx)
		},
	}

	l, err := net.NewDTLSListener(s.dtlsNetwork, s.dtlsAddr, &dtlsConfig)
	if err != nil {
		s.log.Errorf("listen dtls error (%v)", err)
		c <- struct{}{}
		return
	}
	defer l.Close()

	cs := coapdtls.NewServer(coapdtls.WithMux(r),
		//coapdtls.WithKeepAlive(0,0, nil),
		coapdtls.WithOnNewClientConn(func(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
			dc.onNewClientConn(cc, dtlsConn)
		}))

	go func() {
		s.log.Errorf("dtls server stopped (%v)", cs.Serve(l))
		c <- struct{}{}
	}()
	<-ctx.Done()
	cs.Stop()
}

func NewServer(opt ...ServerOption) *Server {
	s := &Server{
		LoggerFactory: NewDefaultLoggerFactory(),
		devices:       make(map[string]*Device),
		epToID:        make(map[string]string),
	}
	for _, o := range opt {
		o(s)
	}
	s.log = s.LoggerFactory.NewLogger("server")
	return s
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
