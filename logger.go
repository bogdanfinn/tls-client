package tls_client

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

func (n noopLogger) Debug(format string, args ...interface{}) {
	return
}

func (n noopLogger) Info(format string, args ...interface{}) {
	return
}

func (n noopLogger) Warn(format string, args ...interface{}) {
	return
}

func (n noopLogger) Error(format string, args ...interface{}) {
	return
}
