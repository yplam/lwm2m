package registration

import (
	"github.com/pion/logging"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

type Handler struct {
	logger logging.LeveledLogger
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
			h.logger.Debug("handle delete")
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

}

func (h *Handler) handleUpdate(w mux.ResponseWriter, r *mux.Message, id string) {

}

func (h *Handler) handleDelete(w mux.ResponseWriter, r *mux.Message, id string) {

}

func NewHandler(r *mux.Router, opts ...Option) *Handler {
	cfg := newConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.logger == nil {
		lf := logging.NewDefaultLoggerFactory()
		cfg.logger = lf.NewLogger("registration")
	}
	h := &Handler{
		logger: cfg.logger,
	}
	_ = r.Handle("/rd", h)
	_ = r.Handle("/rd/", h)
	return h
}
