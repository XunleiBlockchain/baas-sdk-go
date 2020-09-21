package sdk

// Logger interface
type Logger interface {
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
}

var sdklog Logger

func setLogger(log Logger) {
	sdklog = log
}
