package bootstrap

import (
	"context"
	"github.com/pion/logging"
)

type Option func(h *Handler)

func WithLogger(l logging.LeveledLogger) Option {
	return func(h *Handler) {
		h.logger = l
	}
}

func WithParentContext(c context.Context) Option {
	return func(h *Handler) {
		h.ctx = c
	}
}
