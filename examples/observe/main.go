package main

import (
	"context"
	"github.com/pion/logging"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/sirupsen/logrus"
	"github.com/yplam/lwm2m/core"
	"github.com/yplam/lwm2m/node"
	"github.com/yplam/lwm2m/registration"
	"github.com/yplam/lwm2m/server"
	"io"
	"os"
	"os/signal"
	"time"
	"unicode"
)

var lf = NewDefaultLoggerFactory()
var glog = lf.NewLogger("main")

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

func onDeviceStateChange(e core.DeviceEvent, m core.Manager) {
	//glog.Debugf("Device %v Event %v", e.Device.Id, e)
	device := e.Device
	switch e.EventType {
	case core.DevicePostRegister:
		glog.Infof("post register")
		if device.HasObjectWithInstance(3347) {

			p, _ := node.NewPathFromString("/3347/0")
			if err := device.Observe(p, onObserve); err != nil {
				glog.Warnf("Observe error %v", err)
			}

			op, _ := node.NewPathFromString("/3347")
			if err := device.ObserveObject(op, onObserveObject); err != nil {
				glog.Warnf("Observe object error %v", err)
			}

			pr, _ := node.NewPathFromString("/3347/0/5501")
			if err := device.ObserveResource(pr, onObserveResource); err != nil {
				glog.Warnf("Observe resource error %v", err)
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
			server.EnableUDPListener("udp", ":5683"))
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
	logrus.SetLevel(logrus.DebugLevel)
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
