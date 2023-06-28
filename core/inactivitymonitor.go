package core

import (
	"context"
	dtlsServer "github.com/plgd-dev/go-coap/v3/dtls/server"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/net/monitor/inactivity"
	"github.com/plgd-dev/go-coap/v3/options"
	tcpClient "github.com/plgd-dev/go-coap/v3/tcp/client"
	tcpServer "github.com/plgd-dev/go-coap/v3/tcp/server"
	udpClient "github.com/plgd-dev/go-coap/v3/udp/client"
	udpServer "github.com/plgd-dev/go-coap/v3/udp/server"
	"sync/atomic"
	"time"
)

// InactivityMonitorOpt notifies when a connection was inactive for a given duration.
type InactivityMonitorOpt[C options.OnInactiveFunc] struct {
	onInactive C
}

func (o InactivityMonitorOpt[C]) toTCPCreateInactivityMonitor(onInactive options.TCPOnInactive) func() tcpClient.InactivityMonitor {
	return func() tcpClient.InactivityMonitor {
		return NewInactivityMonitor(onInactive)
	}
}

func (o InactivityMonitorOpt[C]) toUDPCreateInactivityMonitor(onInactive options.UDPOnInactive) func() udpClient.InactivityMonitor {
	return func() udpClient.InactivityMonitor {
		return NewInactivityMonitor(onInactive)
	}
}

func (o InactivityMonitorOpt[C]) TCPServerApply(cfg *tcpServer.Config) {
	switch onInactive := any(o.onInactive).(type) {
	case options.TCPOnInactive:
		cfg.CreateInactivityMonitor = o.toTCPCreateInactivityMonitor(onInactive)
	}
}

func (o InactivityMonitorOpt[C]) UDPServerApply(cfg *udpServer.Config) {
	switch onInactive := any(o.onInactive).(type) {
	case options.UDPOnInactive:
		cfg.CreateInactivityMonitor = o.toUDPCreateInactivityMonitor(onInactive)
	default:
	}
}

func (o InactivityMonitorOpt[C]) DTLSServerApply(cfg *dtlsServer.Config) {
	switch onInactive := any(o.onInactive).(type) {
	case options.UDPOnInactive:
		cfg.CreateInactivityMonitor = o.toUDPCreateInactivityMonitor(onInactive)
	}
}

// WithInactivityMonitor set deadline's for read operations over client connection.
func WithInactivityMonitor[C options.OnInactiveFunc](onInactive C) InactivityMonitorOpt[C] {
	return InactivityMonitorOpt[C]{
		onInactive: onInactive,
	}
}

func NewInactivityMonitor[C inactivity.Conn](onInactive inactivity.OnInactiveFunc[C]) *InactivityMonitor[C] {
	m := &InactivityMonitor[C]{
		onInactive: onInactive,
	}
	m.Notify()
	return m
}

type InactivityMonitor[C inactivity.Conn] struct {
	lastActivity atomic.Value
	onInactive   inactivity.OnInactiveFunc[C]
}

func (m *InactivityMonitor[C]) Notify() {
	m.lastActivity.Store(time.Now())
}

func (m *InactivityMonitor[C]) LastActivity() time.Time {
	if t, ok := m.lastActivity.Load().(time.Time); ok {
		return t
	}
	return time.Time{}
}

func CloseConn(cc inactivity.Conn) {
	// call cc.Close() directly to check and handle error if necessary
	_ = cc.Close()
}

func New[C inactivity.Conn](onInactive inactivity.OnInactiveFunc[C]) *InactivityMonitor[C] {
	m := &InactivityMonitor[C]{
		onInactive: onInactive,
	}
	m.Notify()
	return m
}

type lifetimeCtxKeyType string

const lifetimeCtxKey lifetimeCtxKeyType = "core_lifetime"

func WithLifetime(ctx context.Context, lifetime time.Duration) context.Context {
	return context.WithValue(ctx, lifetimeCtxKey, lifetime)
}

func SetLifetime(c mux.Conn, d time.Duration) {
	c.SetContextValue(lifetimeCtxKey, d)
}

func GetLifetime(ctx context.Context) *time.Duration {
	lifetime, ok := ctx.Value(lifetimeCtxKey).(time.Duration)
	if !ok {
		// Log this issue
		return nil
	}
	return &lifetime
}

func (m *InactivityMonitor[C]) CheckInactivity(now time.Time, cc C) {
	if m.onInactive == nil {
		return
	}
	duration := time.Second * 2
	if lifetime := GetLifetime(cc.Context()); lifetime != nil {
		duration = *lifetime
	}
	if now.After(m.LastActivity().Add(duration)) {
		m.onInactive(cc)
	}
}
