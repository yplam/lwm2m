package lwm2m

//func ListenAndServe(ctx context.Context, s *Server, network string, addr string) {
//	r := mux.NewRouter()
//	reg := NewRegistration(s, s.LoggerFactory.NewLogger("registration"))
//	_ = r.Handle("/rd", reg)
//	_ = r.Handle("/rd/", reg)
//
//	switch network {
//	case "udp", "udp4", "udp6", "":
//		l, err := net.NewListenUDP(network, addr)
//		if err != nil {
//			logrus.Fatal(fmt.Errorf("listen udp error (%v)", err))
//		}
//		defer l.Close()
//		us := udp.NewServer(udp.WithMux(r))
//		go func() {
//			logrus.Fatal(us.Serve(l))
//		}()
//		<-ctx.Done()
//		us.Stop()
//	case "tcp", "tcp4", "tcp6":
//		l, err := net.NewTCPListener(network, addr)
//		if err != nil {
//			logrus.Fatal(fmt.Errorf("listen tcp error (%v)", err))
//		}
//		defer l.Close()
//		ts := tcp.NewServer(tcp.WithMux(r))
//		go func() {
//			logrus.Fatal(ts.Serve(l))
//		}()
//		<-ctx.Done()
//		ts.Stop()
//	default:
//		logrus.Fatal(fmt.Errorf("invalid network (%v)", network))
//	}
//}
//
//func ListenAndServeDTLS(ctx context.Context, s *Server, network string, addr string) {
//	dc := NewDTLSConnector(s.store)
//	dtlsConfig := piondtls.Config{
//		CipherSuites:         []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
//		ExtendedMasterSecret: piondtls.DisableExtendedMasterSecret,
//		PSK: func(id []byte) ([]byte, error) {
//			return dc.psk(id)
//		},
//		LoggerFactory: s.LoggerFactory,
//		ConnectContextMaker: func() (context.Context, func()) {
//			return context.WithCancel(ctx)
//		},
//	}
//
//	r := mux.NewRouter()
//	reg := NewRegistration(s, s.LoggerFactory.NewLogger("registration"))
//	_ = r.Handle("/rd", reg)
//	_ = r.Handle("/rd/", reg)
//
//	l, err := net.NewDTLSListener(network, addr, &dtlsConfig)
//	if err != nil {
//		panic(err.Error())
//	}
//	defer l.Close()
//
//	cs := coapdtls.NewServer(coapdtls.WithMux(r),
//		//coapdtls.WithKeepAlive(0,0, nil),
//		coapdtls.WithOnNewClientConn(func(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
//			dc.onNewClientConn(cc, dtlsConn)
//		}))
//
//	go func() {
//		logrus.Fatal(cs.Serve(l))
//	}()
//	<-ctx.Done()
//	cs.Stop()
//}
