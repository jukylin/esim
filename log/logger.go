package log

import (
	"context"

	"go.uber.org/zap"
)

type logger struct {
	debug bool

	json bool

	ez *EsimZap

	logger *zap.Logger

	sugar *zap.SugaredLogger
}

type LoggerOptions struct{}

type Option func(c *logger)

func NewLogger(options ...Option) Logger {
	logger := &logger{}

	for _, option := range options {
		option(logger)
	}

	if logger.ez == nil {
		logger.ez = NewEsimZap(
			WithEsimZapDebug(logger.debug),
			WithEsimZapJSON(logger.json),
		)
	}

	logger.logger = logger.ez.Logger
	logger.sugar = logger.ez.Logger.Sugar()

	return logger
}

func WithDebug(debug bool) Option {
	return func(l *logger) {
		l.debug = debug
	}
}

func WithJSON(json bool) Option {
	return func(l *logger) {
		l.json = json
	}
}

func WithEsimZap(ez *EsimZap) Option {
	return func(l *logger) {
		l.ez = ez
	}
}

func (log *logger) Error(msg string) {
	log.logger.Error(msg)
}

func (log *logger) Debugf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO())...).Debugf(template, args...)
}

func (log *logger) Infof(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO())...).Infof(template, args...)
}

func (log *logger) Warnf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO())...).Warnf(template, args...)
}

func (log *logger) Errorf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO())...).Errorf(template, args...)
}

func (log *logger) DPanicf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO())...).DPanicf(template, args...)
}

func (log *logger) Panicf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO())...).Panicf(template, args...)
}

func (log *logger) Fatalf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO())...).Fatalf(template, args...)
}

func (log *logger) Debugc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx)...).Debugf(template, args...)
}

func (log *logger) Infoc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx)...).Infof(template, args...)
}

func (log *logger) Warnc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx)...).Warnf(template, args...)
}

func (log *logger) Errorc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx)...).Errorf(template, args...)
}

func (log *logger) DPanicc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx)...).DPanicf(template, args...)
}

func (log *logger) Panicc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx)...).Panicf(template, args...)
}

func (log *logger) Fatalc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx)...).Fatalf(template, args...)
}
