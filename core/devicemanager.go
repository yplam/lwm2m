package core

import "github.com/plgd-dev/go-coap/v3/mux"

type DeviceEvent int

const (
	Register DeviceEvent = iota
	Update
	Deregister
	Offline
)

type OnDeviceStateChangeFunc func(d *Device, e DeviceEvent, m *DeviceManager)

type DeviceManager interface {
	Register(ep string, lifetime int, version string, binding string,
		sms string, links []*CoreLink, conn mux.Conn) (*Device, error)
	Update(id string, lifetime int, binding string, sms string,
		links []*CoreLink, conn mux.Conn) error
	Deregister(id string) error
	GetDevice(id string) (*Device, error)
	GetDeviceByEP(ep string) (*Device, error)
	OnDeviceStateChange(f OnDeviceStateChangeFunc)
}

type deviceManager struct {
}

func (d deviceManager) Register(ep string, lifetime int, version string, binding string, sms string, links []*CoreLink, conn mux.Conn) (*Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceManager) Update(id string, lifetime int, binding string, sms string, links []*CoreLink, conn mux.Conn) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceManager) Deregister(id string) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceManager) GetDevice(id string) (*Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceManager) GetDeviceByEP(ep string) (*Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceManager) OnDeviceStateChange(f OnDeviceStateChangeFunc) {
	//TODO implement me
	panic("implement me")
}

func DefaultDeviceManager() DeviceManager {
	return &deviceManager{}
}
