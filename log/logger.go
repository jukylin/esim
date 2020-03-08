package log

import (
	"context"
	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
	"time"
)

var Log Logger

type logger struct {
	debug bool

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

	var level zap.AtomicLevel
	if logger.debug == true {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()

	zapConfig := zap.Config{
		Level:       level,
		Development: logger.debug,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		DisableCaller:    true,
	}
	zapConfig.EncoderConfig.EncodeTime = logger.standardTimeEncoder

	if logger.debug == true {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig.Encoding = "json"
	}
	zapLogger, _ := zapConfig.Build()
	logger.logger = zapLogger
	logger.sugar = zapLogger.Sugar()

	Log = logger

	return logger
}

func (LoggerOptions) WithDebug(debug bool) Option {
	return func(l *logger) {
		l.debug = debug
	}
}

func (log *logger) Error(msg string) {
	log.logger.Error(msg)
}

func (log *logger) Debugf(template string, args ...interface{}) {
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Debugf(template, args...)
}

func (log *logger) Infof(template string, args ...interface{}) {
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Infof(template, args...)
}

func (log *logger) Warnf(template string, args ...interface{}) {
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Warnf(template, args...)
}

func (log *logger) Errorf(template string, args ...interface{}) {
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Errorf(template, args...)
}

func (log *logger) DPanicf(template string, args ...interface{}) {
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).DPanicf(template, args...)
}

func (log *logger) Panicf(template string, args ...interface{}) {
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Panicf(template, args...)
}

func (log *logger) Fatalf(template string, args ...interface{}) {
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).
		Fatalf(template, args...)
}

func (log *logger) Debugc(ctx context.Context, template string, args ...interface{}) {
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		log.sugar.With("tracer_id", tracerId)
	}
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Debugf(template, args...)
}

func (log *logger) Infoc(ctx context.Context, template string, args ...interface{}) {
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		log.sugar.With("tracer_id", tracerId)
	}
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Infof(template, args...)
}

func (log *logger) Warnc(ctx context.Context, template string, args ...interface{}) {
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		log.sugar.With("tracer_id", tracerId)
	}
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Warnf(template, args...)
}

func (log *logger) Errorc(ctx context.Context, template string, args ...interface{}) {
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		log.sugar.With("tracer_id", tracerId)
	}
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Errorf(template, args...)
}

func (log *logger) DPanicc(ctx context.Context, template string, args ...interface{}) {
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		log.sugar.With("tracer_id", tracerId)
	}
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).DPanicf(template, args...)
}

func (log *logger) Panicc(ctx context.Context, template string, args ...interface{}) {
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		log.sugar.With("tracer_id", tracerId)
	}
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Panicf(template, args...)
}

func (log *logger) Fatalc(ctx context.Context, template string, args ...interface{}) {
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		log.sugar.With("tracer_id", tracerId)
	}
	log.sugar.With("caller", log.getCaller(runtime.Caller(1))).Fatalf(template, args...)
}

func (log *logger) getCaller(pc uintptr, file string, line int, ok bool) string {
	return zapcore.NewEntryCaller(pc, file, line, ok).TrimmedPath()
}

func (log *logger) standardTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

//get  tracer_id from opentracing
func (log *logger) getTracerId(ctx context.Context) string {
	sp := opentracing.SpanFromContext(ctx)
	if sp != nil {
		if jaegerSpanContext, ok := sp.Context().(jaeger.SpanContext); ok {
			return jaegerSpanContext.TraceID().String()
		} else {
			return ""
		}
	} else {
		return ""
	}
}
