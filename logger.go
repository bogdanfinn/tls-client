package tls_client

import "fmt"

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

type noopLogger struct {
}

func NewNoopLogger() Logger {
	return &noopLogger{}
}

func (n noopLogger) Debug(_ string, _ ...any) {}

func (n noopLogger) Info(_ string, _ ...any) {}

func (n noopLogger) Warn(_ string, _ ...any) {}

func (n noopLogger) Error(_ string, _ ...any) {}

type debugLogger struct {
	logger Logger
}

func NewDebugLogger(logger Logger) Logger {
	return &debugLogger{
		logger: logger,
	}
}

func (n debugLogger) Debug(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (n debugLogger) Info(format string, args ...any) {
	n.logger.Info(format, args...)
}

func (n debugLogger) Warn(format string, args ...any) {
	n.logger.Warn(format, args...)
}

func (n debugLogger) Error(format string, args ...any) {
	n.logger.Error(format, args...)
}

type logger struct{}

func NewLogger() Logger {
	return &logger{}
}

func (n logger) Debug(_ string, _ ...any) {}

func (n logger) Info(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (n logger) Warn(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (n logger) Error(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// Interface guards are a cheap way to make sure all methods are implemented, this is a static check and does not affect runtime performance.
var _ Logger = (*logger)(nil)
var _ Logger = (*debugLogger)(nil)
var _ Logger = (*noopLogger)(nil)
