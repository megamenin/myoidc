package log

type Fields map[string]interface{}

type Logger interface {
	Debug(msg string)
	Debugf(msg string, args ...interface{})
	Info(msg string)
	Infof(msg string, args ...interface{})
	Warn(msg string)
	Warnf(msg string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
}
