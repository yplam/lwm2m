package device

import (
	"errors"
	"github.com/plgd-dev/go-coap/v2/mux"
	"math/rand"
	"sync"
	"time"
)

type Store interface {
}

type Manager struct {
	lock    sync.RWMutex
	devices map[string]*Device
	epToID  map[string]string
}

type Device struct {
	ID        string
	EndPoint  string
	Version   string
	Lifetime  int
	Client    mux.Client
	Binding   string
	CoreLinks []*CoreLink
	Sms       string
}

func (m *Manager) DeRegister(id string) error {
	old := m.GetByID(id)
	if old == nil {
		return errors.New("device not found")
	}
	m.lock.Lock()
	delete(m.devices, id)
	delete(m.epToID, old.EndPoint)
	m.lock.Unlock()
	return nil
}

func (m *Manager) Register(ep string, lifetime int, version string, binding string,
	sms string, links []*CoreLink, client mux.Client) (*Device, error) {
	if id, err := m.GetIdByEndpoint(ep); err == nil {
		_ = m.DeRegister(id)
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	d := &Device{
		ID:        m.generateRegId(),
		EndPoint:  ep,
		Version:   version,
		Lifetime:  lifetime,
		Client:    client,
		Sms:       sms,
		CoreLinks: links,
		Binding:   binding,
	}
	m.devices[d.ID] = d
	m.epToID[ep] = d.ID
	return d, nil
}

func (m *Manager) Update(id string, lifetime int, binding string, sms string,
	links []*CoreLink) error {
	d, ok := m.devices[id]
	if  !ok {
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
		d.CoreLinks = links
	}
	return nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (m *Manager) generateRegId() string {
	for {
		b := make([]byte, 5)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		if _, ok := m.devices[string(b)]; !ok {
			return string(b)
		}
	}
}

func (m *Manager) GetIdByEndpoint(ep string) (string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if id, ok := m.epToID[ep]; ok {
		return id, nil
	}
	return "", errors.New("id not found")
}

func (m *Manager) GetByID(id string) *Device {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if d, ok := m.devices[id]; ok {
		return d
	}
	return nil
}

func NewManager() *Manager {
	return &Manager{
		devices: make(map[string]*Device),
		epToID:  make(map[string]string),
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
