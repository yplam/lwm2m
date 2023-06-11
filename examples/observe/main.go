package main

import (
	"context"
	"github.com/pion/logging"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/sirupsen/logrus"
	"github.com/yplam/lwm2m/core"
	"github.com/yplam/lwm2m/encoding"
	"github.com/yplam/lwm2m/node"
	"github.com/yplam/lwm2m/registration"
	"github.com/yplam/lwm2m/server"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"time"
	"unicode"
)

var lf = NewDefaultLoggerFactory()
var glog = lf.NewLogger("main")

var hadObserve = false
var createdByServer = false

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		glog.Debugf("ClientAddress %v, %v\n", w.Conn().RemoteAddr(), r.String())
		if body, _ := r.Message.ReadBody(); body != nil {
			r.Message.Body().Seek(0, io.SeekStart)
			bodyStr := string(body)
			if isASCII(bodyStr) {
				glog.Debugf("MessageBody %v", bodyStr)
			} else {
				glog.Debugf("MessageBody [% x]", body)
			}
		}
		next.ServeCOAP(w, r)
	})
}

func observeButtonObject(device *core.Device) (err error) {
	p, _ := node.NewPathFromString("/3347/0")
	if err = device.Observe(p, onObserve); err != nil {
		glog.Warnf("Observe error %v", err)
		return
	}

	op, _ := node.NewPathFromString("/3347")
	if err = device.ObserveObject(op, onObserveObject); err != nil {
		glog.Warnf("Observe object error %v", err)
		return
	}

	pr, _ := node.NewPathFromString("/3347/0/5501")
	if err = device.ObserveResource(pr, onObserveResource); err != nil {
		glog.Warnf("Observe resource error %v", err)
		return
	}
	return
}
func onDeviceStateChange(e core.DeviceEvent, m core.Manager) {
	device := e.Device
	switch e.EventType {
	case core.DevicePostRegister:
		glog.Infof("Post register")
		if device.HasObjectWithInstance(3347) {
			if err := observeButtonObject(device); err == nil {
				hadObserve = true
			}
		} else if device.HasObject(3347) {
			glog.Infof("Create button object instance")
			p, _ := node.NewPathFromString("/3347")
			obi := node.NewObjectInstance(0)
			if true { // Change this condition to create empty object instance
				pr, _ := node.NewPathFromString("/3347/0/5750")
				val := encoding.NewTlv(encoding.TlvSingleResource, 5750, "Created by server")
				res, _ := node.NewSingleResource(pr, val)
				obi.SetResource(5750, res)
			}
			if err := device.Create(context.Background(), p, obi); err != nil {
				glog.Warnf("Create object instance error %v", err)
			}
		}
	case core.DevicePostUpdate:
		if device.HasObjectWithInstance(3347) && createdByServer {
			// Delete object instance created by server
			glog.Infof("Delete object instance created by server")
			pi, _ := node.NewPathFromString("/3347/0")
			if err := device.Delete(context.Background(), pi); err != nil {
				glog.Infof("Delete object instance error %v", err)
			}

		} else if device.HasObjectWithInstance(3347) && hadObserve == false {
			glog.Infof("Observe button instance created by server")
			if err := observeButtonObject(device); err == nil {
				hadObserve = true
				createdByServer = true
				glog.Infof("create ok")
			}
		}

	default:

	}
}

func onObserve(d *core.Device, p node.Path, notify []node.Node) {
	rp, _ := node.NewPathFromString("/3347/0/5501")
	if !rp.IsChildOfOrEq(p) {
		return
	}
	if r, err := node.GetResourceByPath(notify, rp); err == nil {
		glog.Infof("On observe %v", r)
	}
}

func onObserveObject(d *core.Device, p node.Path, notify *node.Object) {
	glog.Infof("On observe object %v", notify)
}

func onObserveResource(d *core.Device, p node.Path, notify *node.Resource) {
	glog.Infof("On observe resource %v", notify)
	if rand.Float32() > 0.8 {
		go func() {
			glog.Infof("Reset counter")
			time.Sleep(time.Second * 1)
			rp, _ := node.NewPathFromString("/3347/0/5505")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			_ = d.Execute(ctx, rp)
		}()
	}
	go func() {
		time.Sleep(time.Second * 2)
		glog.Infof("Reading /3347/0/5750")
		rp, _ := node.NewPathFromString("/3347/0/5750")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if res, err := d.ReadResource(ctx, rp); err == nil {
			glog.Infof("Read resource /3347/0/5750 result %v", res)
			if len(res.Data().StringVal()) == 0 {
				glog.Infof("Writing /3347/0/5750")
				val := encoding.NewTlv(encoding.TlvSingleResource, 5750, "Created by server")
				resVal, _ := node.NewSingleResource(rp, val)
				if err = d.WriteResource(ctx, rp, resVal); err != nil {
					glog.Infof("Write resource /3347/0/5750 error %v", err)
				}
			}
		}
	}()
}

func PSKFromIdentity(hint []byte) ([]byte, error) {
	return []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	r := server.DefaultRouter()
	//r.Use(loggingMiddleware)
	manager := core.DefaultManager(
		core.WithLogger(lf.NewLogger("core")),
		core.WithContext(ctx),
	)
	manager.OnDeviceStateChange(onDeviceStateChange)
	registration.EnableHandler(r, manager,
		registration.WithLogger(lf.NewLogger("registration")),
	)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		glog.Infof("Starting lwm2m server")
		err := server.ListenAndServeWithContext(ctx, r,
			server.WithLogger(lf.NewLogger("server")),
			server.EnableUDPListener("udp", ":5683"),
			server.EnableDTLSListener("udp", ":5684", PSKFromIdentity),
		)
		if err != nil {
			glog.Errorf("Serve lwm2m with err: %v", err)
		}
	}()

	select {
	case <-c:
		glog.Infof("Stopping...")
		cancel()
	case <-ctx.Done():
	}
	signal.Stop(c)
	time.Sleep(time.Second)
}

type logger struct {
	*logrus.Entry
}

type loggerFactory struct {
	log *logrus.Logger
}

func (l loggerFactory) NewLogger(scope string) logging.LeveledLogger {
	return &logger{logrus.NewEntry(l.log).WithField("scope", scope)}

}

func NewDefaultLoggerFactory() logging.LoggerFactory {
	logrus.SetLevel(logrus.TraceLevel)
	l := logrus.StandardLogger()
	return &loggerFactory{
		log: l,
	}
}

func (l *logger) Trace(msg string) { l.Entry.Trace(msg) }
func (l *logger) Debug(msg string) { l.Entry.Debug(msg) }
func (l *logger) Info(msg string)  { l.Entry.Info(msg) }
func (l *logger) Warn(msg string)  { l.Entry.Warn(msg) }
func (l *logger) Error(msg string) { l.Entry.Error(msg) }
