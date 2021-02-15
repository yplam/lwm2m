package dtls

import (
	"bytes"
	"errors"
	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

const (
	PSK_ID_HINT = "psk_id_hint"
)


type Store interface {
	PSKFromIdentity([]byte) ([]byte, error)
	PSKIdentityFromEP([]byte) ([]byte, error)
}

type Server struct {
	store Store
}

func (s *Server) PSK(id []byte) ([]byte, error) {
	return s.store.PSKFromIdentity(id)
}

func (s *Server) OnNewClientConn(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
	cc.SetContextValue(PSK_ID_HINT, dtlsConn.ConnectionState().IdentityHint)
}

func (s *Server) ValidateClientConn(cc *client.ClientConn, ep string) error {
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

func NewServer(s Store) *Server {
	return &Server{
		store: s,
	}
}