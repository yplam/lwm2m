package lwm2m

import (
	"github.com/pion/logging"
	"github.com/sirupsen/logrus"
)

type logger struct {
	*logrus.Entry
}

type loggerFactory struct {
	log *logrus.Logger
}

func (l loggerFactory) NewLogger(scope string) logging.LeveledLogger {
	return &logger{logrus.NewEntry(l.log).WithField("scope", scope)}

}

func NewDefaultLoggerFactory() logging.LoggerFactory {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	return &loggerFactory{
		log: l,
	}
}

func (l *logger) Trace(msg string) { l.Entry.Trace(msg) }
func (l *logger) Debug(msg string) { l.Entry.Debug(msg) }
func (l *logger) Info(msg string)  { l.Entry.Info(msg) }
func (l *logger) Warn(msg string)  { l.Entry.Warn(msg) }
func (l *logger) Error(msg string) { l.Entry.Error(msg) }
