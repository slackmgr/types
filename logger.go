package common

import "io"

type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...any)
	Info(msg string)
	Infof(format string, args ...any)
	Error(msg string)
	Errorf(format string, args ...any)
	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger
	HttpLoggingHandler() io.Writer
}
