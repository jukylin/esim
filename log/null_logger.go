package log

import (
	"context"
)

type nullLogger struct{}

func NewNullLogger() Logger {
	logger := &nullLogger{}

	return logger
}

func (log *nullLogger) Error(msg string) {}

func (log *nullLogger) Debugf(template string, args ...interface{}) {}

func (log *nullLogger) Infof(template string, args ...interface{}) {}

func (log *nullLogger) Warnf(template string, args ...interface{}) {}

func (log *nullLogger) Errorf(template string, args ...interface{}) {}

func (log *nullLogger) DPanicf(template string, args ...interface{}) {}

func (log *nullLogger) Panicf(template string, args ...interface{}) {}

func (log *nullLogger) Fatalf(template string, args ...interface{}) {}

func (log *nullLogger) Debugc(ctx context.Context, template string, args ...interface{}) {}

func (log *nullLogger) Infoc(ctx context.Context, template string, args ...interface{}) {}

func (log *nullLogger) Warnc(ctx context.Context, template string, args ...interface{}) {}

func (log *nullLogger) Errorc(ctx context.Context, template string, args ...interface{}) {}

func (log *nullLogger) DPanicc(ctx context.Context, template string, args ...interface{}) {}

func (log *nullLogger) Panicc(ctx context.Context, template string, args ...interface{}) {}

func (log *nullLogger) Fatalc(ctx context.Context, template string, args ...interface{}) {}
