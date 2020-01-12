package log

import "context"

type Logger interface {

	Error(msg string)

	Debugf(string,  ...interface{})

	Infof(string, ...interface{})

	Warnf(string, ...interface{})

	Errorf(string, ...interface{})

	DPanicf(string, ...interface{})

	Panicf(string, ...interface{})

	Fatalf(string, ...interface{})

	Debugc(context.Context, string,  ...interface{})

	Infoc(context.Context, string, ...interface{})

	Warnc(context.Context, string, ...interface{})

	Errorc(context.Context, string, ...interface{})

	DPanicc(context.Context, string, ...interface{})

	Panicc(context.Context, string, ...interface{})

	Fatalc(context.Context, string, ...interface{})
}
