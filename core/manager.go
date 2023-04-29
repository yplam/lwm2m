package core

import (
	"context"
	"errors"
	"github.com/pion/logging"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/yplam/lwm2m/encoding"
	"github.com/yplam/lwm2m/node"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrCoreEmptyEndpointClientName = errors.New("endpoint client name empty")
	ErrCoreLwm2mVersionNotSupport  = errors.New("lwm2m version not support")
	ErrIDNotFound                  = errors.New("id not found")
	ErrNotFound                    = errors.New("not found")
	ErrDeviceNotFound              = errors.New("device not found")
)

type DeviceEventType int

const (
	DeviceRegister DeviceEventType = iota
	DevicePostRegister
	DeviceUpdate
	DevicePostUpdate
	DeviceDeregister
)

func (e DeviceEventType) String() string {
	switch e {
	case DeviceRegister:
		return "DeviceRegister"
	case DevicePostRegister:
		return "DevicePostRegister"
	case DeviceUpdate:
		return "DeviceUpdate"
	case DevicePostUpdate:
		return "DevicePostUpdate"
	case DeviceDeregister:
		return "DeviceDeregister"
	default:
		return "Unknow"
	}
}

type DeviceEvent struct {
	EventType DeviceEventType
	Device    *Device
}

// OnDeviceStateChangeFunc call when a device state change
// events are fire in order and please do not block in this callback
type OnDeviceStateChangeFunc func(e DeviceEvent, m Manager)

type RegisterRequest struct {
	Ep          string
	Lifetime    int
	Version     string
	BindingMode Binding
	Queue       bool
	SmsNumber   *string
}

func NewRegisterRequest(queries []string) (req *RegisterRequest, err error) {
	req = &RegisterRequest{
		Ep:          "",
		Lifetime:    30, // use 30 as default value
		Version:     "",
		BindingMode: UdpBinding,
		Queue:       false,
		SmsNumber:   nil,
	}
	for _, val := range queries {
		sps := strings.Split(val, "=")
		if len(sps) != 2 {
			continue
		}
		switch sps[0] {
		case "ep":
			req.Ep = sps[1]
		case "lwm2m":
			req.Version = sps[1]
		case "lt":
			req.Lifetime, err = strconv.Atoi(sps[1])
		case "sms":
			req.SmsNumber = &sps[1]
		case "b":
			req.BindingMode = NewBinding(sps[1])
		default:
		}
	}
	if req.Ep == "" {
		err = ErrCoreEmptyEndpointClientName
	}
	if req.Version != "1.0" && req.Version != "1.1" {
		err = ErrCoreLwm2mVersionNotSupport
	}
	return
}

type UpdateRequest struct {
	Lifetime    *int
	BindingMode *Binding
	SmsNumber   *string
}

func NewUpdateRequest(queries []string) (req *UpdateRequest, err error) {
	req = &UpdateRequest{
		Lifetime:    nil,
		BindingMode: nil,
		SmsNumber:   nil,
	}
	for _, val := range queries {
		sps := strings.Split(val, "=")
		if len(sps) != 2 {
			continue
		}
		switch sps[0] {
		case "lt":
			var lt int
			if lt, err = strconv.Atoi(sps[1]); err == nil {
				req.Lifetime = &lt
			}
		case "sms":
			req.SmsNumber = &sps[1]
		case "b":
			binding := NewBinding(sps[1])
			req.BindingMode = &binding
		default:
		}
	}
	return
}

type Manager interface {
	Register(req *RegisterRequest, links []*encoding.CoreLink, conn mux.Conn) (*Device, error)
	// PostRegister call by transport when register response send
	PostRegister(id string)
	Update(id string, req *UpdateRequest, links []*encoding.CoreLink, conn mux.Conn) error
	PostUpdate(id string)
	Deregister(id string) error
	GetDevice(id string) (*Device, error)
	GetDeviceByEP(ep string) (*Device, error)
	OnDeviceStateChange(f OnDeviceStateChangeFunc)
}

type manager struct {
	ctx     context.Context
	lock    sync.RWMutex
	devices map[string]*Device
	epToID  map[string]string

	deviceStateChangeCb OnDeviceStateChangeFunc
	logger              logging.LeveledLogger

	eventChan chan DeviceEvent
}

func (d *manager) postEvent(dev *Device, event DeviceEventType) {
	if d.deviceStateChangeCb != nil {
		e := DeviceEvent{
			EventType: event,
			Device:    dev,
		}
		d.eventChan <- e
	}
}

