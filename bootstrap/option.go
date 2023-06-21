package bootstrap

import (
	"github.com/pion/logging"
	"time"
)

type Option func(h *Handler)

func WithLogger(l logging.LeveledLogger) Option {
	return func(h *Handler) {
		h.logger = l
	}
}

func WithBootstrapTimeout(timeout time.Duration) Option {
	return func(h *Handler) {
		h.timeout = timeout
	}
}
