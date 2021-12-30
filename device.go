package lwm2m

import (
	"context"
	"strconv"
	"strings"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/sirupsen/logrus"
)

// Device is a LWM2M Device connected to server
type Device struct {
	ID       string
	EndPoint string
	Version  string
	Lifetime int
	client   mux.Client
	Binding  string
	Sms      string
	Objs     map[uint16]*Object
}

type Observation = interface {
	Cancel(ctx context.Context) error
}

func (d *Device) Observe(p Path, onMsg func(d *Device, p Path, notify []Node)) (Observation, error) {
	// 2 bytes length
	buf := make([]byte, 2)
	l, err := message.EncodeUint32(buf, uint32(message.AppLwm2mTLV))
	if err != nil {
		// should not happen
		return nil, err
	}
	return d.client.Observe(context.Background(), p.String(), func(notification *message.Message) {
		m, err := DecodeMessage(message.AppLwm2mTLV, p, notification.Body)
		if err != nil {
			return
		}
		// call onMsg first, it may use old shadow value
		go onMsg(d, p, m)
		_ = d.updateValue(p, m...)
	}, message.Option{
		ID:    message.Accept,
		Value: buf[:l],
	})
}

func (d *Device) Write(p Path, vals ...Node) {
	buf := make([]byte, 2)
	l, _ := message.EncodeUint32(buf, uint32(message.AppLwm2mTLV))
	msg, _ := EncodeMessage(message.AppLwm2mTLV, vals)
	_, _ = d.client.Put(context.Background(), p.String(), message.AppLwm2mTLV, msg,
		message.Option{
			ID:    message.Accept,
			Value: buf[:l],
		})
	_ = d.updateValue(p, vals...)
}

// Value Return last shadow value of Path
func (d *Device) Value(p Path) (Node, error) {
	return nil, nil
}

func (d *Device) updateValue(p Path, vals ...Node) error {
	return nil
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
