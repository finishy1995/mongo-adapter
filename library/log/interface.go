package log

type Level uint8

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
)

const DefaultLevel = INFO

type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	SetLevel(Level)
}

var (
	logInstance Logger
)

func init() {
	SetLogger(newGoLoggingLogger())
	SetLevel(DefaultLevel)
}

func SetLogger(logger Logger) {
	logInstance = logger
}

func GetLogger() Logger {
	return logInstance
}

func Debugf(message string, param ...interface{}) {
	logInstance.Debugf(message, param...)
}

func Infof(message string, param ...interface{}) {
	logInstance.Infof(message, param...)
}

func Warnf(message string, param ...interface{}) {
	logInstance.Warnf(message, param...)
}

func Errorf(message string, param ...interface{}) {
	logInstance.Errorf(message, param...)
}

func SetLevel(level Level) {
	logInstance.SetLevel(level)
}
