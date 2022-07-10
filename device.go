package lwm2m

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/sirupsen/logrus"
)

// Device is a LWM2M Device connected to server
type Device struct {
	ID           string
	EndPoint     string
	Version      string
	Lifetime     int
	client       mux.Client
	Binding      string
	Sms          string
	Objs         map[uint16]Object
	S            *Server
	Observations map[Path]Observation
}

type ObserveFunc func(d *Device, p Path, notify []Node)

type Observation struct {
	o  mux.Observation
	cb ObserveFunc
}

var (
	_tlvAcceptOption     message.Option
	_tlvAcceptOptionOnce sync.Once
)

func _acceptOption() message.Option {
	_tlvAcceptOptionOnce.Do(func() {
		buf := make([]byte, 2)
		l, _ := message.EncodeUint32(buf, uint32(message.AppLwm2mTLV))
		_tlvAcceptOption = message.Option{
			ID:    message.Accept,
			Value: buf[:l],
		}
	})
	return _tlvAcceptOption
}

func (d *Device) AddObservation(p Path, onMsg ObserveFunc) {
	d.RemoveObservation(p)
	d.Observations[p] = Observation{
		o:  nil,
		cb: onMsg,
	}
}

func (d *Device) RemoveObservation(p Path) {
	if o, ok := d.Observations[p]; ok {
		if o.o != nil {
			o.o.Cancel(context.Background())
		}
		delete(d.Observations, p)
	}
}

func (d *Device) ProcessObservation() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for k, v := range d.Observations {
		if v.o == nil {
			no, err := d.client.Observe(ctx, k.String(), func(notification *message.Message) {
				if notification.Body == nil {
					return
				}
				m, err := DecodeMessage(message.AppLwm2mTLV, k, notification.Body)
				if err != nil {
					return
				}
				// call onMsg first, it may use old shadow value
				d.Observations[k].cb(d, k, m)
			}, _acceptOption())
			if err != nil {
				logrus.Warnf("observe %v error %v", k, err)
			} else {
				logrus.Infof("observe %v ok", k)
				v.o = no
				d.Observations[k] = v
				return
			}
		}
	}
}

func (d *Device) Read(ctx context.Context, p Path) ([]Node, error) {
	msg, err := d.client.Get(ctx, p.String(), _acceptOption())
	if err != nil {
		return nil, err
	}
	if msg.Body == nil {
		return nil, errors.New("empty body")
	}
	return DecodeMessage(message.AppLwm2mTLV, p, msg.Body)
}

func (d *Device) Write(ctx context.Context, p Path, vals ...Node) error {
	msg, _ := EncodeMessage(message.AppLwm2mTLV, vals)
	logrus.Debugf("write %v", msg)
	_, err := d.client.Put(ctx, p.String(), message.AppLwm2mTLV, msg,
		_acceptOption())
	return err
}

func (d *Device) onUpdate() {
	d.ProcessObservation()
}

func (d *Device) DeRegister() error {
	return d.S.DeRegister(d.ID)
}

func (d *Device) ParseCoreLinks(links []*CoreLink) {
	for _, v := range links {
		logrus.Infof("core link %v", v.Uri)
		sps := strings.Split(strings.Trim(v.Uri, "/"), "/")
		if len(sps) == 0 {
			continue
		}
		id, err := strconv.Atoi(sps[0])
		if err != nil {
			continue
		}
		objId := uint16(id)
		obj, ok := d.Objs[objId]
		if !ok {
			obj = NewObject(objId)
		}
		if len(sps) == 2 {
			if insId, err := strconv.Atoi(sps[1]); err == nil {
				obj.Instances[uint16(insId)] = NewObjectInstance(uint16(insId))
			}
		}
		d.Objs[objId] = obj
	}
}

func (d *Device) HasObject(id uint16) bool {
	if _, okay := d.Objs[id]; okay {
		return true
	}
	return false
}

func (d *Device) HasObjectWithInstance(id uint16) bool {
	if _, okay := d.Objs[id]; okay {
		return len(d.Objs[id].Instances) > 0
	}
	return false
}

func (d *Device) HasObjectInstance(id, iid uint16) bool {
	if _, okay := d.Objs[id]; !okay {
		return false
	}
	if _, okay := d.Objs[id].Instances[iid]; !okay {
		return false
	}
	return true
}
