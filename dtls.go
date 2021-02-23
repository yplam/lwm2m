package lwm2m

import (
	"bytes"
	"errors"
	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

const (
	PSK_ID_HINT = "psk_id_hint"
)


type DTLSServer struct {
	store Store
}

func (s *DTLSServer) PSK(id []byte) ([]byte, error) {
	return s.store.PSKFromIdentity(id)
}

func (s *DTLSServer) OnNewClientConn(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
	cc.SetContextValue(PSK_ID_HINT, dtlsConn.ConnectionState().IdentityHint)
}

func (s *DTLSServer) ValidateClientConn(cc *client.ClientConn, ep string) error {
	hi := cc.Context().Value(PSK_ID_HINT).([]byte)
	ehi, err := s.store.PSKIdentityFromEP([]byte(ep))
	if err != nil {
		return err
	}
	if bytes.Compare(hi, ehi) != 0 {
		return errors.New("endpoint not validate")
	}
	return nil
}

func NewDTLSServer(s Store) *DTLSServer {
	return &DTLSServer{
		store: s,
	}
}