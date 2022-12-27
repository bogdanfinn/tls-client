package tls_client

import "fmt"

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type noopLogger struct {
}

func NewNoopLogger() Logger {
	return &noopLogger{}
}

func (n noopLogger) Debug(format string, args ...interface{}) {}

func (n noopLogger) Info(format string, args ...interface{}) {}

func (n noopLogger) Warn(format string, args ...interface{}) {}

func (n noopLogger) Error(format string, args ...interface{}) {}

type debugLogger struct {
	logger Logger
}

func NewDebugLogger(logger Logger) Logger {
	return &debugLogger{
		logger: logger,
	}
}

func (n debugLogger) Debug(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (n debugLogger) Info(format string, args ...interface{}) {
	n.logger.Info(format, args...)
}

func (n debugLogger) Warn(format string, args ...interface{}) {
	n.logger.Warn(format, args...)
}

func (n debugLogger) Error(format string, args ...interface{}) {
	n.logger.Error(format, args...)
}

type logger struct{}

func NewLogger() Logger {
	return &logger{}
}

func (n logger) Debug(format string, args ...interface{}) {}

func (n logger) Info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (n logger) Warn(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (n logger) Error(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
