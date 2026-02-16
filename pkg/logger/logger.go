package logger

type Logger interface {
	Debug(msg string, keyvals ...any)
}

var nop Logger = &nopLogger{}

func Nop() Logger { return nop }

type nopLogger struct{}

func (*nopLogger) Debug(string, ...any) {}

type Func func(msg string, keyvals ...any)

func (f Func) Debug(msg string, keyvals ...any) { f(msg, keyvals...) }
