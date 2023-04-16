package server

type config struct {
	udpNetwork string
	udpAddr    string
	tcpNetwork string
	tcpAddr    string
}

func newServeConfig() *config {
	return &config{
		udpNetwork: "",
		udpAddr:    "",
		tcpNetwork: "",
		tcpAddr:    "",
	}
}

type Option func(cfg *config)

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
