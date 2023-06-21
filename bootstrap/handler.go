package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pion/logging"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/yplam/lwm2m/core"
	"strconv"
	"strings"
	"time"
)

const DefaultBootstrapTimeout = 30 * time.Second

type Handler struct {
	timeout  time.Duration
	logger   logging.LeveledLogger
	provider Provider
}

func (h *Handler) ServeCOAP(w mux.ResponseWriter, r *mux.Message) {
	h.logger.Debugf("bootstrap serve coap: %v", r)

	ep := ""                                                    // endpoint, optional
	pctOpt := strconv.FormatInt(int64(message.AppLwm2mTLV), 10) // preferred content-type, optional
	opts := r.Options()
	for _, o := range opts {
		h.logger.Debugf("option: %s, %d", string(o.Value), o.ID)
		if o.ID != message.URIQuery {
			continue
		}
		so := strings.Split(string(o.Value), "=")
		if len(so) != 2 {
			w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("bad uri query")))
			return
		}
		switch so[0] {
		case "ep":
			ep = so[1]
		case "pct":
			pctOpt = so[1]
		default:
			w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("unsupported option")))
			return
		}
	}

	ipct, err := strconv.ParseUint(pctOpt, 10, 16)
	if err != nil {
		err = fmt.Errorf("parse pct option: %w", err)
		w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte(err.Error())))
		return
	}
	pct := message.MediaType(ipct)

	// select content type
	switch pct {
	case message.AppLwm2mTLV:
	default:
		err = fmt.Errorf("unsupported content type: %s", pct.String())
		w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte(err.Error())))
		return
	}

	client := &Client{
		Endpoint:   ep,
		conn:       w.Conn(),
		selectedCt: pct,
	}

	core.SetLifetime(client.conn, h.timeout)

	err = w.SetResponse(codes.Changed, client.selectedCt, nil)
	if err != nil {
		h.logger.Errorf("error sending response to bootstrap request: %v", err)
		return
	}

	// limit bootstrap procedure by timeout with ctx
	ctx, cancel := context.WithTimeout(context.TODO(), h.timeout)
	go func() {
		defer func() {
			// send finish message in any case
			err := client.finish(context.TODO())
			if err != nil {
				h.logger.Errorf("error sending bootstrap-finish: %v", err)
			}
			cancel()
		}()
		err := h.provider.HandleBsRequest(ctx, client)
		if err != nil {
			h.logger.Errorf("bootstrap provider err: %v", err)
			return
		}
	}()
}

func EnableHandler(r *mux.Router, p Provider, opts ...Option) {
	h := &Handler{
		provider: p,
		timeout:  DefaultBootstrapTimeout,
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.logger == nil {
		lf := logging.NewDefaultLoggerFactory()
		h.logger = lf.NewLogger("bootstrap")
	}
	_ = r.Handle("/bs", h)
}
