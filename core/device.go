package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/yplam/lwm2m/encoding"
	"github.com/yplam/lwm2m/node"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Binding string

var (
	UdpBinding   Binding = "U"
	TcpBinding   Binding = "T"
	SmsBinding   Binding = "S"
	NonIpBinding Binding = "N"
)

func NewBinding(b string) Binding {
	switch b {
	case "T":
		return TcpBinding
	case "S":
		return SmsBinding
	case "N":
		return NonIpBinding
	default:
		return UdpBinding
	}
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

type observationEvent struct {
	p node.Path
	n []node.Node
}

type Device struct {
	ctx         context.Context
	cancel      context.CancelFunc
	Id          string
	Endpoint    string
	Version     string
	BindingMode Binding
	conn        mux.Conn
	Lifetime    int
	Sms         *string
	Manager     Manager

	objLock sync.RWMutex
	objs    map[uint16]*node.Object

	observations sync.Map //map[node.Path]Observation
	obsChan      chan observationEvent
}

func (d *Device) ParseCoreLinks(links []*encoding.CoreLink) {
	objs := make(map[uint16]*node.Object)
	for _, v := range links {
		sps := strings.Split(strings.Trim(v.Uri, "/"), "/")
		if len(sps) == 0 {
			continue
		}
		id, err := strconv.Atoi(sps[0])
		if err != nil {
			continue
		}
		obj, ok := objs[objId]
		if !ok {
			obj = node.NewObject(objId)
			objs[objId] = obj
		}
		if len(sps) == 2 {
			if insId, err := strconv.Atoi(sps[1]); err == nil {
				obj.Instances[uint16(insId)] = node.NewObjectInstance(uint16(insId))
			}
		}
	}
	d.objLock.Lock()
	d.objs = objs
	d.objLock.Unlock()
}

func (d *Device) ObserveSync(p node.Path, onMsg ObserveFunc) error {
	_ = d.CancelObserve(p)
	no, err := d.processObservation(p)
	if err != nil {
		return err
	}
	d.observations.Store(p, Observation{
		o:  no,
		cb: onMsg,
	})
	return nil
}

func (d *Device) Observe(p node.Path, onMsg ObserveFunc) error {
	_ = d.CancelObserve(p)
	d.observations.Store(p, Observation{
		o:  nil,
		cb: onMsg,
	})
	return nil
}

func (d *Device) ObserveObject(p node.Path, onMsg ObserveObjectFunc) error {
	if !p.IsObject() {
		return node.ErrPathInvalidValue
	}
	return d.Observe(p, wrapObserveObjectFunc(onMsg))
}

func (d *Device) ObserveResource(p node.Path, onMsg ObserveResourceFunc) error {
	if !p.IsResource() {
		return node.ErrPathInvalidValue
	}
	return d.Observe(p, wrapObserveResourceFunc(onMsg))
}

func (d *Device) CancelObserve(p node.Path) error {
	v, ok := d.observations.LoadAndDelete(p)
	if !ok {
		return ErrNotFound
	}
	o := v.(Observation)
	if o.o != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Lifetime)*time.Second)
		go func() {
			o.o.Cancel(ctx)
			defer cancel()
		}()
	}
	return nil
}

func (d *Device) processObservation(k node.Path) (mux.Observation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Lifetime)*time.Second)
	defer cancel()
	return d.conn.Observe(ctx, k.String(), func(notification *pool.Message) {
		if notification.Body() == nil {
			return
		}
		nodes, err := node.DecodeMessage(k, notification)
		if err != nil {
			fmt.Printf("decode error %v\n", err)
			return
		}
		d.obsChan <- observationEvent{
			p: k,
			n: nodes,
		}
	}, _acceptOption())
}

func (d *Device) String() string {
	var b strings.Builder
	b.WriteString("Device {")
	b.WriteString(fmt.Sprintf("id: %v, ", d.Id))
	b.WriteString(fmt.Sprintf("ep: %v, ", d.Endpoint))
	b.WriteString(fmt.Sprintf("version: %v, ", d.Version))
	b.WriteString(fmt.Sprintf("bind: %v, ", d.BindingMode))
	b.WriteString(fmt.Sprintf("lifetime: %v ", d.Lifetime))
	b.WriteString("}")
	return b.String()
}

func (d *Device) Close() {
	d.observations.Range(func(key, value any) bool {
		_ = d.CancelObserve(key.(node.Path))
		return true
	})
	d.cancel()
}

func (d *Device) initOrUpdateObservation() {
	d.observations.Range(func(key, value any) bool {
		k := key.(node.Path)
		v := value.(Observation)
		if v.o == nil || v.o.Canceled() {
			no, err := d.processObservation(k)
			if err == nil {
				v.o = no
				d.observations.Store(key, v)
			}
		}
		return true
	})
}

