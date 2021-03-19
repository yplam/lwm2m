package server

import (
	"context"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/mux"
	"log"
	"lwm2m/corelink"
	"lwm2m/encoding"
	"lwm2m/model"
	"lwm2m/path"
	"strconv"
	"strings"
)

// A LWM2M Device connected to server
type Device struct {
	ID       string
	EndPoint string
	Version  string
	Lifetime int
	client   mux.Client
	Binding  string
	Sms      string
	Objs     map[uint16]*model.Object
}

type Observation = interface {
	Cancel(ctx context.Context) error
}

func (d *Device) Observe(p *path.Path, onMsg func(d *Device, notify []model.Node)) (Observation, error) {
	// 2 bytes length
	buf := make([]byte, 2)
	l, err := message.EncodeUint32(buf, uint32(message.AppLwm2mTLV))
	if err != nil {
		// should not happen
		return nil, err
	}
	return d.client.Observe(context.Background(), p.String(), func(notification *message.Message) {
		m, err := encoding.DecodeMessage(message.AppLwm2mTLV, notification.Body)
		if err != nil {
			return
		}
		// call onMsg first, it may use old shadow value
		onMsg(d, m)
		_ = d.updateValue(p, m...)
	}, message.Option{
		ID:    message.Accept,
		Value: buf[:l],
	})
}

func (d *Device) Write(p *path.Path, vals ...model.Node) {
	buf := make([]byte, 2)
	l, _ := message.EncodeUint32(buf, uint32(message.AppLwm2mTLV))
	msg, _ := encoding.EncodeMessage(message.AppLwm2mTLV, vals)
	_, _ = d.client.Put(context.Background(), p.String(), message.AppLwm2mTLV, msg,
		message.Option{
			ID:    message.Accept,
			Value: buf[:l],
		})
	_ = d.updateValue(p, vals...)
}

// Return last shadow value of path, which is store in server
func (d *Device) Value(p *path.Path) (model.Node, error) {
	return nil, nil
}

// Update the shadow value, which is store in server
func (d *Device) updateValue(p *path.Path, vals ...model.Node) error {
	return nil
}

func (d *Device) ParseCoreLinks(links []*corelink.CoreLink) {
	for _, v := range links {
		log.Printf("%v", v.Uri)
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
			obj = model.NewObject(objId)
		}
		if len(sps) == 2 {
			if insId, err := strconv.Atoi(sps[1]); err == nil {
				obj.Instances[uint16(insId)] = model.NewObjectInstance(uint16(insId))
			}
		}
		d.Objs[objId] = obj
	}
}
