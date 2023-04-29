package server

import "github.com/pion/logging"

type config struct {
	udpNetwork string
	udpAddr    string
	tcpNetwork string
	tcpAddr    string
	logger     logging.LeveledLogger
}

func newServeConfig() *config {
	return &config{
		udpNetwork: "",
		udpAddr:    "",
		tcpNetwork: "",
		tcpAddr:    "",
		logger:     nil,
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
