package logger

import (
	"log"
	"os"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	WithFields(fields map[string]interface{}) Logger
}

type DefaultLogger struct {
	level  Level
	logger *log.Logger
	fields map[string]interface{}
}

func NewDefaultLogger(level Level) *DefaultLogger {
	return &DefaultLogger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
		fields: make(map[string]interface{}),
	}
}

func (l *DefaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.level <= LevelDebug {
		l.log("DEBUG", msg, keysAndValues...)
	}
}

func (l *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	if l.level <= LevelInfo {
		l.log("INFO", msg, keysAndValues...)
	}
}

func (l *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	if l.level <= LevelWarn {
		l.log("WARN", msg, keysAndValues...)
	}
}

func (l *DefaultLogger) Error(msg string, keysAndValues ...interface{}) {
	if l.level <= LevelError {
		l.log("ERROR", msg, keysAndValues...)
	}
}

func (l *DefaultLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &DefaultLogger{
		level:  l.level,
		logger: l.logger,
		fields: newFields,
	}
}

func (l *DefaultLogger) log(level, msg string, keysAndValues ...interface{}) {
	args := make([]interface{}, 0, len(keysAndValues)+len(l.fields)*2+2)
	args = append(args, level, msg)

	for k, v := range l.fields {
		args = append(args, k, v)
	}

	for i := 0; i < len(keysAndValues)-1; i += 2 {
		args = append(args, keysAndValues[i], keysAndValues[i+1])
	}

	l.logger.Println(args...)
}

type NopLogger struct{}

func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

func (l *NopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *NopLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *NopLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *NopLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *NopLogger) WithFields(fields map[string]interface{}) Logger {
	return l
}