func (d *Device) run() {
	<-time.After(time.Second)
	d.initOrUpdateObservation()
	for {
		select {
		case <-time.After(time.Duration(d.Lifetime) * time.Second):
			d.initOrUpdateObservation()
		case <-d.ctx.Done():
			d.Close()
			return
		case e := <-d.obsChan:

			val, ok := d.observations.Load(e.p)
			if ok {
				o := val.(Observation)
				go func() {
					o.cb(d, e.p, e.n)
					d.updateState(e.p, e.n)
				}()
			}
		}
	}
}

// updateState update Device stage storage
func (d *Device) updateState(k node.Path, m []node.Node) {
	d.objLock.Lock()
	defer d.objLock.Unlock()
}

func (d *Device) HasObject(id uint16) bool {
	d.objLock.RLock()
	defer d.objLock.RUnlock()
	if _, okay := d.objs[id]; okay {
		return true
	}
	return false
}

func (d *Device) HasObjectWithInstance(id uint16) bool {
	d.objLock.RLock()
	defer d.objLock.RUnlock()
	if _, okay := d.objs[id]; okay {
		return len(d.objs[id].Instances) > 0
	}
	return false
}

func (d *Device) HasObjectInstance(id, iid uint16) bool {
	d.objLock.RLock()
	defer d.objLock.RUnlock()
	if _, okay := d.objs[id]; !okay {
		return false
	}
	if _, okay := d.objs[id].Instances[iid]; !okay {
		return false
	}
	return true
}

func (d *Device) Read(ctx context.Context, p node.Path) ([]node.Node, error) {
	msg, err := d.conn.Get(ctx, p.String(), _acceptOption())
	if err != nil {
		return nil, err
	}
	if msg.Body() == nil {
		return nil, errors.New("empty body")
	}
	return node.DecodeMessage(p, msg)
}

func (d *Device) ReadObject(ctx context.Context, p node.Path) (*node.Object, error) {
	nodes, err := d.Read(ctx, p)
	if err != nil {
		return nil, err
	}
	return node.GetObjectByPath(nodes, p)
}

func (d *Device) ReadResource(ctx context.Context, p node.Path) (*node.Resource, error) {
	nodes, err := d.Read(ctx, p)
	if err != nil {
		return nil, err
	}
	return node.GetResourceByPath(nodes, p)
}

func (d *Device) Write(ctx context.Context, p node.Path, val ...node.Node) error {
	msg, err := node.EncodeMessage(message.AppLwm2mTLV, val)
	if err != nil {
		return err
	}
	resp, err := d.conn.Put(ctx, p.String(), message.AppLwm2mTLV, msg, _acceptOption())
	if err != nil {
		return err
	}
	if resp.Code() != codes.Changed {
		return ErrUnexpectedResponseCode
	}
	return nil
}

func (d *Device) WriteResource(ctx context.Context, p node.Path, val *node.Resource) error {
	return d.Write(ctx, p, val)
}

func (d *Device) WriteObjectInstance(ctx context.Context, p node.Path, val *node.ObjectInstance) error {
	return d.Write(ctx, p, val)
}

func (d *Device) Discover(ctx context.Context, p node.Path) ([]*encoding.CoreLink, error) {
	buf := make([]byte, 2)
	l, _ := message.EncodeUint32(buf, uint32(message.AppLinkFormat))
	r, err := d.conn.Get(ctx, p.String(), message.Option{
		ID:    message.Accept,
		Value: buf[:l],
	})
	if err != nil {
		return nil, err
	}
	links := make([]*encoding.CoreLink, 0)
	if r.Body() != nil {
		if b, err2 := io.ReadAll(r.Body()); err2 == nil {
			links, _ = encoding.CoreLinksFromString(string(b))
		}
	}
	return links, nil
}

func (d *Device) Execute(ctx context.Context, p node.Path) error {
	if !p.IsResource() {
		return node.ErrPathInvalidValue
	}
	resp, err := d.conn.Post(ctx, p.String(), message.AppLwm2mTLV, nil)
	if err != nil {
		return err
	}
	if resp.Code() != codes.Changed {
		return ErrUnexpectedResponseCode
	}
	return nil
}

func (d *Device) Create(ctx context.Context, p node.Path, val *node.Object) error {
	if !p.IsObject() {
		return node.ErrPathInvalidValue
	}
	msg, err := node.EncodeMessage(message.AppLwm2mTLV, []node.Node{val})
	if err != nil {
		return err
	}
	resp, err := d.conn.Post(ctx, p.String(), message.AppLwm2mTLV, msg)
	if err != nil {
		return err
	}
	if resp.Code() != codes.Created {
		return ErrUnexpectedResponseCode
	}
	return nil
}

func (d *Device) Delete(ctx context.Context, p node.Path) error {
	if !p.IsObjectInstance() {
		return node.ErrPathInvalidValue
	}
	resp, err := d.conn.Delete(ctx, p.String())
	if err != nil {
		return err
	}
	if resp.Code() != codes.DELETE {
		return ErrUnexpectedResponseCode
	}
	return nil
}
