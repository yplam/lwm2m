package lwm2m

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pion/logging"

	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	"github.com/plgd-dev/go-coap/v2/udp/client"
)

type ValidateClientConnCallback func(cc *client.ClientConn, ep string) error

type Registration struct {
	s                  *Server
	log                logging.LeveledLogger
	ValidateClientConn ValidateClientConnCallback
}

func (r *Registration) ServeCOAP(w mux.ResponseWriter, m *mux.Message) {
	//log.Printf("registration resource from %v\n",w.Client().RemoteAddr())
	firstIdx, lastIdx, err := m.Options.Find(message.URIPath)
	if err != nil || string(m.Options[firstIdx].Value) != "rd" {
		r.log.Warnf("wrong path %v", m.Options[firstIdx].Value)
		r.handleBadRequest(w)
		return
	}
	if lastIdx-1 == firstIdx {
		// handle registration
		r.log.Debug("handle registration")
		r.handleRegistration(w, m)
	} else if lastIdx-2 == firstIdx {
		id := string(m.Options[firstIdx+1].Value)
		// handle update
		if m.Code == codes.POST {
			r.log.Debugf("handle update %v", id)
			r.handleUpdate(w, m, id)
		} else if m.Code == codes.DELETE {
			r.log.Debug("handle delete")
			r.handleDelete(w, m, id)
		} else {
			r.log.Warnf("unsupported code %v", m.Code)
			r.handleBadRequest(w)
		}
	} else {
		r.log.Warnf("bad request %v", message.URIPath)
		r.handleBadRequest(w)
	}
}

func (r *Registration) handleBadRequest(w mux.ResponseWriter) {
	if err := w.SetResponse(codes.BadRequest, message.TextPlain, nil); err != nil {
		r.log.Warnf("handling with error: %v", err)
	}
}

func (r *Registration) handleRegistration(w mux.ResponseWriter, m *mux.Message) {
	q, err := m.Options.Queries()
	if err != nil {
		r.handleBadRequest(w)
		return
	}
	var endpoint string
	var lifetime int
	var version string
	var smsNumber string
	var binding string
	params := make(map[string]string)

	for _, val := range q {
		r.log.Debugf("reg q: %v", val)
		sps := strings.Split(val, "=")
		if len(sps) != 2 {
			continue
		}
		switch sps[0] {
		case "ep":
			endpoint = sps[1]
		case "lwm2m":
			version = sps[1]
		case "lt":
			lifetime, err = strconv.Atoi(sps[1])
		case "sms":
			smsNumber = sps[1]
		case "b":
			binding = sps[1]
		default:
			params[sps[0]] = sps[1]
		}
	}
	if err != nil {
		r.handleBadRequest(w)
		return
	}
	// use this callback to validate dtls connection and register endpoint
	if r.ValidateClientConn != nil {
		err = r.ValidateClientConn(w.Client().ClientConn().(*client.ClientConn), endpoint)
		if err != nil {
			_ = w.SetResponse(codes.Forbidden, message.TextPlain, nil)
			return
		}
	}
	var links []*CoreLink
	if m.Body != nil {
		if b, err := ioutil.ReadAll(m.Body); err == nil {
			links, _ = CoreLinksFromString(string(b))
		}
	}
	d, err := r.s.Register(endpoint, lifetime, version, binding, smsNumber, links, w.Client())
	if err != nil {
		r.handleBadRequest(w)
		return
	}
	r.log.Debugf("%v, %v, %v", endpoint, lifetime, version)
	if err = w.SetResponse(codes.Created, message.TextPlain, nil,
		message.Option{ID: message.LocationPath, Value: []byte("rd")},
		message.Option{ID: message.LocationPath, Value: []byte(d.ID)}); err == nil {
		r.s.PostRegister(d.ID)
	}
}

func (r *Registration) handleUpdate(w mux.ResponseWriter, m *mux.Message, id string) {
	q, err := m.Options.Queries()
	var lifetime int
	var smsNumber string
	var binding string
	params := make(map[string]string)
	var links []*CoreLink
	if err == nil {
		for _, val := range q {
			r.log.Debugf("update q: %v", val)
			sps := strings.Split(val, "=")
			if len(sps) != 2 {
				continue
			}
			switch sps[0] {
			case "lt":
				lifetime, err = strconv.Atoi(sps[1])
				if err != nil {
					lifetime = 0
				}
			case "sms":
				smsNumber = sps[1]
			case "b":
				binding = sps[1]
			default:
				params[sps[0]] = sps[1]
			}
		}
	}

	if m.Body != nil {
		if b, err := ioutil.ReadAll(m.Body); err == nil {
			links, _ = CoreLinksFromString(string(b))
		}
	}
	err = r.s.Update(id, lifetime, binding, smsNumber, links, w.Client())
	if err != nil {
		_ = w.SetResponse(codes.NotFound, message.TextPlain, nil)
		return
	}
	if err = w.SetResponse(codes.Changed, message.TextPlain, nil); err == nil {
		r.s.PostUpdate(id)
	}
}

func (r *Registration) handleDelete(w mux.ResponseWriter, m *mux.Message, id string) {
	d := r.s.GetByID(id)
	if d == nil {
		_ = w.SetResponse(codes.NotFound, message.TextPlain, nil)
	} else {
		_ = w.SetResponse(codes.Deleted, message.TextPlain, nil)
	}
	_ = r.s.DeRegister(id)
}

func NewRegistration(s *Server, l logging.LeveledLogger) *Registration {
	return &Registration{
		s:   s,
		log: l,
	}
}
