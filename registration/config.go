package registration

import "github.com/pion/logging"

type config struct {
	logger logging.LeveledLogger
}

func newConfig() *config {
	return &config{
		logger: nil,
	}
}

type Option func(cfg *config)

func WithLogger(l logging.LeveledLogger) Option {
	return func(o *config) {
		o.logger = l
	}
}
