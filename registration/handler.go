package registration

import (
	"github.com/pion/logging"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/yplam/lwm2m/core"
	"github.com/yplam/lwm2m/encoding"
	"io"
)

type ValidateClientConnCallback func(cc mux.Conn, ep string) error

type Handler struct {
	logger     logging.LeveledLogger
	manager    core.Manager
	validateCb ValidateClientConnCallback
}

func (h *Handler) ServeCOAP(w mux.ResponseWriter, r *mux.Message) {
	opts := r.Options()
	firstIdx, lastIdx, err := opts.Find(message.URIPath)
	if err != nil || string(opts[firstIdx].Value) != "rd" {
		h.logger.Warnf("wrong request")
		h.handleBadRequest(w)
		return
	}
	if lastIdx-1 == firstIdx {
		// handle registration
		h.logger.Debug("handle registration")
		h.handleRegistration(w, r)
	} else if lastIdx-2 == firstIdx {
		id := string(opts[firstIdx+1].Value)
		// handle update
		if r.Code() == codes.POST {
			h.logger.Debugf("handle update %v", id)
			h.handleUpdate(w, r, id)
		} else if r.Code() == codes.DELETE {
			h.logger.Debugf("handle delete %v", id)
			h.handleDelete(w, r, id)
		} else {
			h.logger.Warnf("unsupported code %v", r.Code())
			h.handleBadRequest(w)
		}
	} else {
		h.logger.Warnf("bad request %v", message.URIPath)
		h.handleBadRequest(w)
	}
}

func (h *Handler) handleBadRequest(w mux.ResponseWriter) {
	if err := w.SetResponse(codes.BadRequest, message.TextPlain, nil); err != nil {
		h.logger.Warnf("handling with error: %v", err)
	}
}

func (h *Handler) handleRegistration(w mux.ResponseWriter, r *mux.Message) {
	q, err := r.Options().Queries()
	if err != nil {
		h.handleBadRequest(w)
		return
	}
	req, err := core.NewRegisterRequest(q)
	if err != nil {
		h.logger.Warnf("parse registration request err %v", err)
		h.handleBadRequest(w)
		return
	}
	// use this callback to validate dtls connection and register endpoint
	if h.validateCb != nil {
		err = h.validateCb(w.Conn(), req.Ep)
		if err != nil {
			_ = w.SetResponse(codes.Forbidden, message.TextPlain, nil)
			return
		}
	}
	var links []*encoding.CoreLink
	if r.Body() != nil {
		if b, err2 := io.ReadAll(r.Body()); err2 == nil {
			links, _ = encoding.CoreLinksFromString(string(b))
		}
	}
	d, err := h.manager.Register(req, links, w.Conn())
	if err != nil {
		h.handleBadRequest(w)
		return
	}
	h.logger.Debugf("registration: %#v", req)
	if err = w.SetResponse(codes.Created, message.TextPlain, nil,
		message.Option{ID: message.LocationPath, Value: []byte("rd")},
		message.Option{ID: message.LocationPath, Value: []byte(d.Id)}); err == nil {
		h.logger.Debugf("registration ok")
		go func() {
			h.manager.PostRegister(d.Id)
		}()
	} else {
		h.logger.Warnf("registration err")
	}
}

func (h *Handler) handleUpdate(w mux.ResponseWriter, r *mux.Message, id string) {
	req := &core.UpdateRequest{
		Lifetime:    nil,
		BindingMode: nil,
		SmsNumber:   nil,
	}
	q, err := r.Options().Queries()
	if err == nil {
		req, err = core.NewUpdateRequest(q)
		if err != nil {
			h.logger.Warnf("parse update request err %v", err)
			h.handleBadRequest(w)
			return
		}
	}
	var links []*encoding.CoreLink
	if r.Body() != nil {
		if b, err2 := io.ReadAll(r.Body()); err2 == nil {
			links, _ = encoding.CoreLinksFromString(string(b))
		}
	}
	err = h.manager.Update(id, req, links, w.Conn())
	if err != nil {
		_ = w.SetResponse(codes.NotFound, message.TextPlain, nil)
		return
	}
	if err = w.SetResponse(codes.Changed, message.TextPlain, nil); err == nil {
		h.logger.Debugf("update ok")
		h.manager.PostUpdate(id)
	}
}

func (h *Handler) handleDelete(w mux.ResponseWriter, r *mux.Message, id string) {
	if err := h.manager.Deregister(id); err == nil {
		_ = w.SetResponse(codes.Deleted, message.TextPlain, nil)
	} else {
		_ = w.SetResponse(codes.NotFound, message.TextPlain, nil)
	}
}

func EnableHandler(r *mux.Router, m core.Manager, opts ...Option) {
	cfg := newConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.logger == nil {
		lf := logging.NewDefaultLoggerFactory()
		cfg.logger = lf.NewLogger("registration")
	}
	h := &Handler{
		logger:  cfg.logger,
		manager: m,
	}
	_ = r.Handle("/rd", h)
	_ = r.Handle("/rd/{v1}", h)
}
