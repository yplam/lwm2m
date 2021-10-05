package lwm2m

import (
	"bytes"
	"errors"

	piondtls "github.com/pion/dtls/v2"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

const (
	pskIdHint = "psk_id_hint"
)

type DTLSConnector struct {
	store Store
}

func (c *DTLSConnector) psk(id []byte) ([]byte, error) {
	return c.store.PSKFromIdentity(id)
}

func (c *DTLSConnector) onNewClientConn(cc *client.ClientConn, dtlsConn *piondtls.Conn) {
	cc.SetContextValue(pskIdHint, dtlsConn.ConnectionState().IdentityHint)
}

func (c *DTLSConnector) validateClientConn(cc *client.ClientConn, ep string) error {
	hi := cc.Context().Value(pskIdHint).([]byte)
	ehi, err := c.store.PSKIdentityFromEP([]byte(ep))
	if err != nil {
		return err
	}
	if bytes.Compare(hi, ehi) != 0 {
		return errors.New("endpoint not validate")
	}
	return nil
}

func NewDTLSConnector(s Store) *DTLSConnector {
	return &DTLSConnector{
		store: s,
	}
}
