package server

import "github.com/pion/logging"

type config struct {
	udpNetwork  string
	udpAddr     string
	tcpNetwork  string
	tcpAddr     string
	dtlsNetwork string
	dtlsAddr    string
	pskStore    PSKStore
	logger      logging.LeveledLogger
}

func newServeConfig() *config {
	return &config{
		udpNetwork:  "",
		udpAddr:     "",
		tcpNetwork:  "",
		tcpAddr:     "",
		dtlsNetwork: "",
		dtlsAddr:    "",
		pskStore:    nil,
		logger:      nil,
	}
}

type Option func(cfg *config)

func WithLogger(l logging.LeveledLogger) Option {
	return func(o *config) {
		o.logger = l
	}
}

func EnableUDPListener(network, addr string) Option {
	return func(o *config) {
		o.udpNetwork = network
		o.udpAddr = addr
	}
}

func EnableTCPListener(network, addr string) Option {
	return func(o *config) {
		o.tcpNetwork = network
		o.tcpAddr = addr
	}
}

func EnableDTLSListener(network, addr string, store PSKStore) Option {
	return func(o *config) {
		o.dtlsNetwork = network
		o.dtlsAddr = addr
		o.pskStore = store
	}
}
