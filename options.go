package lwm2m

import "github.com/pion/logging"

type ServerOption func(*Server)

// WithOnNewDeviceConn network option.
func WithOnNewDeviceConn(onNewDeviceConn OnNewDeviceConnFunc) ServerOption {
	return func(o *Server) {
		o.onNewDeviceConn = onNewDeviceConn
	}
}

func WithLoggerFactory(l logging.LoggerFactory) ServerOption {
	return func(o *Server) {
		o.LoggerFactory = l
	}
}

func EnableUDPListener(network, addr string) ServerOption {
	return func(o *Server) {
		o.udpNetwork = network
		o.udpAddr = addr
	}
}

func EnableDTLSListener(network, addr string, s Store) ServerOption {
	return func(o *Server) {
		o.dtlsNetwork = network
		o.dtlsAddr = addr
		o.store = s
	}
}