func (d *manager) PostRegister(id string) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if dev, err := d.getDevice(id); err == nil {
		d.postEvent(dev, DevicePostRegister)
	}
}
func (d *manager) Update(id string, req *UpdateRequest, links []*encoding.CoreLink, conn mux.Conn) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	dev, err := d.getDevice(id)
	if err != nil {
		return ErrDeviceNotFound
	}
	if dev.conn.RemoteAddr().String() != conn.RemoteAddr().String() {
		dev.conn = conn
	}
	if req.Lifetime != nil {
		dev.Lifetime = *req.Lifetime
		dev.conn.SetContextValue(lifetimeCtxKey, time.Second*time.Duration(*req.Lifetime))
	}
	if req.BindingMode != nil {
		dev.BindingMode = *req.BindingMode
	}
	if req.SmsNumber != nil {
		dev.Sms = req.SmsNumber
	}
	if links != nil && len(links) > 0 {
		dev.ParseCoreLinks(links)
	}
	d.postEvent(dev, DeviceUpdate)
	return nil
}

func (d *manager) PostUpdate(id string) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if dev, err := d.getDevice(id); err == nil {
		d.postEvent(dev, DevicePostUpdate)
	}
}

func (d *manager) deregister(id string) error {
	dev, err := d.getDevice(id)
	if err != nil {
		return ErrDeviceNotFound
	}
	delete(d.devices, id)
	delete(d.epToID, dev.Endpoint)
	go dev.Close()
	d.postEvent(dev, DeviceDeregister)
	return nil
}

func (d *manager) Deregister(id string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.deregister(id)
}

func (d *manager) getDevice(id string) (*Device, error) {
	if dev, ok := d.devices[id]; ok {
		return dev, nil
	} else {
		return nil, ErrDeviceNotFound
	}
}

func (d *manager) GetDevice(id string) (*Device, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getDevice(id)
}

func (d *manager) getDeviceByEP(ep string) (*Device, error) {
	if id, ok := d.epToID[ep]; ok {
		return d.getDevice(id)
	} else {
		return nil, ErrIDNotFound
	}
}

func (d *manager) GetDeviceByEP(ep string) (*Device, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getDeviceByEP(ep)
}

func (d *manager) OnDeviceStateChange(f OnDeviceStateChangeFunc) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.deviceStateChangeCb = f
}

func (d *manager) Register(req *RegisterRequest, links []*encoding.CoreLink, conn mux.Conn) (*Device, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if dev, err := d.getDeviceByEP(req.Ep); err == nil {
		_ = d.deregister(dev.Id)
	}
	conn.SetContextValue(lifetimeCtxKey, time.Second*time.Duration(req.Lifetime))
	ctx, cancel := context.WithCancel(d.ctx)
	dev := &Device{
		ctx:         ctx,
		cancel:      cancel,
		Id:          d.generateRegId(),
		Endpoint:    req.Ep,
		Version:     req.Version,
		BindingMode: req.BindingMode,
		conn:        conn,
		Lifetime:    req.Lifetime,
		Sms:         req.SmsNumber,
		objs:        make(map[uint16]*node.Object),
		Manager:     d,
		obsChan:     make(chan observationEvent, 10),
	}
	if links != nil && len(links) > 0 {
		dev.ParseCoreLinks(links)
	}
	d.devices[dev.Id] = dev
	d.epToID[req.Ep] = dev.Id
	d.postEvent(dev, DeviceRegister)
	go dev.run()
	return dev, nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (d *manager) generateRegId() string {
	for {
		b := make([]byte, 5)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		if _, ok := d.devices[string(b)]; !ok {
			return string(b)
		}
	}
}

func (d *manager) run() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-d.ctx.Done():
			return
		case e := <-d.eventChan:
			if d.deviceStateChangeCb != nil {
				d.deviceStateChangeCb(e, d)
			}
		case <-ticker.C:
			d.logger.Debugf("tick")
		}
	}
}

type ManagerConfig struct {
	logger logging.LeveledLogger
	ctx    context.Context
}

func newManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		logger: nil,
		ctx:    context.Background(),
	}
}

type ManagerOption func(cfg *ManagerConfig)

func WithLogger(l logging.LeveledLogger) ManagerOption {
	return func(o *ManagerConfig) {
		o.logger = l
	}
}

func WithContext(ctx context.Context) ManagerOption {
	return func(o *ManagerConfig) {
		o.ctx = ctx
	}
}

func DefaultManager(opts ...ManagerOption) Manager {
	cfg := newManagerConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.logger == nil {
		lf := logging.NewDefaultLoggerFactory()
		cfg.logger = lf.NewLogger("device_manager")
	}
	dm := &manager{
		ctx:       cfg.ctx,
		devices:   make(map[string]*Device),
		epToID:    make(map[string]string),
		logger:    cfg.logger,
		eventChan: make(chan DeviceEvent, 1000),
	}
	go dm.run()
	return dm
}
