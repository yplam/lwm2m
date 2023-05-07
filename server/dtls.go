package server

import udpClient "github.com/plgd-dev/go-coap/v3/udp/client"

type PSKStore interface {
	PSKIdentityFromEP([]byte) ([]byte, error)
	PSKFromIdentity([]byte) ([]byte, error)
}

//func onNewClientConn(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
//cc.SetContextValue(pskIdHint, dtlsConn.ConnectionState().IdentityHint)
//}

func dtlsSetPSK(cc *udpClient.Conn) {}
