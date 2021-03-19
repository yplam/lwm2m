package server

// OnNewDeviceConnFunc network option.
type OnNewDeviceConnOpt struct {
	onNewDeviceConn OnNewDeviceConnFunc
}

func (o OnNewDeviceConnOpt) apply(opts *serverOptions) {
	opts.onNewDeviceConn = o.onNewDeviceConn
}
func WithOnNewDeviceConn(onNewDeviceConn OnNewDeviceConnFunc) OnNewDeviceConnOpt {
	return OnNewDeviceConnOpt{
		onNewDeviceConn: onNewDeviceConn,
	}
}
