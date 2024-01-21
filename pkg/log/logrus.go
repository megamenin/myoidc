package log

import (
	"myoidc/pkg/errors"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logger logrus.FieldLogger
}

func NewLogrusLogger(opts ...LogrusOption) Logger {
	o := NewLogrusOptions()
	o.Apply(opts)

	logger := logrus.New()
	logger.SetLevel(o.Level)
	logger.SetFormatter(o.Formatter)

	return &logrusLogger{
		logger: logger,
	}
}

func (l *logrusLogger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *logrusLogger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *logrusLogger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *logrusLogger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *logrusLogger) Fatal(msg string) {
	l.logger.Fatal(msg)
}

func (l *logrusLogger) Debugf(msg string, v ...interface{}) {
	l.logger.Debugf(msg, v...)
}

func (l *logrusLogger) Infof(msg string, v ...interface{}) {
	l.logger.Infof(msg, v...)
}

func (l *logrusLogger) Warnf(msg string, v ...interface{}) {
	l.logger.Warnf(msg, v...)
}

func (l *logrusLogger) Errorf(msg string, v ...interface{}) {
	l.logger.Errorf(msg, v...)
}

func (l *logrusLogger) Fatalf(msg string, v ...interface{}) {
	l.logger.Fatalf(msg, v...)
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{logger: l.logger.WithField(key, value)}
}

func (l *logrusLogger) WithFields(fields Fields) Logger {
	return &logrusLogger{logger: l.logger.WithFields(logrus.Fields(fields))}
}

func (l *logrusLogger) WithError(err error) Logger {
	fields := logrus.Fields{
		"message":    err.Error(),
		"stacktrace": errors.GetStackTrace(err),
	}
	f := errors.GetFields(err)
	for key, value := range f {
		fields[key] = value
	}
	return &logrusLogger{logger: l.logger.WithError(err).WithFields(fields)}
}

type LogrusOptions struct {
	Level     logrus.Level
	Formatter logrus.Formatter
}

func NewLogrusOptions() *LogrusOptions {
	return &LogrusOptions{
		Level:     logrus.InfoLevel,
		Formatter: &logrus.TextFormatter{},
	}
}

func (o *LogrusOptions) Apply(opts []LogrusOption) {
	for _, opt := range opts {
		opt(o)
	}
}

type LogrusOption func(o *LogrusOptions)

func LogrusLogLevel(lvl logrus.Level) LogrusOption {
	return func(o *LogrusOptions) {
		o.Level = lvl
	}
}

func LogrusFormatter(fmt logrus.Formatter) LogrusOption {
	return func(o *LogrusOptions) {
		o.Formatter = fmt
	}
}
